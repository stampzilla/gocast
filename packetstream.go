package gocast

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type packetStream struct {
	stream  io.ReadWriteCloser
	packets chan packetContainer
}

type packetContainer struct {
	payload []byte
	err     error
}

func NewPacketStream(stream io.ReadWriteCloser) *packetStream {
	return &packetStream{
		stream:  stream,
		packets: make(chan packetContainer),
	}
}

func (w *packetStream) readPackets(ctx context.Context) {
	var length uint32

	go func() {
		for {
			if ctx.Err() != nil {
				logrus.Errorf("closing packetStream reader %s", ctx.Err())
			}
			err := binary.Read(w.stream, binary.BigEndian, &length)
			if err != nil {
				logrus.Errorf("Failed binary.Read packet: %s", err)
				w.packets <- packetContainer{err: err, payload: nil}
				return
			}

			if length > 0 {
				packet := make([]byte, length)

				i, err := w.stream.Read(packet)
				if err != nil {
					logrus.Errorf("Failed to read packet: %s", err)
					continue
				}

				if i != int(length) {
					logrus.Errorf("Invalid packet size. Wanted: %d Read: %d", length, i)
					continue
				}

				w.packets <- packetContainer{
					payload: packet,
					err:     nil,
				}
			}
		}
	}()
}

func (w *packetStream) Write(data []byte) (int, error) {
	err := binary.Write(w.stream, binary.BigEndian, uint32(len(data)))
	if err != nil {
		err = fmt.Errorf("Failed to write packet length %d. error:%s\n", len(data), err)
		return 0, err
	}

	return w.stream.Write(data)
}
