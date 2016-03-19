package castv2

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/gogo/protobuf/proto"

	"github.com/ninjasphere/go-castv2/api"
)

type Client struct {
	realConn *tls.Conn
	conn     *packetStream
	channels []*Channel
}

type PayloadHeaders struct {
	Type      string `json:"type"`
	RequestId *int   `json:"requestId,omitempty"`
}

func (h *PayloadHeaders) setRequestId(id int) {
	h.RequestId = &id
}

func (h *PayloadHeaders) getRequestId() int {
	return *h.RequestId
}

func NewClient(host net.IP, port int) (*Client, error) {

	log.Printf("connecting to %s:%d ...", host, port)

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Chromecast. Error:%s", err)
	}

	wrapper := NewPacketStream(conn)

	client := &Client{
		realConn: conn,
		conn:     wrapper,
		channels: make([]*Channel, 0),
	}

	/*connection := client.NewChannel("sender-0", "receiver-0", "urn:x-cast:com.google.cast.tp.connection")
	connection.Send(&PayloadHeaders{Type: "CONNECT"})*/

	go func() {
		for {
			packet := wrapper.Read()

			message := &api.CastMessage{}
			err = proto.Unmarshal(*packet, message)
			if err != nil {
				log.Fatalf("Failed to unmarshal CastMessage: %s", err)
			}

			//spew.Dump("Message!", message)

			var headers PayloadHeaders

			err := json.Unmarshal([]byte(*message.PayloadUtf8), &headers)

			if err != nil {
				log.Fatalf("Failed to unmarshal message: %s", err)
			}

			for _, channel := range client.channels {
				channel.message(message, &headers)
			}

		}
	}()

	/*go func() {

		heartbeat := client.NewChannel("sender-0", "receiver-0", "urn:x-cast:com.google.cast.tp.heartbeat")
		ping := PayloadHeaders{Type: "PING"}
		for {
			time.Sleep(5 * time.Second)
			heartbeat.Send(&ping)
		}
	}()*/

	return client, nil
}

func (c *Client) Close() {
	c.realConn.Close()
}

func (c *Client) NewChannel(sourceId, destinationId, namespace string) *Channel {
	channel := &Channel{
		client:        c,
		sourceId:      sourceId,
		DestinationId: destinationId,
		namespace:     namespace,
		listeners:     make([]channelListener, 0),
		inFlight:      make(map[int]chan *api.CastMessage),
	}

	c.channels = append(c.channels, channel)

	return channel
}

func (c *Client) Send(message *api.CastMessage) error {

	proto.SetDefaults(message)

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	//spew.Dump("Writing", message)

	_, err = c.conn.Write(&data)

	return err

}
