package handlers

import "fmt"

type Connection struct {
	Send func(Headers) error
}

func (c *Connection) SendCallback(send func(Headers) error) {
	c.Send = send
}

func (c *Connection) Connect() {
	c.Send(Headers{Type: "CONNECT"})
}

func (c *Connection) Disconnect() {
	c.Send(Headers{Type: "CLOSE"})
}

func (c *Connection) Unmarshal(message string) {
	fmt.Println("Connection received: ", message)
}
