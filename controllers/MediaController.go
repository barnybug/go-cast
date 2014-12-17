package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-castv2"
	"github.com/ninjasphere/go-castv2/api"
)

type MediaController struct {
	interval       time.Duration
	channel        *castv2.Channel
	Incoming       chan []*MediaStatus
	DestinationID  string
	MediaSessionID int
}

var getMediaStatus = castv2.PayloadHeaders{Type: "GET_STATUS"}

var commandMediaPlay = castv2.PayloadHeaders{Type: "PLAY"}
var commandMediaPause = castv2.PayloadHeaders{Type: "PAUSE"}
var commandMediaStop = castv2.PayloadHeaders{Type: "STOP"}

type MediaCommand struct {
	castv2.PayloadHeaders
	MediaSessionID int `json:"mediaSessionId"`
}

func NewMediaController(client *castv2.Client, sourceId, destinationID string) *MediaController {
	controller := &MediaController{
		channel:       client.NewChannel(sourceId, destinationID, "urn:x-cast:com.google.cast.media"),
		Incoming:      make(chan []*MediaStatus, 0),
		DestinationID: destinationID,
	}

	controller.channel.OnMessage("MEDIA_STATUS", func(message *api.CastMessage) {
		controller.onStatus(message)
	})

	return controller
}

func (c *MediaController) SetDestinationID(id string) {
	c.channel.DestinationId = id
	c.DestinationID = id
}

func (c *MediaController) onStatus(message *api.CastMessage) ([]*MediaStatus, error) {
	spew.Dump("Got media status message", message)

	response := &MediaStatusResponse{}

	err := json.Unmarshal([]byte(*message.PayloadUtf8), response)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal status message:%s - %s", err, *message.PayloadUtf8)
	}

	select {
	case c.Incoming <- response.Status:
	default:
		log.Printf("Incoming status, but we aren't listening. %v", response)
	}

	return response.Status, nil
}

type MediaStatusResponse struct {
	castv2.PayloadHeaders
	Status []*MediaStatus `json:"status,omitempty"`
}

type MediaStatus struct {
	castv2.PayloadHeaders
	MediaSessionID         int                    `json:"mediaSessionId"`
	PlaybackRate           float64                `json:"playbackRate"`
	PlayerState            string                 `json:"playerState"`
	CurrentTime            float64                `json:"currentTime"`
	SupportedMediaCommands int                    `json:"supportedMediaCommands"`
	Volume                 *Volume                `json:"volume,omitempty"`
	CustomData             map[string]interface{} `json:"customData"`
	IdleReason             string                 `json:"idleReason"`
}

func (c *MediaController) GetStatus(timeout time.Duration) ([]*MediaStatus, error) {

	spew.Dump("getting media Status")

	message, err := c.channel.Request(&getMediaStatus, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to get receiver status: %s", err)
	}

	spew.Dump("got media Status", message)

	return c.onStatus(message)
}

func (c *MediaController) Play(timeout time.Duration) (*api.CastMessage, error) {

	message, err := c.channel.Request(&MediaCommand{commandMediaPlay, c.MediaSessionID}, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to send play command: %s", err)
	}

	return message, nil
}

func (c *MediaController) Pause(timeout time.Duration) (*api.CastMessage, error) {

	message, err := c.channel.Request(&MediaCommand{commandMediaPause, c.MediaSessionID}, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to send pause command: %s", err)
	}

	return message, nil
}

func (c *MediaController) Stop(timeout time.Duration) (*api.CastMessage, error) {

	message, err := c.channel.Request(&MediaCommand{commandMediaStop, c.MediaSessionID}, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to send stop command: %s", err)
	}

	return message, nil
}
