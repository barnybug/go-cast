package net

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/barnybug/go-cast/protocol"

	"golang.org/x/net/context"

	"github.com/barnybug/go-cast/api"
	"github.com/barnybug/go-cast/log"
)

type Connection struct {
	conn     *tls.Conn
	channels []*Channel
}

func NewConnection() *Connection {
	return &Connection{
		conn:     nil,
		channels: make([]*Channel, 0),
	}
}

func (c *Connection) NewChannel(sourceId, destinationId, namespace string) *Channel {
	channel := NewChannel(c, sourceId, destinationId, namespace)
	c.channels = append(c.channels, channel)
	return channel
}

func (c *Connection) Connect(ctx context.Context, host net.IP, port int) error {
	var err error
	deadline, _ := ctx.Deadline()
	dialer := &net.Dialer{
		Deadline: deadline,
	}
	c.conn, err = tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:%d", host, port), &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return fmt.Errorf("Failed to connect to Chromecast: %s", err)
	}

	go c.ReceiveLoop()

	return nil
}

func (c *Connection) ReceiveLoop() {
	receiver := protocol.Service{
		Conn: c.conn,
	}
	for {
		message, err := receiver.Receive()
		if err != nil {
			log.Println(err)
			continue
		}

		for _, channel := range c.channels {
			header := PayloadHeaders(message.Header)
			body := api.CastMessage(message.Body.CastMessage)
			channel.Message(&body, &header)
		}
	}
}

func (c *Connection) Send(payload interface{}, sourceId, destinationId, namespace string) error {
	sender := protocol.Service{
		Conn: c.conn,
	}
	return sender.Send(payload, sourceId, destinationId, namespace)
}

func (c *Connection) Close() error {
	// TODO: graceful shutdown
	return c.conn.Close()
}
