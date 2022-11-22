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
	logger  logrus.FieldLogger
}

type packetContainer struct {
	payload []byte
	err     error
}

func NewPacketStream(stream io.ReadWriteCloser, logger logrus.FieldLogger) *packetStream {
	return &packetStream{
		stream:  stream,
		packets: make(chan packetContainer),
		logger:  logger,
	}
}

func (w *packetStream) readPackets(ctx context.Context) {
	var length uint32

	go func() {
		for {
			if ctx.Err() != nil {
				w.logger.Errorf("closing packetStream reader %s", ctx.Err())
			}
			err := binary.Read(w.stream, binary.BigEndian, &length)
			if err != nil {
				w.logger.Errorf("Failed binary.Read packet: %s", err)
				w.packets <- packetContainer{err: err, payload: nil}
				return
			}

			if length > 0 {
				packet := make([]byte, length)

				i, err := io.ReadFull(w.stream, packet)
				if err != nil {
					w.logger.Errorf("Failed to read packet: %s", err)
					continue
				}

				if i != int(length) {
					w.logger.Errorf("Invalid packet size. Wanted: %d Read: %d Data: %s ... (capped)", length, i, string(packet[:500]))
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
