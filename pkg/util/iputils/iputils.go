package iputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
)

const (
	// Annotations for pods
	MaskAnnotationKey   = "mask"
	RoutesAnnotationKey = "routes"
	IPAnnotationKey     = "ips"
	NetworkKey          = "network"
	ChangeGateway       = "changeGateWay"
	// Label for network groups
	GroupedLabel = "networkgroup"
	ClusterName  = "clustername"
	Zone         = "zone"
)

// IpUtils is used for applying/releasing IP from net-manager.
type ipUtils struct {
	baseIPURL    string
	getIPURL     string
	releaseIPURL string
	ipLocation   string
}

type IpUtils interface {
	// Get IP, Mask and other information, return Response directly
	GetIPMaskForPod(reqBytes []byte) (ipResp IpResp, err error)
	// Get and Add IP/Mask/Route to the pod's annotation
	AddIPMaskIfPodLabeled(pod *v1.Pod, namespace string) (ip string, mask int, err error)
	// Get Group and IP from pod annotation
	GetGroupedIpFromPod(pod *v1.Pod) (group, ip string)
	// Release IP from pod's annotation
	ReleaseIPForPod(pod *v1.Pod) error
	// Release IP with group
	ReleaseGroupedIP(namespace, group, ip string) error
}

type IpResp struct {
	Result  IpResult `json:"result,omitempty"`
	Code    int      `json:"code,omitempty"`
	Message string   `json:"message,omitempty"`
}

type IpRequire struct {
	Group     string `json:"group,omitempty"`
	UserId    int    `json:"userId,omitempty"`
	NetType   int    `json:"type,omitempty"`
	Location  string `json:"location,omitempty"`
	Zone      string `json:"zone,omitempty"`
	IsPrivate int    `json:"isPrivate,omitempty"`
}

type IpRelease struct {
	IP     string `json:"ip,omitempty"`
	Group  string `json:"group,omitempty"`
	UserId int    `json:"userId,omitempty"`
}

type IpResult struct {
	Routes   []string `json:"routes,omitempty"`
	IP       string   `json:"ip,omitempty"`
	Mask     int      `json:"mask,omitempty"`
	Occupied int      `json:"occupied,omitempty"`
	Location string   `json:"location,omitempty"`
}

type IpReleaseResp struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

func NewIPUtils(url, location string) IpUtils {
	return &ipUtils{
		baseIPURL:    url,
		getIPURL:     url + "/api/net/ip/occupy",
		releaseIPURL: url + "/api/net/ip/release",
		ipLocation:   location,
	}
}

func (iu *ipUtils) GetIPMaskForPod(reqBytes []byte) (ipResp IpResp, err error) {
	resp, err := http.Post(iu.getIPURL, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &ipResp)
	if err != nil {
		return
	}
	if ipResp.Code != 200 {
		err = fmt.Errorf("%v", ipResp.Message)
	}
	return
}

