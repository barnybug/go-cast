package controllers

import "github.com/barnybug/go-cast/events"
import "github.com/barnybug/go-cast/net"

type ConnectionController struct {
	channel *net.Channel
}

var connect = net.PayloadHeaders{Type: "CONNECT"}
var close = net.PayloadHeaders{Type: "CLOSE"}

func NewConnectionController(conn *net.Connection, eventsCh chan events.Event, sourceId, destinationId string) *ConnectionController {
	controller := &ConnectionController{
		channel: conn.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.tp.connection"),
	}

	return controller
}

func (c *ConnectionController) Connect() error {
	return c.channel.Send(connect)
}

func (c *ConnectionController) Close() error {
	return c.channel.Send(close)
}
