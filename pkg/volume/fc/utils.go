package fc

import (
	"fmt"
	"net/http"
	"time"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"strconv"
	"github.com/golang/glog"
)

type Volume_Mapping struct {
	Format		string 		 `json:"format,omitempty"`
	Access_Mode     string 		 `json:"access_mode,omitempty"`
	Path            string		 `json:"path,omitempty"`
	Instance        string           `json:"instance,omitempty"`
}

type FC struct {
	Lun             int             `json:"lun,omitempty"`
	TargetWWNs      []string        `json:"targetWWNs,omitempty"`
}

type ISCSI struct {
	IQN		string 		`json:"iqn,omitempty"`
	Lun 		int		`json:"lun,omitempty"`
	TargetPortal    string		`json:"targetPortal,omitempty"`
}

type IscsiMisc struct {
	Locker string 			`json:"locker,omitempty"`
	Target []ISCSI			`json:"target,omitempty"`
}

type FCMisc struct {
	Locker string			`json:"locker,omitempty"`
	Lun    int			`json:"lun,omitempty"`
	TargetWNNs []string     	`json:"targetWWNs,omitempty"`
}

type ProviderMisc struct {
	FC    FCMisc                    `json:"fc,omitempty"`
	Iscsi IscsiMisc			`json:"iscsi,omitempty"`
}

type VolumeDetails struct {
	Volid           string 	        `json:"volid,omitempty"`
	Volume	        string		`json:"volume,omitempty"`
	Owner           string          `json:"owner,omitempty"`
	Size            int          	`json:"size,omitempty"`
	Used_Size       int          	`json:"used_size,omitempty"`
	Status          string          `json:"status,omitempty"`
	Attach_Status   string          `json:"attach_status,omitempty"`
	Vol_Type        string          `json:"vol_type,omitempty"`
	Provider_Misc   ProviderMisc    `json:"provider_misc,omitempty"`
	Name 		string 		`json:"name,omitempty"`
	FC 		FC              `json:"fc,omitempty"`
	ISCSI           []ISCSI         `json:"iscsi,omitempty"`
	Volume_Mapping  Volume_Mapping  `json:"volume_mapping,omitempty"`
}

type VolumeInfo struct {
	Code 		string 		`json:"code,omitempty"`
	Message 	string		`json:"message,omitempty"`
	Name		string 		`json:"name,omitempty"`
	Result          VolumeDetails   `json:"result,omitempty"`
}


type AttachResult struct {
	FC 		FC 		`json:"fc,omitempty"`
}

type AttachInfo struct {
	Code 		string 		`json:"code,omitempty"`
	Message 	string		`json:"message,omitempty"`
	Name		string 		`json:"name,omitempty"`
	Result		AttachResult	`json:"result,omitempty"`
}

func GetVolumeStatus(remoteServerAddress string, volumeID string) (podID string, nodeID string, provider_misc ProviderMisc, err error) {
	glog.V(1).Info("RemoteAttach/RemoteDetach Try To Get Volume Infomation")
	httpClient := http.Client{}
	httpClient.Timeout = 30 * time.Second
	requestUrl := remoteServerAddress + "/v1/volume/info?volid=" + volumeID
	response, err := httpClient.Get(requestUrl)
	if err != nil {
		err = fmt.Errorf("Unable To Get Volume: %v Infomation, Error is : %v", volumeID, err)
		return
	}

	var data VolumeInfo
	body, _ := ioutil.ReadAll(response.Body)
	if err = json.Unmarshal(body, &data); err != nil {
		err = fmt.Errorf("Invalid Volume Server Response, Can't Marshal it. Error is: %v", err)
		return
	}

	//if data.Result.Attach_Status == "attached" {
	//	attachedToNode = true
	//} else if data.Result.Attach_Status == "detached" {
	//	attachedToNode = false
	//} else {
	//	err = fmt.Errorf("Invalid Volume Server Response, attach_status is Invalid, attach_status: %v", data.Result.Attach_Status)
	//	return
	//}
	//
	//if data.Result.Status == "idle" {
	//	lockedByPod = false
	//} else if data.Result.Status == "busy" {
	//	lockedByPod = true
	//} else {
	//	err = fmt.Errorf("Invalid Volume Server Response, status is Invalid, status: %v", data.Result.Status)
	//	return
	//}
	podID = data.Result.Provider_Misc.FC.Locker
	nodeID = data.Result.Volume_Mapping.Instance
	provider_misc = data.Result.Provider_Misc
	return
}

