package conn

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

func NewNATSConnJS(nc *nats.Conn) (nats.JetStreamContext, error) {

	streamName := viper.GetString("nats.jetstream.name")
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return nil, fmt.Errorf("Error creating JetStream: %w", err)
	}

	topic := viper.GetString("nats.jetstream.subject")
	js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{topic},
		Storage:  nats.FileStorage, // default: nats.FileStorage
		NoAck:    false,
	})

	// We dont need to save the consumer info returned, since it is accessible
	// from the NATS API
	durableName := viper.GetString("nats.jetstream.consumer.durableName")
	_, err = js.AddConsumer(streamName, &nats.ConsumerConfig{
		// A durable consumer will pick up where it left
		// off on a re-connection from the subscriber.

		Durable:   durableName,
		AckPolicy: nats.AckExplicitPolicy,
	})
	if err != nil {
		return js, fmt.Errorf("Could not add consumer on stream %s: %w", streamName, err)
	}

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
