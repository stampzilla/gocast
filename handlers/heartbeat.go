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

	// TODO take context from parent
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
			delay := time.NewTimer(time.Second * 10)
			select {
			case <-delay.C:
				h.OnFailure()
			case <-ctx.Done():
				if !delay.Stop() {
					<-delay.C
				}
				return
			case <-h.receivedAnswer:
				if !delay.Stop() {
					<-delay.C
				}
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
	delay := time.NewTimer(time.Second)
	select {
	case h.receivedAnswer <- struct{}{}:
		if !delay.Stop() {
			<-delay.C
		}
	case <-delay.C:
	}
}

func (h *Heartbeat) Ping() {
	h.Send(&responses.Headers{Type: "PING"})
}

func (h *Heartbeat) Pong() {
	h.Send(&responses.Headers{Type: "PONG"})
}
