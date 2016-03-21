package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/barnybug/go-cast/api"
	"github.com/barnybug/go-cast/log"
	"github.com/barnybug/go-cast/net"
)

type ReceiverController struct {
	interval time.Duration
	channel  *net.Channel
	Incoming chan *ReceiverStatus
	status   *ReceiverStatus
}

var getStatus = net.PayloadHeaders{Type: "GET_STATUS"}
var commandLaunch = net.PayloadHeaders{Type: "LAUNCH"}
var commandStop = net.PayloadHeaders{Type: "STOP"}

func NewReceiverController(conn *net.Connection, sourceId, destinationId string) *ReceiverController {
	controller := &ReceiverController{
		channel:  conn.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.receiver"),
		Incoming: make(chan *ReceiverStatus, 0),
	}

	controller.channel.OnMessage("RECEIVER_STATUS", controller.onStatus)

	return controller
}

func (c *ReceiverController) onStatus(message *api.CastMessage) {
	response := &StatusResponse{}
	err := json.Unmarshal([]byte(*message.PayloadUtf8), response)
	if err != nil {
		log.Errorf("Failed to unmarshal status message:%s - %s", err, *message.PayloadUtf8)
		return
	}

	c.status = response.Status
	select {
	case c.Incoming <- response.Status:
	case <-time.After(time.Second):
	}
}

type StatusResponse struct {
	net.PayloadHeaders
	Status *ReceiverStatus `json:"status,omitempty"`
}

type ReceiverStatus struct {
	net.PayloadHeaders
	Applications []*ApplicationSession `json:"applications"`
	Volume       *Volume               `json:"volume,omitempty"`
}

type LaunchRequest struct {
	net.PayloadHeaders
	AppId string `json:"appId"`
}

func (s *ReceiverStatus) GetSessionByNamespace(namespace string) *ApplicationSession {
	for _, app := range s.Applications {
		for _, ns := range app.Namespaces {
			if ns.Name == namespace {
				return app
			}
		}
	}
	return nil
}

func (s *ReceiverStatus) GetSessionByAppId(appId string) *ApplicationSession {
	for _, app := range s.Applications {
		if *app.AppID == appId {
			return app
		}
	}
	return nil
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

func (c *ReceiverController) GetStatus(timeout time.Duration) (*ReceiverStatus, error) {
	message, err := c.channel.Request(&getStatus, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to get receiver status: %s", err)
	}
	c.onStatus(message)

	response := &StatusResponse{}
	err = json.Unmarshal([]byte(*message.PayloadUtf8), response)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal status message: %s - %s", err, *message.PayloadUtf8)
	}

	return response.Status, nil
}

func (c *ReceiverController) SetVolume(volume *Volume, timeout time.Duration) (*api.CastMessage, error) {
	return c.channel.Request(&ReceiverStatus{
		PayloadHeaders: net.PayloadHeaders{Type: "SET_VOLUME"},
		Volume:         volume,
	}, timeout)
}

func (c *ReceiverController) GetVolume(timeout time.Duration) (*Volume, error) {
	status, err := c.GetStatus(timeout)
	if err != nil {
		return nil, err
	}
	return status.Volume, err
}

func (c *ReceiverController) LaunchApp(appId string, timeout time.Duration) (*ReceiverStatus, error) {
	message, err := c.channel.Request(&LaunchRequest{
		PayloadHeaders: commandLaunch,
		AppId:          appId,
	}, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed sending request: %s", err)
	}

	response := &StatusResponse{}
	err = json.Unmarshal([]byte(*message.PayloadUtf8), response)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal status message: %s - %s", err, *message.PayloadUtf8)
	}
	return response.Status, nil
}

func (c *ReceiverController) QuitApp(timeout time.Duration) (*api.CastMessage, error) {
	return c.channel.Request(&commandStop, timeout)
}
