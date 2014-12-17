package castv2

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ninjasphere/go-castv2/api"
)

type Channel struct {
	client        *Client
	sourceId      string
	DestinationId string
	namespace     string
	requestId     int
	inFlight      map[int]chan *api.CastMessage
	listeners     []channelListener
}

type channelListener struct {
	responseType string
	callback     func(*api.CastMessage)
}

type hasRequestId interface {
	setRequestId(id int)
	getRequestId() int
}

func (c *Channel) message(message *api.CastMessage, headers *PayloadHeaders) {

	//	spew.Dump("XXX", message, c)

	if *message.DestinationId != "*" && (*message.SourceId != c.DestinationId || *message.DestinationId != c.sourceId || *message.Namespace != c.namespace) {
		return
	}

	if *message.DestinationId != "*" && headers.RequestId != nil {
		listener, ok := c.inFlight[*headers.RequestId]
		if !ok {
			log.Printf("Warning: Unknown incoming response id: %d to destination:%s", *headers.RequestId, c.DestinationId)
			return
		}
		listener <- message
		delete(c.inFlight, *headers.RequestId)
		return
	}

	if headers.Type == "" {
		log.Printf("Warning: No message type. Don't know what to do. headers:%v message:%v", headers, message)
		return
	}

	for _, listener := range c.listeners {
		if listener.responseType == headers.Type {
			listener.callback(message)
		}
	}

}

func (c *Channel) OnMessage(responseType string, cb func(*api.CastMessage)) {
	c.listeners = append(c.listeners, channelListener{responseType, cb})
}

func (c *Channel) Send(payload interface{}) error {

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	payloadString := string(payloadJson)

	message := &api.CastMessage{
		ProtocolVersion: api.CastMessage_CASTV2_1_0.Enum(),
		SourceId:        &c.sourceId,
		DestinationId:   &c.DestinationId,
		Namespace:       &c.namespace,
		PayloadType:     api.CastMessage_STRING.Enum(),
		PayloadUtf8:     &payloadString,
	}

	return c.client.Send(message)
}

func (c *Channel) Request(payload hasRequestId, timeout time.Duration) (*api.CastMessage, error) {

	// TODO: Need locking here
	c.requestId++

	payload.setRequestId(c.requestId)

	response := make(chan *api.CastMessage)

	c.inFlight[payload.getRequestId()] = response

	err := c.Send(payload)

	if err != nil {
		delete(c.inFlight, payload.getRequestId())
		return nil, err
	}

	select {
	case reply := <-response:
		return reply, nil
	case <-time.After(timeout):
		delete(c.inFlight, payload.getRequestId())
		return nil, fmt.Errorf("Call to cast channel %s - timed out after %d seconds", c.DestinationId, timeout/time.Second)
	}

}
