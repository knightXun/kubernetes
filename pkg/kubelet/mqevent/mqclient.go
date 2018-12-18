package mqevent

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/golang/glog"
)

type MqEvents struct {
	mqManager 	*MqManager

}

func NewMqEvents(MqUrl, MQUsername, MQPasswd, MqType, VHost string) *MqEvents {
	mqManager := NewRabbitMQConn(MqUrl, MQUsername, MQPasswd, MqType, VHost)
	err := mqManager.Init()
	if err != nil {
		glog.V(3).Infof("Can't Create MqEvents %v", err)
	}
	return &MqEvents{
		mqManager: mqManager,
	}
}

func (mq *MqEvents) Event(object runtime.Object, eventtype, reason, message string) {
	go mq.mqManager.SendMsgAck(message)
}

// Eventf is just like Event, but with Sprintf for the message field.
func (mq *MqEvents) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}){
}

// PastEventf is just like Eventf, but with an option to specify the event's 'timestamp' field.
func (mq *MqEvents) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
}