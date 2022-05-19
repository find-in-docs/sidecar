package conn

import (
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

type Conn struct {
	nc  *nats.Conn
	js  nats.JetStreamContext
	Url string
}

func NewNATSConn(url string) (*Conn, error) {

	nc, err := nats.Connect(url, nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(3*time.Second),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			fmt.Printf("Reconnecting to NATS server ...\n")
		}))

	if err != nil {
		return nil, fmt.Errorf("Error connecting to NATS server. err: %w", err)
	}

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return nil, fmt.Errorf("Error setting PublishAsyncMaxPending on Jetstream: %w\n",
			err)
	}

	c := Conn{nc, js, url}
	return &c, nil
}

func (c *Conn) SubscribeJS(t string, group string) (*nats.Subscription, error) {

	s, err := c.js.PullSubscribe(t, group, nats.PullMaxWaiting(128))
	if err != nil {
		return nil, fmt.Errorf("Error creating pull subscriber: %w", err)
	}

	return s, nil
}

func (c *Conn) Subscribe(t string, f func(*nats.Msg)) (*nats.Subscription, error) {

	s, err := c.nc.Subscribe(t, f)
	if err != nil {
		fmt.Printf("Error subscribing to NATS server:\n\tserver: %s\n\ttopic: %s\n\terr: %v\n",
			c.Url, t, err)
		os.Exit(-1)
	}

	return s, nil
}

func (c *Conn) Publish(t string, data []byte) error {

	err := c.nc.Publish(t, data)
	if err != nil {
		fmt.Printf("Error publishing to NATS server:\n\tserver: %s\n\ttopic: %s\n\terr: %v\n",
			c.Url, t, err)
		os.Exit(-1)
	}

	return nil
}

func BlockForever() {
	select {} // An empty select blocks forever.
}
