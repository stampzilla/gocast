package handlers

import (
	"fmt"

	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

type Connection struct {
	Dispatch func(events.Event)
	send     func(interface{}) error
}

func (c *Connection) RegisterDispatch(dispatch func(events.Event)) {
	c.Dispatch = dispatch
}
func (c *Connection) RegisterSend(send func(interface{}) error) {
	c.send = send
}

func (c *Connection) Send(p interface{}) error {
	return c.send(p)
}

func (c *Connection) Connect() {
	c.Send(responses.Headers{Type: "CONNECT"})
}

func (c *Connection) Disconnect() {
	c.Send(responses.Headers{Type: "CLOSE"})
}

func (c *Connection) Unmarshal(message string) {
	fmt.Println("Connection received: ", message)
}
