package handlers

import (
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast/responses"
)

type Connection struct {
	baseHandler
}

func (c *Connection) Connect() {
	c.Send(&responses.Headers{Type: "CONNECT"})
}

func (c *Connection) Disconnect() {
	c.Send(&responses.Headers{Type: "CLOSE"})
}

func (c *Connection) Unmarshal(message string) {
	logrus.Info("Connection received: ", message)
}
