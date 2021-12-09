package handlers

import (
	"context"
	"time"

	"github.com/stampzilla/gocast/responses"
)

type Heartbeat struct {
	OnFailure func()
	baseHandler
	receivedAnswer chan struct{}
	stop           context.CancelFunc
}

func NewHeartbeat() *Heartbeat {
	return &Heartbeat{
		receivedAnswer: make(chan struct{}),
	}
}

func (h *Heartbeat) Connect() {
	if h.stop != nil {
		h.stop()
	}

	//TODO take context from parent
	ctx, s := context.WithCancel(context.Background())
	h.stop = s

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			// Send out a ping
			select {
			case <-ticker.C:
				h.Ping()
			case <-ctx.Done():
				return
			}

			// Wait for it to be received
			select {
			case <-time.After(time.Second * 10):
				h.OnFailure()
			case <-ctx.Done():
				return
			case <-h.receivedAnswer:
				// everything great, carry on
			}
		}
	}()
}

func (h *Heartbeat) Disconnect() {
	if h.stop != nil {
		h.stop()
	}
}

// Unmarshal takes the message and notifies our timeout goroutine to check if we get pong or not.
func (h *Heartbeat) Unmarshal(message string) {
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
