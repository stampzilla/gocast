package handlers

import (
	"fmt"
	"time"

	"github.com/stampzilla/gocast/events"
)

type Heartbeat struct {
	Dispatch func(events.Event)
	Send     func(Headers) error

	ticker *time.Ticker
}

func (h *Heartbeat) RegisterDispatch(dispatch func(events.Event)) {
	h.Dispatch = dispatch
}
func (h *Heartbeat) RegisterSend(send func(Headers) error) {
	h.Send = send
}

func (h *Heartbeat) Connect() {
	if h.ticker != nil {
		h.ticker.Stop()
		h.ticker = nil
	}

	h.ticker = time.NewTicker(time.Second * 5)
	go func() {
		for {
			<-h.ticker.C
			h.Ping()
		}
	}()

}

func (h *Heartbeat) Disconnect() {
	if h.ticker != nil {
		h.ticker.Stop()
		h.ticker = nil
	}
}

func (h *Heartbeat) Unmarshal(message string) {
	fmt.Println("Heartbeat received: ", message)
}

func (h *Heartbeat) Ping() {
	h.Send(Headers{Type: "PING"})
}

func (h *Heartbeat) Pong() {
	h.Send(Headers{Type: "PONG"})
}
