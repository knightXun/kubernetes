package logdir

import (
	"k8s.io/apimachinery/pkg/runtime"
	"testing"

	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// TestAdmission verifies all create requests for pods result in every container's image pull policy
// set to Always
func TestAdmission(t *testing.T) {
	namespace := "test-namespace"
	podName := "podName"
	containerName1 := "con1"
	containerName2 := "con2"
	handler := &logDir{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Annotations: map[string]string{
				"logDir.con1.0": "/var/log/java",
				"logDir.con2.0": "/home/journal",
			},
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: containerName1, Image: "image"},
				{Name: containerName2, Image: "image"},
			},
		},
	}
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	if len(pod.Spec.Volumes) != 2 {
		t.Errorf("Expected volume size 2, got %v", len(pod.Spec.Volumes))
	}
	expectedPath := fmt.Sprintf("/var/log/containers/%s_%s_%s-0", podName, namespace, containerName1)
	if pod.Spec.Volumes[0].HostPath.Path != expectedPath {
		t.Errorf("Expected hostpath %s, got %s", expectedPath, pod.Spec.Volumes[0].HostPath.Path)
	}
	expectedPath = fmt.Sprintf("/var/log/containers/%s_%s_%s-0", podName, namespace, containerName2)
	if pod.Spec.Volumes[1].HostPath.Path != expectedPath {
		t.Errorf("Expected hostpath %s, got %s", expectedPath, pod.Spec.Volumes[1].HostPath.Path)
	}

	if len(pod.Spec.Containers[0].VolumeMounts) != 1 || len(pod.Spec.Containers[1].VolumeMounts) != 1 {
		t.Errorf("Expected VolumeMounts size (1,1), got %v, %v", len(pod.Spec.Containers[0].VolumeMounts), len(pod.Spec.Containers[1].VolumeMounts))
	}
	if pod.Spec.Containers[0].VolumeMounts[0].MountPath != "/var/log/java" || pod.Spec.Containers[1].VolumeMounts[0].MountPath != "/home/journal" {
		t.Errorf("MountPath error")
	}
}

// TestOtherResources ensures that this admission controller is a no-op for other resources,
// subresources, and non-pods.
func TestOtherResources(t *testing.T) {
	namespace := "testnamespace"
	name := "testname"
	pod := &api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "ctr2", Image: "image", ImagePullPolicy: api.PullNever},
			},
		},
	}
	tests := []struct {
		name        string
		kind        string
		resource    string
		subresource string
		object      runtime.Object
		expectError bool
	}{
		{
			name:     "non-pod resource",
			kind:     "Foo",
			resource: "foos",
			object:   pod,
		},
		{
			name:        "pod subresource",
			kind:        "Pod",
			resource:    "pods",
			subresource: "exec",
			object:      pod,
		},
		{
			name:        "non-pod object",
			kind:        "Pod",
			resource:    "pods",
			object:      &api.Service{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		handler := &logDir{}

		err := handler.Admit(admission.NewAttributesRecord(tc.object, nil, api.Kind(tc.kind).WithVersion("version"), namespace, name, api.Resource(tc.resource).WithVersion("version"), tc.subresource, admission.Create, nil))

		if tc.expectError {
			if err == nil {
				t.Errorf("%s: unexpected nil error", tc.name)
			}
			continue
		}

		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}

		if len(pod.Spec.Containers[0].VolumeMounts) != 0 {
			t.Errorf("%s: should not have volumeMount", tc.name)
		}
	}
}
