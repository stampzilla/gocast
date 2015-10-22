package handlers

type Connection struct {
	Send func(Headers)
}

func (c *Connection) SendCallback(send func(Headers)) {
	c.Send = send
}

func (c *Connection) Connect() {
	c.Send(Headers{Type: "CONNECT"})
}

func (c *Connection) Disconnect() {
	c.Send(Headers{Type: "CLOSE"})
}

func (c *Connection) Unmarshal(message string) {
}
