package kubelet

import (
	"testing"
	"net/http/httptest"
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
)

func Test1(t *testing.T) {
	remoteUrl := "http://172.25.9.142:56789"
	machineID := "iqn.1993-08.org.debian:01:fcc61ab8ac7c"
	res := getInstanceID(remoteUrl, machineID)
	fmt.Println(res)
}

func prepareFCFile(path, content string) string {
	name, _ := ioutil.TempDir("/tmp",path)
	name1, _ := ioutil.TempDir(name, "test")
	name2, _ := ioutil.TempDir(name, "test")
	ioutil.WriteFile(name1 + "/port_name", []byte(content), 0666 )
	ioutil.WriteFile(name2 + "/port_name", []byte("ssssssss"), 0666 )
	return name
}

func prepareIscsiFile(path, content string) string {
	name, _ := ioutil.TempDir("/tmp", path)
	ioutil.WriteFile(name + "/initiatorname.iscsi", []byte(content), 0666)
	return name + "/initiatorname.iscsi"
}

func TestGetInstanceID(t *testing.T) {
	testHttpServce := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if r.Method != "GET" {
			w.WriteHeader(400)
		}
		fmt.Println(r.RequestURI)
		if r.RequestURI == "/v1/instance/info?hba=iiiiiiiiii" {
			resContent := remoteResp{
				Code: "200",
				Message: "",
				Result: Result {
					Instance: "1111111",
				},
			}
			res, _ := json.Marshal(resContent)
			w.Write(res)
			w.WriteHeader(200)
			return
		}
		if r.RequestURI == "/v1/instance/info?hba=oooooooooo" {
			resContent := remoteResp{
				Code: "200",
				Message: "",
				Result: Result {
					Instance: "2222222",
				},
			}
			res, _ := json.Marshal(resContent)
			w.WriteHeader(200)
			w.Write(res)
			return
		}
		if r.RequestURI == "/v1/instance/info?hba=iqn.1993-08.org.debian:01:ad46d12a23c8" {
			resContent := remoteResp{
				Code: "200",
				Message: "",
				Result: Result {
					Instance: "3333333",
				},
			}
			res, _ := json.Marshal(resContent)
			w.Write(res)
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(400)
	}))
	defer testHttpServce.Close()

	fcPath := "fc"
	fccontent1 := "0xiiiiiiiiii"
	fccontent2 := "0xoooooooooo"
	fcPath1 := prepareFCFile(fcPath, fccontent1)
	fcPath2 := prepareFCFile(fcPath, fccontent2)

	isPath := "iscsi"
	iscsicontent := "InitiatorName=iqn.1993-08.org.debian:01:ad46d12a23c8"
	iscsiPath1 := prepareIscsiFile(isPath, iscsicontent)

	testServerAddr := testHttpServce.URL
	intanceID, _ := getInstanceIDFromFcDir(testServerAddr, fcPath1)
	if intanceID != "1111111" {
		fmt.Println(fcPath1, intanceID)
		t.Fatal("Should Success")
	}

	intanceID, _ = getInstanceIDFromFcDir(testServerAddr, fcPath2)
	if intanceID != "2222222" {
		t.Fatal("Should Success")
	}

	instanceID, _ := getInstanceIDFromIscsiFile(testServerAddr, iscsiPath1)
	if instanceID != "3333333" {
		t.Fatal("Should Success")
	}
}
