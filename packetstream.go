package gocast

import (
	"encoding/binary"
	"fmt"
	"io"
)

type packetStream struct {
	stream  io.ReadWriteCloser
	packets chan packetContainer
}

type packetContainer struct {
	payload *[]byte
	err     error
}

func NewPacketStream(stream io.ReadWriteCloser) *packetStream {
	wrapper := packetStream{
		stream:  stream,
		packets: make(chan packetContainer),
	}
	wrapper.readPackets()

	return &wrapper
}

func (w *packetStream) readPackets() {
	var length uint32

	go func() {
		for {

			err := binary.Read(w.stream, binary.BigEndian, &length)
			if err != nil {
				fmt.Printf("Failed binary.Read packet: %s", err)
				w.packets <- packetContainer{err: err, payload: nil}
				return
			}

			//TODO make sure this goroutine is killed on disconnect

			if length > 0 {
				packet := make([]byte, length)

				i, err := w.stream.Read(packet)
				if err != nil {
					fmt.Printf("Failed to read packet: %s", err)
					continue
				}

				if i != int(length) {
					fmt.Printf("Invalid packet size. Wanted: %d Read: %d", length, i)
					continue
				}

				w.packets <- packetContainer{
					payload: &packet,
					err:     nil,
				}
			}

		}
	}()
}

func (w *packetStream) Read() (*[]byte, error) {
	pkt := <-w.packets
	if pkt.err != nil {
		close(w.packets)
	}
	return pkt.payload, pkt.err
}

func (w *packetStream) Write(data []byte) (int, error) {

	err := binary.Write(w.stream, binary.BigEndian, uint32(len(data)))

	if err != nil {
		err = fmt.Errorf("Failed to write packet length %d. error:%s\n", len(data), err)
		return 0, err
	}

	return w.stream.Write(data)
}
