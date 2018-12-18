package secret

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

type httpSecretManager struct {
	url string
	client      *http.Client
}

func NewHttpSecretManager(url string) Manager {
	return &httpSecretManager{
		url:	url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *httpSecretManager) GetSecret(namespace, name string) (*v1.Secret, error) {
	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("SecretName", name)
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
		// Emit an update with an empty Secret to allow HTTPSource to be marked as seen
		return nil, fmt.Errorf("zero-length data received from %v", s.url)
	}

	// First try as it is a single Secret.
	secret, err := tryDecodeSecret(data)
	if err == nil {
		return secret, nil
	} else {
		return nil, err
	}

	return nil, nil
}

func (s *httpSecretManager) RegisterPod(pod *v1.Pod) {
}

func (s *httpSecretManager) UnregisterPod(pod *v1.Pod) {
}

func tryDecodeSecret(data []byte) (secret *v1.Secret, err error) {
	// JSON is valid YAML, so this should work for everything.
	json, err := utilyaml.ToJSON(data)
	if err != nil {
		return nil, err
	}
	obj, err := runtime.Decode(legacyscheme.Codecs.UniversalDecoder(), json)
	if err != nil {
		return  nil, err
	}

	newSecret, ok := obj.(*api.Secret)
	// Check whether the object could be converted to single Secret.
	if !ok {
		return  nil, fmt.Errorf("invalid Secret: %#v", obj)
	}

	if errs := validation.ValidateSecret(newSecret); len(errs) > 0 {
		return secret, fmt.Errorf("invalid secret: %v", errs)
	}
	v1Secret := &v1.Secret{}
	if err := k8s_api_v1.Convert_core_Secret_To_v1_Secret(newSecret, v1Secret, nil); err != nil {
		glog.Errorf("Secret %q failed to convert to v1", newSecret.Name)
		return nil, err
	}
	return  v1Secret, nil
}


