package messagestream

import (
	"booking-service/config"
	"fmt"
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type ampq struct {
	cfg *config.MessageStreamConfig
}

func NewAmpq(cfg *config.MessageStreamConfig) MessageStream {
	return &ampq{
		cfg: cfg,
	}
}

func (m *ampq) NewSubscriber() (message.Subscriber, error) {
	var ampqURI string
	if m.cfg.SSL {
		ampqURI = fmt.Sprintf("amqps://%s:%s@%s:%s/", m.cfg.Username, m.cfg.Password, m.cfg.Host, m.cfg.Port)
	} else {
		ampqURI = fmt.Sprintf("amqp://%s:%s@%s:%s/", m.cfg.Username, m.cfg.Password, m.cfg.Host, m.cfg.Port)
	}
	ampqConfig := amqp.NewDurableQueueConfig(ampqURI)

	subscriber, err := amqp.NewSubscriber(
		ampqConfig,
		watermill.NewStdLogger(stateLog, stateLog),
	)
	if err != nil {
		log.Fatal(err)
	}

	return subscriber, err
}

func (m *ampq) NewPublisher() (message.Publisher, error) {
	ampqURI := fmt.Sprintf("amqp://%s:%s@%s:%s/", m.cfg.Username, m.cfg.Password, m.cfg.Host, m.cfg.Port)
	ampqConfig := amqp.NewDurableQueueConfig(ampqURI)

	publisher, err := amqp.NewPublisher(
		ampqConfig,
		watermill.NewStdLogger(stateLog, stateLog),
	)
	if err != nil {
		log.Fatal(err)
	}

	return publisher, err
}

func ProcessMessages(messages <-chan *message.Message) {
	for msg := range messages {
		log.Printf("Got message: %s", string(msg.Payload))
		msg.Ack()
	}
}
