/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logdir

import (
	"fmt"
	"io"
	"strings"

	"github.com/golang/glog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/storage/names"
	api "k8s.io/kubernetes/pkg/apis/core"
)

const (
	logDirAnnotation = "logDir"
	pathPrefix       = "/var/log/containers"
	PluginName       = "LogDir"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		logDirPlug := NewLogDir()
		return logDirPlug, nil
	})
}

// logDir is an implementation of admission.Interface.
// It looks at all new pods and add log volumes if they are annotated.
type logDir struct {
	*admission.Handler
}

func addLogVolumesIfNeeded(pod *api.Pod, namespace string) bool {
	flag := false
	if len(pod.Annotations) <= 1 {
		return flag
	}
	if pod.Name == "" {
		pod.Name = names.SimpleNameGenerator.GenerateName(pod.GenerateName)
	}
	for key, mountPath := range pod.Annotations {
		// Annotations like:
		// logDir.container1.javalog: /var/log/java
		// logDir.container2.systemd: /mnt/log/systemd
		if strings.HasPrefix(key, logDirAnnotation) {
			strs := strings.Split(key, ".")
			if len(strs) == 3 {
				flag = true
				containerName := strs[1]
				fileTag := strs[2]
				for idx, con := range pod.Spec.Containers {
					if con.Name == containerName {
						hostPath := fmt.Sprintf("%s/%s_%s_%s-%s", pathPrefix, pod.Name, namespace, containerName, fileTag)
						glog.V(6).Infof("Add logVolume's hostpath %v for pod: %v, container %v", hostPath, pod.Name, containerName)
						volMount := api.VolumeMount{
							Name:      containerName + "-logvol-" + fileTag,
							MountPath: mountPath,
						}
						con.VolumeMounts = append(con.VolumeMounts, volMount)
						pod.Spec.Containers[idx] = con
						vol := api.Volume{
							Name: containerName + "-logvol-" + fileTag,
							VolumeSource: api.VolumeSource{
								HostPath: &api.HostPathVolumeSource{
									Path: hostPath,
								},
							},
						}
						pod.Spec.Volumes = append(pod.Spec.Volumes, vol)
					}
				}
			}
		}
	}
	return flag
}

func (a *logDir) Admit(attributes admission.Attributes) (err error) {
	// Ignore all calls to subresources or resources other than pods.
	if len(attributes.GetSubresource()) != 0 || attributes.GetResource().GroupResource() != api.Resource("pods") {
		return nil
	}
	pod, ok := attributes.GetObject().(*api.Pod)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Pod but was unable to be converted")
	}
	addLogVolumesIfNeeded(pod, attributes.GetNamespace())
	return nil
}

// NewLogDir creates a new admission control handler for adding log volumes
func NewLogDir() admission.Interface {
	return &logDir{
		Handler: admission.NewHandler(admission.Create),
	}
}
