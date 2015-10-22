package handlers

type Receiver struct {
	Send func(Headers)
}

func (r *Receiver) SendCallback(send func(Headers)) {
	r.Send = send
}

func (r *Receiver) Connect() {
}

func (r *Receiver) Disconnect() {
}

func (r *Receiver) Unmarshal(message string) {
}
