package podchange

import (
	"encoding/json"
	"k8s.io/client-go/tools/record"
	"k8s.io/api/core/v1"
	"time"
	"github.com/golang/glog"
)

type NodeMessage struct {
	Action        string 		`json:"EventAction"`
	Message       string 		`json:"EventMessage"`
	Reason        string        `json:"EventReason"`
	Time          int64   	`json:"Time"`
}

type NodeMQEventBody struct {
	Kind          string 		`json:"Kind"`
	NodeName      string 		`json:"NodeName"`
	EventType     string 		`json:"EventType"`
	Reason        string 		`json:"Reason"`
	Message       NodeMessage 	`json:"Message"`
}

type PodMessage struct {
	ResourceVersion 	string 		`json:"ResourceVersion"`
	Status              string      `json:"Status"`
	Reason        		string 		`json:"EventReason"`
	Phase               string      `json:"PodPhase"`
	NodeName      		string 		`json:"NodeName"`
	Message       		string 		`json:"EventMessage"`
	PodName             string      `json:"PodName"`
	Pod           		v1.Pod 		`json:"Pod"`
	Time                int64   	`json:"Time"`
}

type PodMQEventBody struct {
	Kind          string 		`json:"Kind"`
	EventType     string 		`json:"EventType"`
	Reason        string 		`json:"Reason"`
	Message       PodMessage 	`json:"Message"`
}

type MQMsg struct {
	TaskCmd       string         `json:"taskCmd"`
	Body          interface{}    `json:"body"`
}

func RecordPodLevelEvent(recorder record.EventRecorder, pod *v1.Pod, eventType, phase, status, reason, msg string) {

	ref := &v1.ObjectReference{
		Kind:      "Kubelet",
	}

	if reason == "MissingClusterDNS" {
		// do nothing
		return
	}

	mqEvent := PodMQEventBody{
		Kind:       "Pod",
		EventType:  eventType,
		Reason:     "PodUpdate",
		Message:    PodMessage{
			ResourceVersion: "",
			Status:    status,
			Reason:    reason,
			NodeName:  pod.Spec.NodeName,
			Phase:     phase,
			Pod:	   *pod,
			PodName:   pod.Name,
			Message:   msg,
			Time:      time.Now().UnixNano() / ( 1000 * 1000),
		},
	}

	mqMsg := MQMsg{
		Body:  mqEvent,
		TaskCmd: "kubelet.PodLevel",
	}
	message, _ := json.Marshal(mqMsg)

	glog.V(1).Infof("Send PodMQEvent %v", string(message))
	recorder.Event(ref, eventType, "PodUpdate",  string(message))
}

func RecordNodeLevelEvent(recorder record.EventRecorder, nodeName, eventType, action, reason, message string) {
	ref := &v1.ObjectReference{
		Kind:      "Kubelet",
	}

	mqEvent := NodeMQEventBody{
		Kind:       "Node",
		NodeName:    nodeName,
		EventType:   eventType,
		Reason:      "NodeUpdate",
		Message:     NodeMessage{
			Reason: 	reason,
			Action:   	action,
			Message:  	message,
			Time:       time.Now().UnixNano() / ( 1000 * 1000),
		},
	}

	mqMsg := MQMsg{
		Body:  mqEvent,
		TaskCmd: "kubelet.NodeLevel",
	}
	msg, _ := json.Marshal(mqMsg)

	glog.V(1).Infof("Send NodeMQEvent %v", string(msg))
	recorder.Event(ref, eventType, "NodeUpdate", string(msg))
}
