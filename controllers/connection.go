package controllers

import "github.com/barnybug/go-cast"

type connectionController struct {
	channel *cast.Channel
}

var connect = cast.PayloadHeaders{Type: "CONNECT"}
var close = cast.PayloadHeaders{Type: "CLOSE"}

func NewConnectionController(client *cast.Client, sourceId, destinationId string) *connectionController {
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
