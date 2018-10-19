package podchange

import (
	"testing"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api"
)

func init() {

}

type FakeEventRecorder struct {
}

func (record FakeEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	fmt.Println(eventtype)
	fmt.Println(reason)
	fmt.Println(message)
}

func (record FakeEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}){
	fmt.Println(eventtype)
	fmt.Println(reason)
	fmt.Println(messageFmt)
	fmt.Println(args...)
}

func (record FakeEventRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}){
	fmt.Println(eventtype)
	fmt.Println(reason)
	fmt.Println(messageFmt)
	fmt.Println(args...)
}

func TestRecorcRCAutoScaleEvent(t *testing.T) {
	rcName := "testrc"
	namespace := "testnamespace"
	eventType := api.EventTypeNormal
	var currentNum int32 = 1
	var desiredNum int32 = 10
	status := "ScaleBegin"

	RecorcRCAutoScaleEvent(FakeEventRecorder{}, rcName, namespace, eventType, currentNum, desiredNum, status)

	currentNum = 10
	desiredNum = 10
	status = "ScaleEnd"
	RecorcRCAutoScaleEvent(FakeEventRecorder{}, rcName, namespace, eventType, currentNum, desiredNum, status)

}