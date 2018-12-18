package configmap

import (
	"k8s.io/api/core/v1"
	"net/http"
	"time"
	"io/ioutil"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	api "k8s.io/kubernetes/pkg/apis/core"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/core/validation"
	k8s_api_v1 "k8s.io/kubernetes/pkg/apis/core/v1"
	"github.com/golang/glog"
)

type httpConfigMapManager struct {
	url         string
	client      *http.Client
}

func (s *httpConfigMapManager) GetConfigMap(namespace, name string) (*v1.ConfigMap, error) {
	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("ConfigMapName", name)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v: %v", s.url, resp.Status)
	}
	if len(data) == 0 {
		// Emit an update with an empty ConfigMap to allow HTTPSource to be marked as seen
		return nil, fmt.Errorf("zero-length data received from %v", s.url)
	}

	// First try as it is a single configmap.
	configMap, err := tryDecodeConfigMap(data)
	if err == nil {
		return configMap, nil
	} else {
		return nil, err
	}

	return nil, nil
}

func (s *httpConfigMapManager) RegisterPod(pod *v1.Pod) {
}

func (s *httpConfigMapManager) UnregisterPod(pod *v1.Pod) {
}

func NewHttpConfigMapManager(url string) Manager {
	return &httpConfigMapManager{
		url:	url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func tryDecodeConfigMap(data []byte) (configMap *v1.ConfigMap, err error) {
	// JSON is valid YAML, so this should work for everything.
	json, err := utilyaml.ToJSON(data)
	if err != nil {
		return nil, err
	}
	obj, err := runtime.Decode(legacyscheme.Codecs.UniversalDecoder(), json)
	if err != nil {
		return  configMap, err
	}

	newConfigMap, ok := obj.(*api.ConfigMap)
	// Check whether the object could be converted to single ConfigMap.
	if !ok {
		return  configMap, fmt.Errorf("invalid ConfigMap: %#v", obj)
	}

	if errs := validation.ValidateConfigMap(newConfigMap); len(errs) > 0 {
		return configMap, fmt.Errorf("invalid ConfigMap: %v", errs)
	}
	v1ConfigMap := &v1.ConfigMap{}
	if err := k8s_api_v1.Convert_core_ConfigMap_To_v1_ConfigMap(newConfigMap, v1ConfigMap, nil); err != nil {
		glog.Errorf("ConfigMap %q failed to convert to v1", newConfigMap.Name)
		return nil, err
	}
	return  v1ConfigMap, nil
}
