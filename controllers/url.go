package controllers

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/net/context"

	"github.com/barnybug/go-cast/api"
	"github.com/barnybug/go-cast/events"
	"github.com/barnybug/go-cast/log"
	"github.com/barnybug/go-cast/net"
)

type URLController struct {
	channel       *net.Channel
	eventsCh      chan events.Event
	DestinationID string
	URLSessionID  int
}

const NamespaceURL = "urn:x-cast:com.url.cast"

var getURLStatus = net.PayloadHeaders{Type: "GET_STATUS"}

var commandURLLoad = net.PayloadHeaders{Type: "LOAD"}

type LoadURLCommand struct {
	net.PayloadHeaders
	URL  string `json:"url"`
	Type string `json:"type"`
}

type URLStatusURL struct {
	ContentId   string  `json:"contentId"`
	StreamType  string  `json:"streamType"`
	ContentType string  `json:"contentType"`
	Duration    float64 `json:"duration"`
}

func NewURLController(conn *net.Connection, eventsCh chan events.Event, sourceId, destinationID string) *URLController {
	controller := &URLController{
		channel:       conn.NewChannel(sourceId, destinationID, NamespaceURL),
		eventsCh:      eventsCh,
		DestinationID: destinationID,
	}

	controller.channel.OnMessage("URL_STATUS", controller.onStatus)

	return controller
}

func (c *URLController) SetDestinationID(id string) {
	c.channel.DestinationId = id
	c.DestinationID = id
}

func (c *URLController) sendEvent(event events.Event) {
	select {
	case c.eventsCh <- event:
	default:
		log.Printf("Dropped event: %#v", event)
	}
}

func (c *URLController) onStatus(message *api.CastMessage) {
	response, err := c.parseStatus(message)
	if err != nil {
		log.Errorf("Error parsing status: %s", err)
	}

	for _, status := range response.Status {
		c.sendEvent(*status)
	}
}

func (c *URLController) parseStatus(message *api.CastMessage) (*URLStatusResponse, error) {
	response := &URLStatusResponse{}

	err := json.Unmarshal([]byte(*message.PayloadUtf8), response)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal status message:%s - %s", err, *message.PayloadUtf8)
	}

	for _, status := range response.Status {
		c.URLSessionID = status.URLSessionID
	}

	return response, nil
}

type URLStatusResponse struct {
	net.PayloadHeaders
	Status []*URLStatus `json:"status,omitempty"`
}

type URLStatus struct {
	net.PayloadHeaders
	URLSessionID         int                    `json:"mediaSessionId"`
	PlaybackRate         float64                `json:"playbackRate"`
	PlayerState          string                 `json:"playerState"`
	CurrentTime          float64                `json:"currentTime"`
	SupportedURLCommands int                    `json:"supportedURLCommands"`
	Volume               *Volume                `json:"volume,omitempty"`
	URL                  *URLStatusURL          `json:"media"`
	CustomData           map[string]interface{} `json:"customData"`
	RepeatMode           string                 `json:"repeatMode"`
	IdleReason           string                 `json:"idleReason"`
}

func (c *URLController) Start(ctx context.Context) error {
	_, err := c.GetStatus(ctx)
	return err
}

func (c *URLController) GetStatus(ctx context.Context) (*URLStatusResponse, error) {
	message, err := c.channel.Request(ctx, &getURLStatus)
	if err != nil {
		return nil, fmt.Errorf("Failed to get receiver status: %s", err)
	}

	return c.parseStatus(message)
}

func (c *URLController) LoadURL(ctx context.Context, url string) (*api.CastMessage, error) {
	message, err := c.channel.Request(ctx, &LoadURLCommand{
		PayloadHeaders: commandURLLoad,
		URL:            url,
		Type:           "loc",
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to send load command: %s", err)
	}

	response := &net.PayloadHeaders{}
	err = json.Unmarshal([]byte(*message.PayloadUtf8), response)
	if err != nil {
		return nil, err
	}
	if response.Type == "LOAD_FAILED" {
		return nil, errors.New("Load URL failed")
	}

	return message, nil
}
