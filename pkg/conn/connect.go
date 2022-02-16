package conn

import (
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
)

type Conn struct {
	nc  *nats.Conn
	Url string
}

func NewNATSConn(url string) (*Conn, error) {

	nc, err := nats.Connect(url) /* nats.RetryOnFailedConnect(true),
	nats.MaxReconnects(10),
	nats.ReconnectWait(time.Second),
	nats.ReconnectHandler(func(_ *nats.Conn) {
		fmt.Printf("Reconnecting to NATS server ...\n")
	}) */

	if err != nil {
		fmt.Printf("Error connecting to NATS server.\n\terr: %v", err)
		return nil, err
	}

	c := Conn{nc, url}
	return &c, nil
}

func (c *Conn) Subscribe(t string, f func(*nats.Msg)) (*nats.Subscription, error) {

	s, err := c.nc.Subscribe(t, f)
	if err != nil {
		fmt.Printf("Error subscribing to NATS server:\n\tserver: %s\n\ttopic: %s\n\terr: %v\n",
			c.Url, t, err)
		os.Exit(-1)
	}

	fmt.Printf("Subscription: %v\n", s)
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
