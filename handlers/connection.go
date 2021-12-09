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
	logrus.Debug("sending disconnect from connection handler")
	err := c.Send(&responses.Headers{Type: "CLOSE"})
	if err != nil {
		logrus.Error("error sending disconnect: ", err)
	}
}

func (c *Connection) Unmarshal(message string) {
	logrus.Debug("Connection received: ", message)
}
