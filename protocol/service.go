package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/barnybug/go-cast/api"
	"github.com/barnybug/go-cast/log"
	"github.com/gogo/protobuf/proto"
)

type Message struct {
	Header
	Body
}
type Header struct {
	Type      string `json:"type"`
	RequestId *int   `json:"requestId,omitempty"`
}
type Body struct {
	api.CastMessage
}

type Service struct {
	Conn io.ReadWriteCloser
}

// Receive receives a message
func (s Service) Receive() (*Message, error) {
	var length uint32
	err := binary.Read(s.Conn, binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %s", err)
	}
	if length == 0 {
		return nil, fmt.Errorf("empty packet")
	}

	packet := make([]byte, length)
	_, err = io.ReadFull(s.Conn, packet)
	if err != nil {
		return nil, fmt.Errorf("failed to read full packet: %s", err)
	}

	message := &Message{}
	err = proto.Unmarshal(packet, &message.Body.CastMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %s", err)
	}

	log.Printf("%s ⇐ %s [%s]: %+v",
		*message.DestinationId, *message.SourceId, *message.Namespace, *message.PayloadUtf8)

	err = json.Unmarshal([]byte(*message.PayloadUtf8), &message.Header)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal header: %s", err)
	}
	return message, err
}

// Send sends a payload
func (s Service) Send(payload interface{}, sourceId, destinationId, namespace string) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %s", err)
	}
	payloadString := string(payloadJSON)
	message := &api.CastMessage{
		ProtocolVersion: api.CastMessage_CASTV2_1_0.Enum(),
		SourceId:        &sourceId,
		DestinationId:   &destinationId,
		Namespace:       &namespace,
		PayloadType:     api.CastMessage_STRING.Enum(),
		PayloadUtf8:     &payloadString,
	}

	proto.SetDefaults(message)

	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %s", err)
	}

	log.Printf("%s ⇒ %s [%s]: %s", *message.SourceId, *message.DestinationId, *message.Namespace, *message.PayloadUtf8)

	err = binary.Write(s.Conn, binary.BigEndian, uint32(len(data)))
	if err != nil {
		return fmt.Errorf("failed to write length: %s", err)
	}
	l, err := s.Conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %s", err)
	} else if l != len(data) {
		return fmt.Errorf("data written partially")
	}
	return nil
}

// Close closes the underlying ReadWriteCloser
func (s Service) Close() error {
	return s.Conn.Close()
}
