package castv2

import (
	"encoding/binary"
	"io"
	"log"
)

type packetStream struct {
	stream  io.ReadWriteCloser
	packets chan *[]byte
}

func NewPacketStream(stream io.ReadWriteCloser) *packetStream {
	wrapper := packetStream{stream, make(chan *[]byte)}
	wrapper.readPackets()

	return &wrapper
}

func (w *packetStream) readPackets() {
	var length uint32

	go func() {
		for {

			err := binary.Read(w.stream, binary.BigEndian, &length)
			if err != nil {
				log.Fatalf("Failed to read packet length: %s", err)
			}

			if length > 0 {
				packet := make([]byte, length)

				i, err := w.stream.Read(packet)
				if err != nil {
					log.Fatalf("Failed to read packet: %s", err)
				}

				if i != int(length) {
					log.Fatalf("Invalid packet size. Wanted: %d Read: %d", length, i)
				}

				log.Printf("Got packet of size %d", length)

				w.packets <- &packet
			}

		}
	}()
}

func (w *packetStream) Read() *[]byte {
	return <-w.packets
}

func (w *packetStream) Write(data *[]byte) (int, error) {

	err := binary.Write(w.stream, binary.BigEndian, uint32(len(*data)))

	if err != nil {
		log.Fatalf("Failed to write packet length %d. error:%s", len(*data), err)
	}

	return w.stream.Write(*data)
}
