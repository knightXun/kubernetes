package mqevent

import (
	"sync"
	"github.com/streadway/amqp"

	"strings"
	"github.com/golang/glog"
	"time"
	"fmt"
)

const 	producerQueue     = "service_manage_queue_back"
const 	producerExchange    = "PodLevel"
const   nodeProducerQueue = "NodeLevel_Queue"
const   nodeProducerExchange = "NodeLevel"

var (
	dial    = amqp.Dial
	dialTLS = amqp.DialTLS
)

type MqManager struct {
	Conn      *amqp.Connection
	PodProducer  *amqp.Channel
	NodeProducer *amqp.Channel
	BuildCh   *amqp.Channel
	Incoming  <-chan amqp.Delivery
	url       string
	connected bool
	close     chan bool
	closed    bool
	mtx       sync.Mutex
}

// task's ack message struct
type AckMsg struct {
	//Action         string        `json:"action,omitempty"`
	//Name           string        `json:"name,omitempty"`
	//Namespace      string        `json:"namespace,omitempty"`
	Body           interface{}   `json:"body,omitempty"`
}


func NewRabbitMQConn(MqUrl, MQUsername, MQPasswd, MqType, VHost string) *MqManager {
	if MqUrl == "" || MqUrl[0] == ' ' || !strings.Contains(MqType, "amqp") {
		glog.Errorf("Failed to create rabbitMQ's connection, please enter a valid parameter.")
		return nil
	}
	//连接字符串
	url := MqType + "://" + MQUsername + ":" + MQPasswd + "@" + MqUrl + "/" + VHost

	glog.V(3).Infof("RabbitMQ url %v ", url)
	return &MqManager{
		url:      url,
		close:    make(chan bool),
	}
}

func (mq *MqManager) Init() error {
	glog.Infof("Start to link RabbitMQ: %s", mq.url)
	if err := mq.tryToConnect(); err != nil {
		return err
	}
	mq.connected = true
	go mq.LoopConnect()
	return nil
}

func (mq *MqManager) LoopConnect() {
	for {
		if err := mq.tryToConnect(); err != nil {
			glog.Warningf("Failed to connect rabbitMQ(%s): %v", mq.url, err)
			time.Sleep(1 * time.Second)
			continue
		}
		mq.connected = true

		notyfyClose := make(chan *amqp.Error)
		mq.Conn.NotifyClose(notyfyClose)

		select {
		case <-notyfyClose:
			mq.connected = false
			glog.Warning("MQ Conn closed!")
		case <-mq.close:
			if err := mq.Conn.Close(); err != nil {
				glog.Errorf("Failed to close rabbitMQ connect: %v", err)
			}
			mq.connected = false
			return
		}
	}
}

func (mq *MqManager) IsConnected() bool {
	return mq.connected
}

func (mq *MqManager) Close() {
	mq.mtx.Lock()
	defer mq.mtx.Unlock()

	if mq.closed {
		return
	}

	close(mq.close)
	mq.closed = true
}

// TODO: need to support TLS(amqps)
func (mq *MqManager) tryToConnect() error {
	if mq.connected {
		return nil
	}
	var err error

	mq.Conn, err = dial(mq.url)
	if err != nil {
		return err
	}

	err = mq.createPodProducer()
	if err != nil {
		glog.Warningf("failed to create pod producer by rabbitMQ(%s): %v", mq.url, err)
		return err
	}

	err = mq.createNodeProducer()
	if err != nil {
		glog.Warningf("failed to create node producer by rabbitMQ(%s): %v", mq.url, err)
		return err
	}
	return nil
}

func (mq *MqManager) createPodProducer() error {
	glog.V(6).Infof("Creating Pod Producer.")

	if mq.Conn == nil {
		err := fmt.Errorf("Connection is't existing ")
		glog.Errorf("Please create RabbitMQ connect: %v", err)
		return err
	}
	//create channel
	ch, err := mq.Conn.Channel()
	if err != nil {
		glog.Errorf("Failed to Open a channel: %v", err)
		return err
	}

	err = ch.ExchangeDeclare(
		producerExchange, // name
		"topic",          // type
		true,             // durable
		true,             // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)

	if err != nil {
		glog.Errorf("Failed to declare an exchange: %v", err)
		return err
	}

	_, err = ch.QueueDeclare(
		producerQueue, // name
		true,          // durable
		false,         // delete when usused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)

	if err != nil {
		glog.Errorf("Failed to declare an queue: %v", err)
		return err
	}

	err = ch.QueueBind(
		producerQueue,           // queue name
		"pod.update.response",                // routing key
		producerExchange, // exchange
		false,
		nil,
	)

	if err != nil {
		glog.Errorf("Failed to Bind an queue: %v", err)
		return err
	}
	mq.PodProducer = ch
	return nil
}

func (mq *MqManager) createNodeProducer() error {
	glog.V(6).Infof("Creating Node Producer.")

	if mq.Conn == nil {
		err := fmt.Errorf("Connection is't existing ")
		glog.Errorf("Please create RabbitMQ connect: %v", err)
		return err
	}
	//create channel
	ch, err := mq.Conn.Channel()
	if err != nil {
		glog.Errorf("Failed to Open a channel: %v", err)
		return err
	}

	err = ch.ExchangeDeclare(
		nodeProducerExchange, // name
		"topic",          // type
		true,             // durable
		true,             // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)

	if err != nil {
		glog.Errorf("Failed to declare an Node exchange: %v", err)
		return err
	}

	_, err = ch.QueueDeclare(
		nodeProducerQueue, // name
		true,          // durable
		false,         // delete when usused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)

	if err != nil {
		glog.Errorf("Failed to declare an queue: %v", err)
		return err
	}

	err = ch.QueueBind(
		nodeProducerQueue,           // queue name
		"node.update.response",                // routing key
		nodeProducerExchange, // exchange
		false,
		nil,
	)

	if err != nil {
		glog.Errorf("Failed to Bind an Node queue: %v", err)
		return err
	}

	mq.NodeProducer = ch

	return nil
}

// TODO: Here have a problem filling msg.Body.
func (mqManager *MqManager) SendMsgAck(body string) error {
	if !mqManager.IsConnected() {
		err := mqManager.tryToConnect()
		if err != nil {
			glog.Errorf("Failed to send msg to rabbitMQ, create mq connection: %v", err)
		}
		return err
	}

	if strings.Contains(body, "kubelet.PodLevel") {
		if mqManager.PodProducer == nil {
			err := fmt.Errorf("Channel is't existing ")
			glog.Fatalf("Failed to send Ackmsg: %v", err)
		}


		glog.V(6).Infof("Send a message: Body: %s", string(body))
		msgPsh := amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		}

		err := mqManager.PodProducer.Publish(
			producerExchange, // exchange
			"pod.update.response",      // routing key
			false,            // mandatory
			false,            // immediate
			msgPsh,
		)
		if err != nil {
			glog.Errorf("Failed to publish a message: %v", err)
		}
		return err

	} else if strings.Contains(body, "kubelet.NodeLevel") {
		if mqManager.NodeProducer == nil {
			err := fmt.Errorf("Channel is't existing ")
			glog.Fatalf("Failed to send Ackmsg: %v", err)
		}


		glog.V(6).Infof("Send a message: Body: %s", string(body))
		msgPsh := amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		}

		err := mqManager.NodeProducer.Publish(
			nodeProducerExchange, // exchange
			"node.update.response",      // routing key
			false,            // mandatory
			false,            // immediate
			msgPsh,
		)
		if err != nil {
			glog.Errorf("Failed to publish a message: %v", err)
		}
		return err
	}

	return nil
}