func (iu *ipUtils) sendReleaseIpReq(reqBytes []byte) (code int, err error) {
	resp, err := http.Post(iu.releaseIPURL, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var ipResp IpReleaseResp
	err = json.Unmarshal(body, &ipResp)
	if err != nil {
		return
	}
	code = ipResp.Code
	if code != 200 {
		err = fmt.Errorf("%v", ipResp.Message)
	}
	return
}

func (iu *ipUtils) ReleaseGroupedIP(namespace, group, ip string) error {
	glog.V(6).Infof("ReleaseIP %v ip %v for group: %v", namespace, ip, group)
	userIds := strings.Split(namespace, "-")
	lenIds := len(userIds)
	if lenIds <= 1 {
		err := fmt.Errorf("Wrong Namespace format %v !", namespace)
		return err
	}
	userId := userIds[lenIds-1]
	uid, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	req := IpRelease{
		IP:     ip,
		UserId: uid,
	}
	if group != "" {
		req.Group = group
	}
	glog.V(6).Infof("ReleaseIPReq: %v", req)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	// Retry 3 times in case of network error.
	for i := 0; i < 3; i++ {
		code, err := iu.sendReleaseIpReq(reqBytes)
		if err == nil {
			return nil
		}
		glog.Errorf("Failed to release ip %v: %v", ip, err)
		if code != 0 {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	glog.V(6).Infof("IP %v successfully released.", ip)
	return err
}

func (iu *ipUtils) AddIPMaskIfPodLabeled(pod *v1.Pod, namespace string) (ip string, mask int, err error) {
	// TODO: too many ifs
	// No needs to add ips if no label or "ips" has already been added.
	if pod.Annotations[IPAnnotationKey] != "" || pod.Labels[NetworkKey] == "" {
		return
	}
	nets := strings.Split(pod.Labels[NetworkKey], "-")
	if len(nets) != 2 {
		err = fmt.Errorf("Illegal network label: %v", pod.Labels[NetworkKey])
		return
	}
	userIds := strings.Split(namespace, "-")
	lenIds := len(userIds)
	if lenIds <= 1 {
		err = fmt.Errorf("Wrong Namespace format %v !", pod.Namespace)
		return
	}
	userId := userIds[lenIds-1]
	uid, err := strconv.Atoi(userId)
	if err != nil {
		return
	}
	groupLabel := pod.Labels[GroupedLabel]

	location := iu.ipLocation

	if pod.Labels["location"] != "" {
		location = pod.Labels["location"]
	}
	req := IpRequire{
		UserId: uid,
	}
	if groupLabel != "" {
		req.Group = groupLabel
	} else {
		// If groupLabel is added, we do NOT need location.
		req.Location = location
	}

	switch nets[1] {
	case "InnerNet":
		req.NetType = 1 // Production environment: 1, Debug env: 2
		if pod.Labels["isPrivate"] == "1" {
			req.IsPrivate = 1
			req.Zone = pod.Labels["zone"]
		}
	case "OuterNet":
		req.NetType = 3
	case "PrivateNet":
		req.NetType = 4
	}
	glog.V(6).Infof("%v Get IP Req: %v", pod.GenerateName, req)

	reqBytes, _ := json.Marshal(req)

	var ipResp IpResp
	// Retry 3 times in case of network error.
	// TODO: add UUID to ensure idempotence.
	for i := 0; i < 3; i++ {
		ipResp, err = iu.GetIPMaskForPod(reqBytes)
		// code = 0 means connection error
		if err != nil {
			glog.Errorf("Failed to GetIPMaskForPod %v: %v.  Req: %v", pod.Name, err, req)
			// If code is 0, network fails so retry.
			if ipResp.Code != 0 {
				return
			}
		} else {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	ip = ipResp.Result.IP
	mask = ipResp.Result.Mask
	location = ipResp.Result.Location
	if ip == "" {
		glog.Errorf("ipResp: %v", ipResp)
		return
	}
	// Pass from labels to annotaions so Kubelet can handle
	if pod.Labels[ChangeGateway] == "true" {
		pod.Annotations[ChangeGateway] = "true"
	}
	pod.Annotations[IPAnnotationKey] = fmt.Sprintf("%s-%s", nets[0], ip)
	pod.Annotations[MaskAnnotationKey] = fmt.Sprintf("%s-%d", nets[0], mask)
	pod.Annotations[RoutesAnnotationKey] = strings.Join(ipResp.Result.Routes, ";")
	if location != "" {
		pod.Annotations["location"] = location
		if pod.Spec.NodeSelector == nil {
			pod.Spec.NodeSelector = make(map[string]string)
		}
		pod.Spec.NodeSelector["location"] = location
	}
	if pod.Labels["zone"] != "" {
		pod.Spec.NodeSelector["zone"] = pod.Labels["zone"]
	}
	glog.V(6).Infof("Get IP: %v, Mask: %v, ForPod: %v ", ip, mask, pod.ObjectMeta)
	return
}

func (iu *ipUtils) GetGroupedIpFromPod(pod *v1.Pod) (group, ip string) {
	group = pod.Labels[GroupedLabel]
	if ips := pod.Annotations[IPAnnotationKey]; ips != "" {
		ipArr := strings.Split(ips, "-")
		if len(ipArr) == 2 {
			ip = ipArr[1]
		}
	}
	return
}

func (iu *ipUtils) ReleaseIPForPod(pod *v1.Pod) error {
	if group, ip := iu.GetGroupedIpFromPod(pod); ip != "" && ip != "none" && ip != "empty" {
		glog.Infof("Releasing IP %v for pod %v", ip, pod.ObjectMeta)
		err := iu.ReleaseGroupedIP(pod.Namespace, group, ip)
		return err
	}
	return nil
}
