package controllers

import "github.com/ninjasphere/go-castv2"

type connectionController struct {
	channel *castv2.Channel
}

var connect = castv2.PayloadHeaders{Type: "CONNECT"}
var close = castv2.PayloadHeaders{Type: "CLOSE"}

func NewConnectionController(client *castv2.Client, sourceId, destinationId string) *connectionController {
	controller := &connectionController{
		channel: client.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.tp.connection"),
	}

	return controller
}

func (c *connectionController) Connect() {
	c.channel.Send(connect)
}

func (c *connectionController) Close() {
	c.channel.Send(close)
}
