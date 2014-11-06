package controllers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ninjasphere/go-cast"
	"github.com/ninjasphere/go-cast/api"
)

type ReceiverController struct {
	interval time.Duration
	channel  *cast.Channel
	Incoming chan *StatusResponse
}

var getStatus = cast.PayloadHeaders{Type: "GET_STATUS"}

func NewReceiverController(client *cast.Client, sourceId, destinationId string) *ReceiverController {
	controller := &ReceiverController{
		channel:  client.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.receiver"),
		Incoming: make(chan *StatusResponse, 0),
	}

	controller.channel.OnMessage("RECEIVER_STATUS", controller.onStatus)

	return controller
}

func (c *ReceiverController) onStatus(message *api.CastMessage) {
	//spew.Dump("Got status message", message)

	response := &StatusResponse{}

	err := json.Unmarshal([]byte(*message.PayloadUtf8), response)

	if err != nil {
		log.Printf("Failed to unmarshal status message:%s - %s", err, *message.PayloadUtf8)
		return
	}

	select {
	case c.Incoming <- response:
	case <-time.After(time.Second * 1):
		log.Printf("Incoming status, but we aren't listening. %v", response)
	}

}

type StatusResponse struct {
	cast.PayloadHeaders
	Status *ReceiverStatus `json:"status,omitempty"`
}

type ReceiverStatus struct {
	cast.PayloadHeaders
	Applications []*ApplicationSession `json:"applications"`
	Volume       *Volume               `json:"volume,omitempty"`
}

type ApplicationSession struct {
	AppID       *string      `json:"appId,omitempty"`
	DisplayName *string      `json:"displayName,omitempty"`
	Namespaces  []*Namespace `json:"namespaces"`
	SessionID   *string      `json:"sessionId,omitempty"`
	StatusText  *string      `json:"statusText,omitempty"`
	TransportId *string      `json:"transportId,omitempty"`
}

type Namespace struct {
	Name string `json:"name"`
}

type Volume struct {
	Level *float64 `json:"level,omitempty"`
	Muted *bool    `json:"muted,omitempty"`
}

func (c *ReceiverController) GetStatus(timeout time.Duration) (*api.CastMessage, error) {
	return c.channel.Request(&getStatus, timeout)
}

func (c *ReceiverController) SetVolume(volume *Volume, timeout time.Duration) (*api.CastMessage, error) {
	return c.channel.Request(&ReceiverStatus{
		PayloadHeaders: cast.PayloadHeaders{Type: "SET_VOLUME"},
		Volume:         volume,
	}, timeout)
}

func (c *ReceiverController) GetVolume(timeout time.Duration) (*Volume, error) {
	message, err := c.GetStatus(timeout)

	if err != nil {
		return nil, err
	}

	response := StatusResponse{}

	err = json.Unmarshal([]byte(*message.PayloadUtf8), &response)

	return response.Status.Volume, err
}
