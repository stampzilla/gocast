package handlers

import (
	"fmt"
	"time"

	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

type Heartbeat struct {
	Dispatch func(events.Event)
	send     func(interface{}) error

	ticker   *time.Ticker
	shutdown chan struct{}
}

func (h *Heartbeat) RegisterDispatch(dispatch func(events.Event)) {
	h.Dispatch = dispatch
}
func (h *Heartbeat) RegisterSend(send func(interface{}) error) {
	h.send = send
}

func (h *Heartbeat) Send(p interface{}) error {
	return h.send(p)
}

func (h *Heartbeat) Connect() {
	if h.ticker != nil {
		h.ticker.Stop()
		if h.shutdown != nil {
			close(h.shutdown)
			h.shutdown = nil
		}
	}

	h.ticker = time.NewTicker(time.Second * 5)
	h.shutdown = make(chan struct{})
	go func() {
		for {
			select {
			case <-h.ticker.C:
				h.Ping()
			case <-h.shutdown:
				return
			}
		}
	}()

}

func (h *Heartbeat) Disconnect() {
	if h.ticker != nil {
		h.ticker.Stop()
		if h.shutdown != nil {
			close(h.shutdown)
			h.shutdown = nil
		}
	}
}

func (h *Heartbeat) Unmarshal(message string) {
	fmt.Println("Heartbeat received: ", message)
}

func (h *Heartbeat) Ping() {
	h.Send(responses.Headers{Type: "PING"})
}

func (h *Heartbeat) Pong() {
	h.Send(responses.Headers{Type: "PONG"})
}
