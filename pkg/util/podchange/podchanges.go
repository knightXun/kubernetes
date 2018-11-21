package podchange

import (
	"encoding/json"
	"k8s.io/client-go/tools/record"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

// Both RcPodChange and RcStatusChange will use it.
type RcChangeEvent struct {
	EventType     string `json:"eventType,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	PodName       string `json:"podName,omitempty"`
	RcName        string `json:"rcName,omitempty"`
	Action        string `json:"action,omitempty"`
	Room          string `json:"room,omitempty"`
	ReadyReplicas int32  `json:"readyReplicas"`
}

// Depracated
type StatefulsetChangeEvent struct {
	EventType     string `json:"eventType,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	PodName       string `json:"podName,omitempty"`
	SsName        string `json:"ssName,omitempty"`
	Action        string `json:"action,omitempty"`
	Room          string `json:"room,omitempty"`
	ReadyReplicas int32  `json:"readyReplicas"`
}

type JobChangeEvent struct {
	EventType string `json:"eventType,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	PodName   string `json:"podName,omitempty"`
	JobName   string `json:"jobName,omitempty"`
	Action    string `json:"action,omitempty"`
}

type RcAutoScaleInfo struct {
	EventType  string `json:"eventType,omitempty"`
	RcName     string `json:"rcName,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	CurrentNum int32  `json:"currentNum,omitempty"`
	DesiredNum int32  `json:"desiredNum,omitempty"`
	Status     string `json:"status,omitempty"`
}

type PodChangeEvent struct {
	EventType string `json:"eventType,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	PodName   string `json:"podName,omitempty"`
	SsName    string `json:"ssName,omitempty"`
	RcName    string `json:"rcName,omitempty"`
	Action    string `json:"action,omitempty"`
	JobName   string `json:"jobName,omitempty"`
	Room      string `json:"room,omitempty"`
}

func RecordPodEvent(recorder record.EventRecorder, pod *v1.Pod, event, action string) {
	if len(pod.OwnerReferences) != 1 {
		glog.Warningf("%v/%v's OwnerReferences is illegal!", pod.Namespace, pod.Name)
		return
	}
	kind := pod.OwnerReferences[0].Kind
	podChangeEvent := PodChangeEvent{
		Namespace: pod.Namespace,
		PodName:   pod.Name,
		EventType: event,
		Action:    action,
	}
	var reason string
	switch kind {
	case "ReplicationController":
		podChangeEvent.RcName = pod.OwnerReferences[0].Name
		reason = "RcUpdate"
	case "StatefulSet":
		podChangeEvent.SsName = pod.OwnerReferences[0].Name
		reason = "StatefulSetStatusUpdate"
	case "Job":
		podChangeEvent.JobName = pod.OwnerReferences[0].Name
		reason = "JobUpdate"
	}
	message, _ := json.Marshal(podChangeEvent)

	recorder.Event(pod, v1.EventTypeNormal, reason, string(message))
}

func RecorcRCAutoScaleEvent(recorder record.EventRecorder, rcName, namespace, eventType string, currentNum, desiredNum int32, status string) {
	ref := &v1.ObjectReference{
		Kind:      "replication-controller",
		Name:      "",
		Namespace: namespace,
	}

	autoScaleInfo := RcAutoScaleInfo{
		Namespace:  namespace,
		RcName:     rcName,
		CurrentNum: currentNum,
		DesiredNum: desiredNum,
		EventType:  eventType,
		Status:     status,
	}
	message, _ := json.Marshal(autoScaleInfo)

	recorder.Eventf(ref, v1.EventTypeNormal, "RcUpdate", "%s", string(message))
}

// TODO: should extract useful information and pass them through events. Eliminate the HTTP requests to get pods/RCs in servicemanager!
func RecordRCStatusEvent(recorder record.EventRecorder, rcName, namespace, event, action string, labels map[string]string, readyReplicas int32) {
	ref := &v1.ObjectReference{
		Kind:      "replication-controller",
		Name:      rcName,
		Namespace: namespace,
	}
	rcChangeEvent := RcChangeEvent{
		RcName:        rcName,
		Namespace:     namespace,
		EventType:     event,
		Action:        action,
		Room:          labels["room"],
		ReadyReplicas: readyReplicas,
	}
	message, _ := json.Marshal(rcChangeEvent)

	recorder.Eventf(ref, v1.EventTypeNormal, "RcStatusUpdate", "%s", string(message))
}

func RecordRCPodEvent(recorder record.EventRecorder, rcName, namespace, podName, event, action string) {
	ref := &v1.ObjectReference{
		Kind:      "replication-controller",
		Name:      podName,
		Namespace: namespace,
	}
	rcChangeEvent := RcChangeEvent{
		RcName:    rcName,
		Namespace: namespace,
		PodName:   podName,
		EventType: event,
		Action:    action,
	}
	message, _ := json.Marshal(rcChangeEvent)

	recorder.Eventf(ref, v1.EventTypeNormal, "RcUpdate", "%s", string(message))
}

func RecordStatefulSetStatusEvent(recorder record.EventRecorder, ssName, namespace, event, action string, labels map[string]string, readyReplicas int32) {
	ref := &v1.ObjectReference{
		Kind:      "StatefulSet",
		Name:      ssName,
		Namespace: namespace,
	}
	changeEvent := StatefulsetChangeEvent{
		SsName:        ssName,
		Namespace:     namespace,
		EventType:     event,
		Action:        action,
		Room:          labels["room"],
		ReadyReplicas: readyReplicas,
	}
	message, _ := json.Marshal(changeEvent)

	recorder.Eventf(ref, v1.EventTypeNormal, "StatefulSetStatusUpdate", "%s", string(message))
}

// Depracated: use RecordPodEvent instead.
func RecordStatefulSetPodEvent(recorder record.EventRecorder, po *v1.Pod, ssName, namespace, event, action string) {
	ref := &v1.ObjectReference{
		Kind:      "StatefulSet",
		Name:      ssName,
		Namespace: namespace,
	}
	changeEvent := StatefulsetChangeEvent{
		SsName:    ssName,
		Namespace: namespace,
		PodName:   po.Name,
		EventType: event,
		Action:    action,
	}
	message, _ := json.Marshal(changeEvent)

	recorder.Eventf(ref, v1.EventTypeNormal, "StatefulSetUpdate", "%s", string(message))
}

func RecordJobPodEvent(recorder record.EventRecorder, jobName, namespace, podName, event, action string) {
	ref := &v1.ObjectReference{
		Kind:      "job-controller",
		Name:      podName,
		Namespace: namespace,
	}

	changeEvent := JobChangeEvent{
		JobName:   jobName,
		Namespace: namespace,
		PodName:   podName,
		EventType: event,
		Action:    action,
	}
	message, _ := json.Marshal(changeEvent)

	recorder.Eventf(ref, v1.EventTypeNormal, "JobUpdate", "%s", string(message))
}
