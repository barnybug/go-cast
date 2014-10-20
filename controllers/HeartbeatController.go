package controllers

import (
	"time"

	"github.com/ninjasphere/go-castv2"
	"github.com/ninjasphere/go-castv2/api"
)

// TODO: Send pings and wait for pongs - https://github.com/thibauts/node-castv2-client/blob/master/lib/controllers/heartbeat.js

const interval = time.Second * 5
const timeoutFactor = 3 // timeouts after 3 intervals

type heartbeatController struct {
	ticker  *time.Ticker
	channel *castv2.Channel
}

var ping = castv2.PayloadHeaders{Type: "PING"}
var pong = castv2.PayloadHeaders{Type: "PONG"}

func NewHeartbeatController(client *castv2.Client, sourceId, destinationId string) *heartbeatController {
	controller := &heartbeatController{
		channel: client.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.tp.heartbeat"),
	}

	controller.channel.OnMessage("PING", controller.onPing)

	return controller
}

func (c *heartbeatController) onPing(_ *api.CastMessage) {
	c.channel.Send(pong)
}

func (c *heartbeatController) Start() {

	if c.ticker != nil {
		c.Stop()
	}

	c.ticker = time.NewTicker(interval)
	go func() {
		for {
			<-c.ticker.C
			c.channel.Send(ping)
		}
	}()

}

func (c *heartbeatController) Stop() {

	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}

}
