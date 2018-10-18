package kubelet

import (
	"bufio"
	"os"
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"io"
	"time"
	"encoding/json"
	"github.com/golang/glog"
	"bytes"
)

const fcDir = "/sys/class/fc_host/"
const iscsiPath = "/etc/iscsi/initiatorname.iscsi"

type Result struct {
	Instance 	string 		`json:"instance,omitempty"`
}

type remoteResp struct {
	Code 		string 		`json:"code,omitempty"`
	Message 	string 		`json:"message,omitempty"`
	Result          Result		`json:"result,omitempty"`
}

func getInstanceIDFromIscsiFile(remoteServerAddr, path string) (string, error) {
	glog.V(1).Info("Try to Get iscsi Infomathion")
	glog.V(1).Infof("read %v get iscsi intanceId", path)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	iscsi := bufio.NewReader(f)
	for {
		line, err := iscsi.ReadString('\n')
		if err != nil || io.EOF == err {
			if io.EOF == err {
				if strings.HasPrefix(line, "InitiatorName") {
					sep := strings.Split(line, "=")
					if len(sep) != 2 {
						continue
					}
					res := getInstanceID(remoteServerAddr, sep[1])
					if res != "" {
						return res, nil
					}
				}
			}
			break
		}
		if strings.HasPrefix(line, "InitiatorName") {
			sep := strings.Split(line, "=")
			if len(sep) != 2 {
				continue
			}
			res := getInstanceID(remoteServerAddr, sep[1])
			if res != "" {
				return res, nil
			}
		}
	}
	return "", fmt.Errorf("Unable Find Iscsi UID")
}

func getInstanceIDFromFcDir(remoteServerAddr, path string) (string, error) {
	glog.V(1).Info("Try to Get FC Infomathion")
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		glog.V(1).Infof("checkout fc dir: %v", path + file.Name())
		glog.V(1).Infof("read %v get fc intanceId", path + "/" + file.Name() + "/port_name")
		port_path := path + "/" + file.Name() + "/port_name"
		f, err := os.Open(port_path)
		if err != nil {
			glog.V(1).Infof("Open %v : %v", port_path, err)
			continue
		}
		defer f.Close()
		fc := bufio.NewReader(f)
		for {
			line, err := fc.ReadString('\n')
			glog.V(1).Infof("File %v Line: %v",  port_path, line)
			if err != nil || io.EOF == err {
				if io.EOF == err {
					if strings.HasPrefix(line, "0x") {
						line = strings.TrimPrefix(line, "0x")
						res := getInstanceID(remoteServerAddr, line)
						if res != "" {
							return res, nil
						}
					}
				}
				break
			}
			if strings.HasPrefix(line, "0x") {
				line = strings.TrimPrefix(line, "0x")
				res := getInstanceID(remoteServerAddr, line)
				if res != "" {
					return res, nil
				}
			}
		}
	}
	return "", fmt.Errorf("Unable Find FibreChannel UID")
}

type HBA struct {
	Hba string	`json:"hba,omitempty"`
}

func getInstanceID(remoteServerAddr, machineid string) string {
	httpClient := &http.Client{}
	httpClient.Timeout = 5 * time.Second
	url := remoteServerAddr + "/v1/instance/info"
	hba := HBA{
		Hba:  machineid,
	}

	resquestContent, _ := json.Marshal(hba)
	requestBody := bytes.NewReader([]byte(resquestContent))
	request, err := http.NewRequest("POST", url, requestBody )
	if err != nil {
		glog.Info("Try To Get Instance Failed:Create HttpGet %v", err)
		glog.Errorf("Try To Get Instance Failed:Create HttpGet %v", err)
		return ""
	}
	glog.V(1).Info("Try to Request Url: ", url)
	//resp, err := httpClient.Get(url)
	response, err := httpClient.Do(request)
	if err != nil {
		glog.Info("Try To Get Instance:%s Volume Type Failed: %v", machineid, err)
		glog.Errorf("Try To Get Instance:%s Volume Type Failed: %v", machineid, err)
		return ""
	}
	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		glog.Infof("Response Body: %v", string(body))
		glog.Infof("Try To Get Instance:%s Volume Type Failed, Remote Server Deny", machineid)
		glog.Infof("host= %v, requestURI= %v", response.Request.Host, response.Request.URL.RequestURI())
		glog.Errorf("Try To Get Instance:%s Volume Type Failed, Remote Server Deny", machineid)
		return ""
	}
	res := remoteResp{}
	//body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &res)
	if err != nil {
		return ""
	}
	return res.Result.Instance
}

func (kl *Kubelet) GetInstanceID() {
	res, err := getInstanceIDFromFcDir(kl.remoteVolumeServerAddr, fcDir)
	if err == nil {
		kl.diskType = "FibreChannel"
		kl.instanceID = res
		glog.V(1).Infof("InstanceID=%v, diskType=%v", kl.instanceID, kl.diskType)
		return
	} else {
		glog.V(1).Infof("Get FC InstanceID meet error: %v", err)
	}

	res, err = getInstanceIDFromIscsiFile(kl.remoteVolumeServerAddr, iscsiPath)
	if err == nil {
		kl.diskType = "Iscsi"
		kl.instanceID = res
		glog.V(1).Infof("InstanceID=%v, diskType=%v", kl.instanceID, kl.diskType)
		return
	} else {
		glog.V(1).Infof("Get Iscsi InstanceID meet error: %v", err)
	}
	glog.V(1).Infof("Unable To Get InstanceID and diskType")
}