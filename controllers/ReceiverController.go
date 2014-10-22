package controllers

import (
	"encoding/json"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-cast"
	"github.com/ninjasphere/go-cast/api"
)

type ReceiverController struct {
	interval time.Duration
	channel  *cast.Channel
}

var getStatus = cast.PayloadHeaders{Type: "GET_STATUS"}

func NewReceiverController(client *cast.Client, sourceId, destinationId string) *ReceiverController {
	controller := &ReceiverController{
		channel: client.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.receiver"),
	}

	controller.channel.OnMessage("RECEIVER_STATUS", controller.onStatus)

	return controller
}

func (c *ReceiverController) onStatus(message *api.CastMessage) {
	spew.Dump("Got status message", message)
}

type StatusResponse struct {
	Status *ReceiverStatus `json:"status,omitempty"`
}

type ReceiverStatus struct {
	Volume *VolumePayload `json:"volume,omitempty"`
}

type VolumePayload struct {
	cast.PayloadHeaders
	Level *float64 `json:"level,omitempty"`
	Muted *bool    `json:"muted,omitempty"`
}

func (c *ReceiverController) GetStatus(timeout time.Duration) (*api.CastMessage, error) {
	return c.channel.Request(&getStatus, timeout)
}

func (c *ReceiverController) SetVolume(volume *VolumePayload, timeout time.Duration) (*api.CastMessage, error) {
	return c.channel.Request(volume, timeout)
}

func (c *ReceiverController) GetVolume(timeout time.Duration) (*VolumePayload, error) {
	message, err := c.GetStatus(timeout)

	if err != nil {
		return nil, err
	}

	response := StatusResponse{}

	err = json.Unmarshal([]byte(*message.PayloadUtf8), &response)

	return response.Status.Volume, err
}
