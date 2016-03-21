package controllers

import (
	"time"

	"golang.org/x/net/context"

	"github.com/barnybug/go-cast/api"
	"github.com/barnybug/go-cast/log"
	"github.com/barnybug/go-cast/net"
)

const interval = time.Second * 5
const timeoutFactor = 3 // timeouts after 3 intervals

type HeartbeatController struct {
	ticker  *time.Ticker
	channel *net.Channel
}

var ping = net.PayloadHeaders{Type: "PING"}
var pong = net.PayloadHeaders{Type: "PONG"}

func NewHeartbeatController(conn *net.Connection, sourceId, destinationId string) *HeartbeatController {
	controller := &HeartbeatController{
		channel: conn.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.tp.heartbeat"),
	}

	controller.channel.OnMessage("PING", controller.onPing)

	return controller
}

func (c *HeartbeatController) onPing(_ *api.CastMessage) {
	c.channel.Send(pong)
}

func (c *HeartbeatController) Start(ctx context.Context) {
	if c.ticker != nil {
		c.Stop()
	}

	c.ticker = time.NewTicker(interval)
	go func() {
	LOOP:
		for {
			select {
			case <-c.ticker.C:
				c.channel.Send(ping)
				// TODO: handle error
			case <-ctx.Done():
				log.Println("Heartbeat stopped")
				break LOOP
			}
		}
	}()

	log.Println("Heartbeat started")
}

func (c *HeartbeatController) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}
}
