package mqevent

import (
	"k8s.io/api/core/v1"
	"github.com/streadway/amqp"
	"sync"
)

const MSG_TIMEOUT = 3 * 60 // 3min

type MqOptions struct {
	MqUrl    string // rabbitmq's URL
	Username string // rabbitmq's username
	Passwd   string // rabbitmq's passwd
	MqType   string // amqp or amqps
	//TLSConfig tls.config
}

type MqManager struct {
	Conn      *amqp.Connection
	Consummer *amqp.Channel
	Producer  *amqp.Channel
	BuildCh   *amqp.Channel
	Incoming  <-chan amqp.Delivery
	url       string
	connected bool
	close     chan bool
	closed    bool
	mtx       sync.Mutex
}

type MqEvents struct {

}

func NewMqEvents() *MqEvents {
	return nil
}

func (mq *MqEvents) CreateWithEventNamespace(*v1.Event) (*v1.Event, error) {
	return nil, nil
}

func (mq *MqEvents) UpdateWithEventNamespace(*v1.Event) (*v1.Event, error) {
	return nil, nil
}

func (mq *MqEvents) PatchWithEventNamespace(event *v1.Event, data []byte) (result *v1.Event, err error) {
	return nil, nil
}