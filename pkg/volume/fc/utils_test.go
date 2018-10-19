package fc

import (
	"testing"
	"fmt"
	"net/http/httptest"
	"net/http"
	"encoding/json"
	"time"
	"io/ioutil"
	"strconv"
)

func Test1(t *testing.T) {
	a1 := `{
	        "code":"200",
	        "message":"OK",
	        "result":{
	                  "fc":{
	                           "lun":1,
	                           "targetWWNs":["5000d31000d88233","5000d31000d88234","5000d31000d88232","5000d31000d88231"]
	                        },
	                  "name":  "bbb749b7-9062-4f8a-b518-dc837bc15ef7"
	                 }
	       }`

	a2 := `{"code":"200","message":"OK","result":{"attach_status":"attached","create_time":"2018-03-30 16:43:20","owner":"3256","provider_misc":{"fc":{"locker":"751964cd-3656-11e8-96a4-0cc47ab1f7be","lun":2,"targetWWNs":["5000d31000d88231","5000d31000d88232","5000d31000d88234","5000d31000d88233"]}},"size":5,"status":"busy",
"update_time":"2018-04-02 17:30:57","used_size":7,"vol_type":"dellsc","volid":"27d0e31c-16d6-40a0-ba8d-0df1c4f51cab","volume":"volume3","volume_mapping":{"access_mode":"ReadWriteOnce","format":"ext4","instance":"10.6.5.205","path":""}}}`

	result1 := VolumeInfo{}
	result2 := VolumeInfo{}

	json.Unmarshal([]byte(a1), &result1)
	fmt.Println(string(strconv.Itoa(result1.Result.FC.Lun)))

	err := json.Unmarshal([]byte(a2), &result2)
	if err != nil {
		fmt.Println(err.Error())
	}
	//fmt.Println(result2)
	//fmt.Println(result2.Result)
	//fmt.Println(result2.Result.Status, result2.Result.Attach_Status)
	fmt.Println(result2.Result.Status, result2.Result.Provider_Misc.FC)
}

func TestLockToPod(t *testing.T) {
	volumeName1 := "aaaaaa"
	pod1 := "1"
	volumeName2 := "bbbbbb"
	pod2 := "2"
	volumeName3 := "cccccc"
	pod3 := "3"
	testHttpServce := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if r.Method != "POST" {
			w.WriteHeader(400)
		}
		if r.RequestURI == "/v1/volume/lock" {
			requestBody := map[string]string{}
			rawBody, _ := ioutil.ReadAll(r.Body)
			json.Unmarshal(rawBody, &requestBody)
			if requestBody["id"] == volumeName1 && requestBody["locker"] == pod1 {
				w.WriteHeader(200)
				return
			}
			if requestBody["id"] == volumeName2 && requestBody["locker"] == pod2 {
				w.WriteHeader(400)
				return
			}
			if requestBody["id"] == volumeName3 && requestBody["locker"] == pod3 {
				time.Sleep( 10 * time.Second)
				w.WriteHeader(200)
				return
			}
		}
	}))
	defer testHttpServce.Close()
	remoteVolumeServerAddress := testHttpServce.URL
	err := LockToPod(remoteVolumeServerAddress, volumeName1, pod1)
	if err != nil {
		t.Fatal("Should Success")
	}

	err = LockToPod(remoteVolumeServerAddress, volumeName2, pod2)
	if err == nil {
		t.Fatal("Should Fail")
	}

	err = LockToPod(remoteVolumeServerAddress, volumeName3, pod3)
	if err == nil {
		t.Fatal("Should Fail")
	}
}

func TestUnlockFromPod(t *testing.T) {
	volumeName1 := "aaaaaa"
	pod1 := "1"
	volumeName2 := "bbbbbb"
	pod2 := "2"
	volumeName3 := "cccccc"
	pod3 := "3"
	testHttpServce := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if r.Method != "POST" {
			w.WriteHeader(400)
		}
		if r.RequestURI == "/v1/volume/unlock" {
			requestBody := map[string]string{}
			rawBody, _ := ioutil.ReadAll(r.Body)
			json.Unmarshal(rawBody, &requestBody)
			if requestBody["id"] == volumeName1 && requestBody["locker"] == pod1 {
				w.WriteHeader(200)
				return
			}
			if requestBody["id"] == volumeName2 && requestBody["locker"] == pod2 {
				w.WriteHeader(400)
				return
			}
			if requestBody["id"] == volumeName3 && requestBody["locker"] == pod3 {
				time.Sleep( 10 * time.Second)
				w.WriteHeader(200)
				return
			}
		}
	}))
	defer testHttpServce.Close()
	remoteVolumeServerAddress := testHttpServce.URL
	err := UnlockFromPod(remoteVolumeServerAddress, volumeName1, pod1)
	if err != nil {
		t.Fatal("Should Success")
	}

	err = UnlockFromPod(remoteVolumeServerAddress, volumeName2, pod2)
	if err == nil {
		t.Fatal("Should Fail")
	}

	err = UnlockFromPod(remoteVolumeServerAddress, volumeName3, pod3)
	if err == nil {
		t.Fatal("Should Fail")
	}
}

