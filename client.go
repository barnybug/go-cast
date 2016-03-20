package cast

import (
	"net"
	"time"

	"golang.org/x/net/context"

	"github.com/barnybug/go-cast/controllers"
	"github.com/barnybug/go-cast/log"
	castnet "github.com/barnybug/go-cast/net"
)

type Client struct {
	conn       *castnet.Connection
	ctx        context.Context
	cancel     context.CancelFunc
	heartbeat  *controllers.HeartbeatController
	connection *controllers.ConnectionController
	receiver   *controllers.ReceiverController
	media      *controllers.MediaController
}

const DefaultSender = "sender-0"
const DefaultReceiver = "receiver-0"

func NewClient() *Client {
	return &Client{ctx: context.Background()}
}

func (c *Client) Connect(host net.IP, port int) error {
	c.conn = castnet.NewConnection()
	c.conn.Connect(host, port)

	ctx, cancel := context.WithCancel(c.ctx)
	c.cancel = cancel

	// connect channel
	c.connection = controllers.NewConnectionController(c.conn, DefaultSender, DefaultReceiver)
	c.connection.Connect()

	// start heartbeat
	c.heartbeat = controllers.NewHeartbeatController(c.conn, DefaultSender, DefaultReceiver)
	c.heartbeat.Start(ctx)

	// start receiver
	c.receiver = controllers.NewReceiverController(c.conn, DefaultSender, DefaultReceiver)

	return nil
}

func (c *Client) NewChannel(sourceId, destinationId, namespace string) *castnet.Channel {
	return c.conn.NewChannel(sourceId, destinationId, namespace)
}

func (c *Client) Close() {
	c.cancel()
	c.conn.Close()
}

func (c *Client) Receiver() *controllers.ReceiverController {
	return c.receiver
}

func (c *Client) launchMediaApp() string {
	// get transport id
	status, err := c.receiver.GetStatus(5 * time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	transportId := status.GetTransportId(controllers.NamespaceMedia)
	if transportId != "" {
		return transportId
	}
	// needs launching
	status, err = c.receiver.LaunchApp(AppMedia, 5*time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	transportId = status.GetTransportId(controllers.NamespaceMedia)
	return transportId
}

func (c *Client) IsPlaying() bool {
	status, err := c.receiver.GetStatus(5 * time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	app := status.GetSessionByNamespace(controllers.NamespaceMedia)
	if app == nil {
		return false
	}
	if *app.StatusText == "Ready To Cast" {
		return false
	}
	return true
}

func (c *Client) Media() *controllers.MediaController {
	if c.media == nil {
		transportId := c.launchMediaApp()
		conn := controllers.NewConnectionController(c.conn, DefaultSender, transportId)
		conn.Connect()
		c.media = controllers.NewMediaController(c.conn, DefaultSender, transportId)
		c.media.GetStatus(5 * time.Second)
	}
	return c.media
}
