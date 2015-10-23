package handlers

import "fmt"

type Receiver struct {
	Send func(Headers) error
}

func (r *Receiver) SendCallback(send func(Headers) error) {
	r.Send = send
}

func (r *Receiver) Connect() {
	r.Send(Headers{Type: "GET_STATUS"})
}

func (r *Receiver) Disconnect() {
}

func (r *Receiver) Unmarshal(message string) {
	fmt.Println("Receiver received: ", message)
}
