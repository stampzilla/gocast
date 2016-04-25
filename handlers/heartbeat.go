package handlers

import (
	"time"

	"github.com/stampzilla/gocast/responses"
)

type Heartbeat struct {
	baseHandler
	ticker   *time.Ticker
	shutdown chan struct{}
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
	//fmt.Println("Heartbeat received: ", message)
}

func (h *Heartbeat) Ping() {
	h.Send(&responses.Headers{Type: "PING"})
}

func (h *Heartbeat) Pong() {
	h.Send(&responses.Headers{Type: "PONG"})
}
