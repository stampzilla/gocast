package gocast

import (
	"crypto/sha256"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/responses"
)

type Subscription struct {
	Urn           string
	SourceId      string
	DestinationId string
	Handler       Handler
	Device        *Device
	requestId     int64
	inFlight      map[int]chan *api.CastMessage
}

func (s *Subscription) Sha256() string {
	data := s.Urn + s.SourceId + s.DestinationId
	sum := sha256.Sum256([]byte(data))
	return string(sum[:])
}

func (s *Subscription) Send(payload responses.Payload) error {
	requestId := int(atomic.AddInt64(&s.requestId, 1))
	payload.SetRequestId(requestId)
	return s.Device.Send(s.Urn, s.SourceId, s.DestinationId, payload)
}

// Request works like send, but waits for resposne to requestId before returning.
func (s *Subscription) Request(payload responses.Payload) (*api.CastMessage, error) {
	requestId := int(atomic.AddInt64(&s.requestId, 1))
	payload.SetRequestId(requestId)

	response := make(chan *api.CastMessage)
	s.inFlight[requestId] = response

	// err := s.Send(payload)
	err := s.Device.Send(s.Urn, s.SourceId, s.DestinationId, payload)
	if err != nil {
		delete(s.inFlight, requestId)
		return nil, err
	}

	delay := time.NewTimer(time.Second * 10)
	select {
	case reply := <-response:
		if !delay.Stop() {
			<-delay.C
		}
		return reply, nil
	case <-delay.C:
		delete(s.inFlight, requestId)
		return nil, fmt.Errorf("Timeout sending: %s", spew.Sdump(payload))
	}
}

func (s *Subscription) Receive(message *api.CastMessage, headers *responses.Headers) bool {
	// Just skip the message if it isnt to this subscription

	if *message.SourceId != s.DestinationId || (*message.DestinationId != s.SourceId && *message.DestinationId != "*") || *message.Namespace != s.Urn {
		return false
	}

	s.Handler.Unmarshal(message.GetPayloadUtf8())

	// if this is a request we must send the response back to the pending request
	if headers.RequestId != nil && *headers.RequestId != 0 {
		if listener, ok := s.inFlight[*headers.RequestId]; ok {
			listener <- message
			delete(s.inFlight, *headers.RequestId)
		}
	}

	return true
}
