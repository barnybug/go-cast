package controllers

import (
	"encoding/json"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-castv2"
	"github.com/ninjasphere/go-castv2/api"
)

type ReceiverController struct {
	interval time.Duration
	channel  *castv2.Channel
}

var getStatus = castv2.PayloadHeaders{Type: "GET_STATUS"}

func NewReceiverController(client *castv2.Client, sourceId, destinationId string) *ReceiverController {
	controller := &ReceiverController{
		channel: client.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.receiver"),
	}

	controller.channel.OnMessage("RECEIVER_STATUS", controller.onStatus)

	return controller
}

func (c *ReceiverController) onStatus(message *api.CastMessage) {
	spew.Dump("Got status message", message)
}

type VolumePayload struct {
	castv2.PayloadHeaders
	Volume *float64 `json:"volume,omitempty"`
	Mute   *bool    `json:"mute,omitempty"`
}

func (c *ReceiverController) GetStatus(timeout time.Duration) (*api.CastMessage, error) {
	return c.channel.Request(&getStatus, timeout)
}

func (c *ReceiverController) SetVolume(volume *VolumePayload, timeout time.Duration) (*api.CastMessage, error) {
	volume.Type = "SET_VOLUME"
	return c.channel.Request(volume, timeout)
}

func (c *ReceiverController) GetVolume(timeout time.Duration) (*VolumePayload, error) {
	message, err := c.GetStatus(timeout)

	if err != nil {
		return nil, err
	}

	var volume VolumePayload

	err = json.Unmarshal([]byte(*message.PayloadUtf8), &volume)

	return &volume, err
}
