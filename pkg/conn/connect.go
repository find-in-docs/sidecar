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
		// nats.MaxReconnects(10),   // Defaults to 60 attempts
		nats.ReconnectWait(3*time.Second),  // Defaults to 2s
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			fmt.Printf("Got disconnected! Reason: %q\n", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Printf("Connection closed. Reason: %q\n", nc.LastError())
		}))

	if err != nil {
		return nil, fmt.Errorf("Error connecting to NATS server. err: %w", err)
	}

	c := Conn{nc, nil, url}

	return &c, nil
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
