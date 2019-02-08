package handlers

import (
	"time"

	"github.com/stampzilla/gocast/responses"
)

type Heartbeat struct {
	OnFailure func()
	baseHandler
	ticker         *time.Ticker
	shutdown       chan struct{}
	receivedAnswer chan struct{}
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
	h.receivedAnswer = make(chan struct{})
	go func() {
		for {
			// Send out a ping
			select {
			case <-h.ticker.C:
				h.Ping()
			case <-h.shutdown:
				return
			}

			// Wait for it to be received
			select {
			case <-time.After(time.Second * 10):
				h.OnFailure()
			case <-h.shutdown:
				return
			case <-h.receivedAnswer:
				// everything great, carry on
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
	// fmt.Println("Heartbeat received: ", message)

	// Try to notify our timeout montor
	select {
	case h.receivedAnswer <- struct{}{}:
	case <-time.After(time.Second):
	}
}

func (h *Heartbeat) Ping() {
	h.Send(&responses.Headers{Type: "PING"})
}

func (h *Heartbeat) Pong() {
	h.Send(&responses.Headers{Type: "PONG"})
}
