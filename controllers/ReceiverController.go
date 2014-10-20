package controllers

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-cast"
	"github.com/ninjasphere/go-cast/api"
)

type receiverController struct {
	interval time.Duration
	channel  *cast.Channel
}

var getStatus = cast.PayloadHeaders{Type: "GET_STATUS"}

func NewReceiverController(client *cast.Client, sourceId, destinationId string) *receiverController {
	controller := &receiverController{
		channel: client.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.receiver"),
	}

	controller.channel.OnMessage("RECEIVER_STATUS", controller.onStatus)

	return controller
}

func (c *receiverController) onStatus(message *api.CastMessage) {
	spew.Dump("Got status message", message)
}

func (c *receiverController) GetStatus(timeout time.Duration) (*api.CastMessage, error) {
	return c.channel.Request(&getStatus, timeout)
}