func FCAttachToServer(remoteVolumeServerAddress, instanceID, volumeID string) (lun int, targetWWns []string, err error) {
	glog.V(1).Info("FibreChannel RemoteAttach Begin")
	glog.V(1).Info("RemoteAttach FibreChannel: " + instanceID + ";" + remoteVolumeServerAddress + ";" + volumeID)
	httpClient := http.Client{}
	httpClient.Timeout = 30 * time.Second
	requestUrl := remoteVolumeServerAddress + "/v1/volume/attach/" + volumeID
	glog.V(1).Info("RemoteAttach FibreChannel URL : " + requestUrl )
	requestData := "{\"instance\":\"" + instanceID  + "\",\"protocol\":\"FibreChannel\"}"
	glog.V(1).Info("RemoteAttach FibreChannel Body: " + string(requestData))

	response, err := httpClient.Post(requestUrl,"application/json", bytes.NewReader([]byte(requestData)))
	if err != nil {
		return
	}
	glog.V(1).Info("RemoteAttach FibreChannel Response : " ,  response.StatusCode, response.Header )
	var data AttachInfo
	//var data map[string]interface{}
	body, _ := ioutil.ReadAll(response.Body)
	glog.V(1).Infof("RemoteAttach FibreChannel RemoteResponse: %v", string(body))
	if err = json.Unmarshal(body, &data); err != nil {
		glog.V(1).Infof("RemoteAttach FibreChannel error: %v", err)
		return
	}
	glog.V(1).Info("RemoteAttach FibreChannel ReturnBody: " , data)
	if response.StatusCode != 200 && response.StatusCode != 704 {
		err = fmt.Errorf(data.Message)
		return
	}

	if response.StatusCode == 704 {
		 _, _, misc ,err := GetVolumeStatus(remoteVolumeServerAddress, volumeID)
		if err != nil {
			glog.V(1).Infof("RemoteAttach FibreChannel: Volume has Already Attached to this Node, but get volume status failed: %v", err)
			err = fmt.Errorf("RemoteAttach FibreChannel: Volume has Already Attached to this Node, but get volume status failed: %v", err)
			return lun, targetWWns, err
		}
		lun = misc.FC.Lun
		targetWWns = misc.FC.TargetWNNs
		return lun, targetWWns, err
	}

	if data.Result.FC.TargetWWNs == nil {
		err = fmt.Errorf("RemoteAttach FibreChannel ReturnBody Invalid: targetWWNs is nil!")
		return
	}

	lun = data.Result.FC.Lun
	if  lun < 0 {
		err = fmt.Errorf("RemoteAttach FibreChannel ReturnBody Invalid: lun < 0 ")
		return
	}
	targetWWns = data.Result.FC.TargetWWNs
	if len(targetWWns) == 0 {
		err = fmt.Errorf("RemoteAttach FibreChannel ReturnBody Invalid: len(targetWWns) == 0 ")
		return
	}
	glog.V(1).Info("RemoteAttach FibreChannel Success")
	glog.V(1).Infof("RemoteAttach FibreChannel Success: Lun=%v TargetWWNs=%v", lun, targetWWns)
	return lun, targetWWns , nil
}

func LockToPod(remoteVolumeServerAddress, volumeID, podID string) error {
	glog.V(1).Info("FibreChannel LockToPod Begin")
	glog.V(1).Info("FibreChannel LockToPod Infomation: PodID=%v VolumeServer=%v VolumeID=%v", podID, remoteVolumeServerAddress, volumeID)
	httpClient := http.Client{}
	httpClient.Timeout = 30 * time.Second
	requestUrl := remoteVolumeServerAddress + "/v1/volume/lock"
	requestData := "{\"id\":\"" + volumeID  + "\",\"locker\":\"" + podID + "\"}"
	glog.V(1).Info("FibreChannel LockToPod RequestInfo: %v", requestData)
	response, err := httpClient.Post(requestUrl,"application/json", bytes.NewReader([]byte(requestData)))
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		glog.V(1).Info("FibreChannel LockToPod Failed: %v", requestData)
		return fmt.Errorf("VolumeID=%v PodID=%v LockToPod Failed", volumeID, podID)
	}
	glog.V(1).Info("FibreChannel LockToPod Success, VolumeID=%v", requestData)
	return nil
}


