package conn

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

func NewNATSConnJS(nc *nats.Conn) (nats.JetStreamContext, error) {

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return nil, fmt.Errorf("Error creating JetStream: %w", err)
	}

	js.AddStream(&nats.StreamConfig{
		Name:     viper.GetString("nats.jetstream.name"),
		Subjects: []string{viper.GetString("nats.jetstream.subject")},
		Storage:  nats.FileStorage, // default: nats.FileStorage
	})

	return js, nil
}

func (c *Conn) SubscribeJS(topic string, group string) (*nats.Subscription, error) {

	s, err := c.js.PullSubscribe(topic, group, nats.PullMaxWaiting(128))
	if err != nil {
		return nil, fmt.Errorf("Error creating pull subscriber: %w", err)
	}

	return s, nil
}

func (c *Conn) PublishJS(topic string, data []byte) error {

	ret1, ret2 := c.js.Publish(topic, data)
	fmt.Printf("ret1 type: %T ret2 type: %T\n", ret1, ret2)

	return nil
}
