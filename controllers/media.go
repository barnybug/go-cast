package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/barnybug/go-castv2/api"
	"github.com/barnybug/go-castv2/net"
)

type MediaController struct {
	interval       time.Duration
	channel        *net.Channel
	Incoming       chan []*MediaStatus
	DestinationID  string
	MediaSessionID int
}

const NamespaceMedia = "urn:x-cast:com.google.cast.media"

var getMediaStatus = net.PayloadHeaders{Type: "GET_STATUS"}

var commandMediaPlay = net.PayloadHeaders{Type: "PLAY"}
var commandMediaPause = net.PayloadHeaders{Type: "PAUSE"}
var commandMediaStop = net.PayloadHeaders{Type: "STOP"}
var commandMediaLoad = net.PayloadHeaders{Type: "LOAD"}

type MediaCommand struct {
	net.PayloadHeaders
	MediaSessionID int `json:"mediaSessionId"`
}

type LoadMediaCommand struct {
	MediaCommand
	Media       MediaItem   `json:"media"`
	CurrentTime int         `json:"currentTime"`
	Autoplay    bool        `json:"autoplay"`
	CustomData  interface{} `json:"customData"`
}

type MediaItem struct {
	ContentId   string `json:"contentId"`
	StreamType  string `json:"streamType"`
	ContentType string `json:"contentType"`
}

func NewMediaController(conn *net.Connection, sourceId, destinationID string) *MediaController {
	controller := &MediaController{
		channel:       conn.NewChannel(sourceId, destinationID, NamespaceMedia),
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
	response := &MediaStatusResponse{}

	err := json.Unmarshal([]byte(*message.PayloadUtf8), response)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal status message:%s - %s", err, *message.PayloadUtf8)
	}

	for _, status := range response.Status {
		c.MediaSessionID = status.MediaSessionID
	}

	select {
	case c.Incoming <- response.Status:
	default:
	}

	return response.Status, nil
}

type MediaStatusResponse struct {
	net.PayloadHeaders
	Status []*MediaStatus `json:"status,omitempty"`
}

type MediaStatus struct {
	net.PayloadHeaders
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
	message, err := c.channel.Request(&getMediaStatus, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to get receiver status: %s", err)
	}

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
	if c.MediaSessionID == 0 {
		// no current session to stop
		return nil, nil
	}
	message, err := c.channel.Request(&MediaCommand{commandMediaStop, c.MediaSessionID}, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to send stop command: %s", err)
	}
	return message, nil
}

func (c *MediaController) LoadMedia(media MediaItem, currentTime int, autoplay bool, customData interface{}, timeout time.Duration) (*api.CastMessage, error) {
	message, err := c.channel.Request(&LoadMediaCommand{
		MediaCommand: MediaCommand{
			PayloadHeaders: commandMediaLoad,
			MediaSessionID: c.MediaSessionID,
		},
		Media:       media,
		CurrentTime: currentTime,
		Autoplay:    autoplay,
		CustomData:  customData,
	}, timeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to send stop command: %s", err)
	}
	return message, nil
}