func DetachFromServer(remoteVolumeServerAddress, instanceID, volumeID string) error {
	glog.V(1).Info("FibreChannel RemoteDetach Begin")
	glog.V(1).Info("FibreChannel RemoteDetach FibreChannel: " + instanceID + ";" + remoteVolumeServerAddress + ";" + volumeID)
	httpClient := http.Client{}
	httpClient.Timeout = 30 * time.Second
	requestUrl := remoteVolumeServerAddress + "/v1/volume/detach/" + volumeID
	response, err := httpClient.Post(requestUrl,"application/json", bytes.NewReader([]byte("")))
	if err != nil {
		return err
	}
	var data VolumeInfo
	body, _ := ioutil.ReadAll(response.Body)
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	glog.V(1).Info("Dell RemoteDetach ReturnBody: " , data)
	if response.StatusCode != 200 {
		return fmt.Errorf(data.Message)
	}
	glog.V(1).Info("Dell RemoteDetach Success")
	return nil
}

func UnlockFromPod(remoteVolumeServerAddress, volumeID, podID string) error {
	glog.V(1).Info("FibreChannel UnlockFromPod Begin: VolumeID=%v PodID=%v", volumeID, podID)
	httpClient := http.Client{}
	httpClient.Timeout = 30 * time.Second
	requestUrl := remoteVolumeServerAddress + "/v1/volume/unlock"
	requestData := "{\"id\":\"" + volumeID  + "\",\"locker\":\"" + podID + "\"}"
	glog.V(1).Info("FibreChannel UnlockFromPod RequestContent: %v", requestData)
	response, err := httpClient.Post(requestUrl,"application/json", bytes.NewReader([]byte(requestData)))
	if err != nil {
		glog.V(1).Info("FibreChannel UnlockFromPod Failed: %v", err)
		return err
	}

	if response.StatusCode != 200 {
		glog.V(1).Info("FibreChannel UnlockFromPod Failed, RemoteServer Refuse")
		return fmt.Errorf("FibreChannel Unlock volume %v from pod: %v failed", volumeID, podID)
	}
	glog.V(1).Info("FibreChannel UnlockFromPod Success")
	return nil
}

//two Phase: 1. Unmap To Server; 2. Unlock from Pod
func Unlock(remoteVolumeServerAddress, volumeID, podID, instanceID string) error {
	glog.V(1).Info("FibreChannel Unlock Begin")
	glog.V(1).Info("FibreChannel Unlock, Try to UnlockFromPod Begin")
	err := UnlockFromPod(remoteVolumeServerAddress, volumeID, podID)
	if err != nil {
		glog.V(1).Info("FibreChannel Unlock, UnlockFromPod Failed: %v", err)
		return err
	}

	glog.V(1).Info("FibreChannel Unlock, Try to RemoteDetach from Server")
	err = DetachFromServer(remoteVolumeServerAddress, instanceID, volumeID)
	if err != nil {
		glog.V(1).Info("FibreChannel Unlock, RemoteDetach Failed: %v", err)
		return err
	}
	return nil
}