func TestGetVolumeInfo(t *testing.T) {
	volumeName1 := "aaaaaa"
	volumeName2 := "bbbbbb"
	volumeName3 := "cccccc"
	volumeName7 := "gggggg"
	volumeName12 := "mmmmmm"
	testHttpServce := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if r.Method != "POST" {
			w.WriteHeader(400)
		}
		if r.RequestURI == ("/v1/volume/info?volid=" + volumeName1 ){
			data :=  VolumeInfo{}
			data.Result.Attach_Status = "attached"
			data.Result.Status = "busy"
			data.Result.Volume_Mapping.Instance = "node1"
			data.Result.Provider_Misc.FC.Locker = "pod1"
			body,_ := json.Marshal(data)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == ("/v1/volume/info?volid=" + volumeName2 ){
			data :=  VolumeInfo{}
			data.Result.Attach_Status = "attached"
			data.Result.Status = "idle"
			data.Result.Volume_Mapping.Instance = "node2"
			data.Result.Provider_Misc.FC.Locker = "pod2"
			body,_ := json.Marshal(data)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == ("/v1/volume/info?volid=" + volumeName3 ){
			data :=  VolumeInfo{}
			data.Result.Attach_Status = "detached"
			data.Result.Status = "idle"
			data.Result.Volume_Mapping.Instance = "node3"
			data.Result.Provider_Misc.FC.Locker = "pod3"
			body,_ := json.Marshal(data)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == ("/v1/volume/info?volid=" + volumeName7 ){
			data :=  VolumeInfo{}
			data.Result.Attach_Status = "detached"
			data.Result.Status = "idle"
			data.Result.Volume_Mapping.Instance = "node3"
			body,_ := json.Marshal(data)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == ("/v1/volume/info?volid=" + volumeName12 ){
			time.Sleep(10 * time.Second)
			data :=  VolumeInfo{}
			data.Result.Attach_Status = "attached"
			data.Result.Status = "busy"
			data.Result.Volume_Mapping.Instance = "node1"
			data.Result.Provider_Misc.FC.Locker = "pod1"
			body,_ := json.Marshal(data)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
	}))
	defer testHttpServce.Close()
	remoteVolumeServerAddress := testHttpServce.URL

	 podID, nodeID, _, err := GetVolumeStatus(remoteVolumeServerAddress, volumeName1)

	if podID != "pod1" || nodeID != "node1" || err != nil {
		t.Errorf("podID: %v; nodeID: %v; err: %v", podID, nodeID, err)
		t.Fatal("Should Success")
	}

	podID, nodeID, _, err = GetVolumeStatus(remoteVolumeServerAddress, volumeName2)

	if podID != "pod2" || nodeID != "node2" || err != nil {
		t.Errorf("podID: %v; nodeID: %v; err: %v", podID, nodeID, err)
		t.Fatal("Should Success")
	}

	podID, nodeID, _, err = GetVolumeStatus(remoteVolumeServerAddress, volumeName3)

	if podID != "pod3" || nodeID != "node3" || err != nil {
		t.Errorf("podID: %v; nodeID: %v; err: %v", podID, nodeID, err)
		t.Fatal("Should Success")
	}

	podID, nodeID, _,  err = GetVolumeStatus(remoteVolumeServerAddress, volumeName7)

	if  err != nil || podID != "" {
		t.Fatal("podID should empty")
	}
	podID, nodeID, _, err = GetVolumeStatus(remoteVolumeServerAddress, volumeName12)
	if err == nil {
		t.Errorf("podID: %v; nodeID: %v; err: %v", podID, nodeID, err)
		t.Fatal("Should Fail")
	}
}

func TestFCRemoteAttach(t *testing.T) {
	volumeName1 := "aaaaaa"
	volumeName2 := "bbbbbb"
	volumeName3 := "cccccc"
	volumeName4 := "dddddd"
	volumeName5 := "eeeeee"
	testHttpServce := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if r.Method != "POST" {
			w.WriteHeader(400)
		}
		if r.RequestURI == ("/v1/volume/attach/" + volumeName1 ){
			var data AttachInfo
			data.Code = "200"
			data.Message = "OK"
			data.Result.FC.Lun = 1
			data.Result.FC.TargetWWNs = []string{"5000d31000d88233","5000d31000d88232","5000d31000d88234","5000d31000d88231"}
			body,_ := json.Marshal(data)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == ("/v1/volume/attach/" + volumeName4 ){
			var data AttachInfo
			data.Code = "200"
			data.Message = "OK"
			data.Result.FC.Lun = -1
			data.Result.FC.TargetWWNs = []string{"5000d31000d88233","5000d31000d88232","5000d31000d88234","5000d31000d88231"}
			body,_ := json.Marshal(data)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == ("/v1/volume/attach/" + volumeName2 ){
			var data AttachInfo
			data.Code = "433"
			data.Message = "{\"result\":\"StorageCenterError - Exception Message: Error creating a Mapping Profile: Server already mapped to volume\"}"
			body,_ := json.Marshal(data)
			w.WriteHeader(433)
			w.Write(body)
			return
		}
		if r.RequestURI == ("/v1/volume/attach/" + volumeName3 ){
			time.Sleep(10 * time.Second)
			var data AttachInfo
			data.Code = "433"
			data.Message = "{\"result\":\"StorageCenterError - Exception Message: Error creating a Mapping Profile: Server already mapped to volume\"}"
			body,_ := json.Marshal(data)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == ("/v1/volume/attach/" + volumeName5) {
			respData := `{"code":"200","message":"OK","result":{"fc":{"lun":1,"targetWWNs":["5000d31000d8822e","5000d31000d8822d","5000d31000d88230","5000d31000d8822f"]},"name":"355224f7-1074-445d-8ba4-e71988fec8da"}}`
			w.Write([]byte(respData))
			w.WriteHeader(200)
			return
		}
	}))

	defer testHttpServce.Close()
	remoteVolumeServerAddress := testHttpServce.URL
	volName := volumeName1
	instanceID := "aaaaaa"
	_, _, err := FCAttachToServer(remoteVolumeServerAddress, volName, instanceID)
	if err != nil {
		t.Fatal("volume aaaaaa should succeed")
	}

	volName = volumeName2
	instanceID = "bbbbbb"
	_, _, err = FCAttachToServer(remoteVolumeServerAddress, volName, instanceID)
	if err == nil {
		t.Fatal("volume bbbbbb should fail")
	}

	volName = volumeName3
	instanceID = "cccccc"
	_, _, err = FCAttachToServer(remoteVolumeServerAddress, volName, instanceID)
	if err == nil {
		t.Fatal("volume ccccc should fail")
	}

	volName = volumeName4
	instanceID = "dddddd"
	_, _, err = FCAttachToServer(remoteVolumeServerAddress, volName, instanceID)
	if err == nil {
		t.Fatal("volume ddddddd should fail")
	}

	volName1 := volumeName5
	instanceID1 := "eeeeee"
	_, _, err1 := FCAttachToServer(remoteVolumeServerAddress, volName1, instanceID1)
	if err1 != nil {
		fmt.Println(err1.Error())
		t.Fatal("volume ddddddd should succeed")
	}
}

func TestRemoteDetach(t *testing.T) {
	volumeName1 := "aaaaaa"
	volumeName2 := "bbbbbb"
	volumeName3 := "cccccc"
	testHttpServce := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if r.Method != "POST" {
			w.WriteHeader(400)
		}
		if r.RequestURI == ("/v1/volume/detach/" + volumeName1 ){
			var res VolumeInfo
			res.Code = "200"
			res.Message = "OK"
			res.Result.FC.Lun = 1
			res.Result.FC.TargetWWNs = []string{"5000d31000d88233","5000d31000d88232","5000d31000d88234","5000d31000d88231"}
			res.Result.Name = "aaaaaa"
			body,_ := json.Marshal(res)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == ("/v1/volume/detach/" + volumeName2 ){
			var res VolumeInfo
			res.Code = "433"
			res.Message = "对应卷已解除映射"
			body,_ := json.Marshal(res)
			w.WriteHeader(433)
			w.Write(body)
			return
		}
		if r.RequestURI == ("/v1/volume/detach/" + volumeName3 ){
			time.Sleep(10*time.Second)
			var res VolumeInfo
			res.Code = "433"
			res.Message = "对应卷已解除映射"
			body,_ := json.Marshal(res)
			w.Write(body)
			w.WriteHeader(200)
			return
		}
	}))
	defer testHttpServce.Close()

	remoteVolumeServerAddress := testHttpServce.URL
	volName := volumeName1
	instanceID := "aaaaaa"
	err := DetachFromServer(remoteVolumeServerAddress, volName, instanceID)
	if err != nil {
		t.Fatal("volume aaaaaa should succeed")
	}

	volName = volumeName2
	instanceID = "bbbbbb"
	err = DetachFromServer(remoteVolumeServerAddress, volName, instanceID)
	if err == nil {
		t.Fatal("volume bbbbbb should fail")
	}

	volName = volumeName3
	instanceID = "cccccc"
	err = DetachFromServer(remoteVolumeServerAddress, volName, instanceID)
	if err == nil {
		t.Fatal("volume ccccc should fail")
	}
}
