package handlers

type Heartbeat struct {
	Send func(Headers) error
}

func (h *Heartbeat) SendCallback(send func(Headers) error) {
	h.Send = send
}

func (h *Heartbeat) Connect() {
}

func (h *Heartbeat) Disconnect() {
}

func (h *Heartbeat) Unmarshal(message string) {
}