func LockFibreChannel(remoteVolumeServerAddress, volumeID, instanceID, podID string) (lun string, wwns []string, matched bool, err error) {
	glog.V(1).Info("FibreChannel Lock Volume Begin")
	glog.V(1).Infof("FibreChannel Lock Volume Begin: InstanceID=%v VolumeID=%v", instanceID, volumeID)
	oldPodID, oldNodeID, provide_misc,  err := GetVolumeStatus(remoteVolumeServerAddress, volumeID)
	if err != nil {
		glog.V(1).Info("FibreChannel Lock Volume Failed,Cause We Can't Get Information")
		matched = false
		err = fmt.Errorf("FibreChannel Get Volume Info Error: %v",err)
		return
	}

	if oldNodeID == "" {
		if oldPodID == "" {
			glog.V(1).Info("FibreChannel Try To RemoteAttach")
			targetlun, targetWWns, err := FCAttachToServer(remoteVolumeServerAddress,instanceID, volumeID)
			if err != nil {
				glog.V(1).Info("FibreChannel RemoteAttach Failed %v", err)
				matched = false
				return lun, wwns, matched, err
			}

			lun = strconv.Itoa(targetlun)
			wwns = targetWWns
			glog.V(1).Info("FibreChannel Try To LockToPod")
			err = LockToPod(remoteVolumeServerAddress, volumeID, podID)
			if err != nil {
				glog.V(1).Info("FibreChannel LockToPod Failed")
				matched = false
				return lun, wwns, matched, err
			}
			matched = true
			err = nil
			return lun, wwns, matched, err
		} else {
			matched = false
			err = fmt.Errorf("Remote Server Locker Error: This Volume Belong to A Pod but not belong to a Node")
			return lun, wwns, matched, err
		}
	} else if oldNodeID != instanceID {
		if oldPodID == "" {
			glog.V(1).Info("FibreChannel Try To RemoteDetach From Node: " + oldNodeID)
			err := DetachFromServer(remoteVolumeServerAddress, oldNodeID, volumeID)
			if err != nil {
				matched = false
				err = fmt.Errorf("FibreChannel Volume belong to Node: %v, but not belong to Any Pod,We try release it then MapTo %v,Meet Error: %v", oldNodeID, instanceID, err)
				return lun, wwns, matched, err
			}
			glog.V(1).Info("FibreChannel Try To RemoteAttach")
			targetlun, targetWWns, err := FCAttachToServer(remoteVolumeServerAddress,instanceID, volumeID)
			if err != nil {
				glog.V(1).Info("FibreChannel RemoteAttach Failed %v", err)
				matched = false
				return lun, wwns, matched, err
			}

			lun = strconv.Itoa(targetlun)
			wwns = targetWWns
			glog.V(1).Info("FibreChannel Try To LockToPod")
			err = LockToPod(remoteVolumeServerAddress, volumeID, podID)
			if err != nil {
				glog.V(1).Info("FibreChannel LockToPod Failed")
				matched = false
				return lun, wwns, matched, err
			}
			matched = true
			err = nil
			return lun, wwns, matched, err
		} else if podID != oldNodeID {
			matched = false
			err = fmt.Errorf("FibreChannel Already Locked by another Pod: %v", oldNodeID)
			return lun, wwns, matched, err
		} else {
			// pod transfer to another node
			err := Unlock(remoteVolumeServerAddress, volumeID, podID , oldNodeID)
			if err != nil {
				matched = true
				err = fmt.Errorf("FibreChannel Volume belong to Another Node, but belong to this Pod, Try Unlock Volume ,But Meet Error: %v", err)
				return lun, wwns, matched, err
			}

			glog.V(1).Info("FibreChannel Try To RemoteAttach")
			targetLun, targetWWns, err := FCAttachToServer(remoteVolumeServerAddress, instanceID, volumeID)
			if err != nil {
				glog.V(1).Info("FibreChannel RemoteAttach Failed %v", err)
				matched = false
				return lun, wwns, matched, err
			}

			lun = strconv.Itoa(targetLun)
			wwns = targetWWns
			glog.V(1).Info("FibreChannel Try To LockToPod")
			err = LockToPod(remoteVolumeServerAddress, volumeID, podID)
			if err != nil {
				glog.V(1).Info("FibreChannel LockToPod Failed")
				matched = false
				return lun, wwns, matched, err
			}
			matched = true
			return lun, wwns, matched, err
		}
	} else {
		if podID == "" {
			glog.V(1).Info("FibreChannel Try To RemoteDetach From Node: " + instanceID)
			err := DetachFromServer(remoteVolumeServerAddress, instanceID, volumeID)
			if err != nil {
				matched, err = false, fmt.Errorf("FibreChannel Volume belong to Node: %v, but not belong to Any Pod,We try release it then MapTo %v,Meet Error: %v", instanceID, podID, err)
				return lun, wwns, matched, err
			}
			glog.V(1).Info("FibreChannel Try To RemoteAttach")
			targetLun, targetWWns, err := FCAttachToServer(remoteVolumeServerAddress, instanceID, volumeID)
			if err != nil {
				glog.V(1).Info("FibreChannel RemoteAttach Failed %v", err)
				matched = false
				return lun, wwns, matched, err
			}

			lun = strconv.Itoa(targetLun)
			wwns = targetWWns
			glog.V(1).Info("FibreChannel Try To LockToPod")
			err = LockToPod(remoteVolumeServerAddress, volumeID, podID)
			if err != nil {
				glog.V(1).Info("FibreChannel LockToPod Failed")
				matched = false
				return lun, wwns, matched, err
			}
			matched = true
			err = nil
			return lun, wwns, matched, err
		} else if podID == oldPodID{
			if provide_misc.FC.Locker != "" {
				lun = strconv.Itoa(provide_misc.FC.Lun)
				wwns = provide_misc.FC.TargetWNNs
				matched = true
				err = nil
				return lun, wwns, matched, err

			} else {
				matched, err = false, fmt.Errorf("FibreChannel RemoteAttach Failed: Invalid Provider_Misc, Can't get wwns,lun")
				return lun, wwns, matched, err
			}
		} else {
			matched, err = false, fmt.Errorf("This Volume is held by another Pod on this Node")
			return lun, wwns, matched, err
		}
	}
	matched = true
	err = nil
	return lun, wwns, matched, err
}

func Lock(remoteVolumeServerAddress, volumeID, instanceID, podID string) (string, []string, bool, error) {
	return LockFibreChannel(remoteVolumeServerAddress, volumeID, instanceID, podID)
}
