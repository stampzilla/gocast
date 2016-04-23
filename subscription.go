package gocast

import (
	"fmt"
	"sync/atomic"
	"time"

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

func (s *Subscription) Send(payload responses.Payload) error {
	requestId := int(atomic.AddInt64(&s.requestId, 1))
	payload.SetRequestId(requestId)
	return s.Device.Send(s.Urn, s.SourceId, s.DestinationId, payload)
}

// Request works like send, but waits for resposne to requestId before returning
func (s *Subscription) Request(payload responses.Payload) (*api.CastMessage, error) {
	response := make(chan *api.CastMessage)
	requestId := int(atomic.AddInt64(&s.requestId, 1))
	payload.SetRequestId(requestId)
	s.inFlight[requestId] = response

	err := s.Send(payload)
	if err != nil {
		delete(s.inFlight, requestId)
		return nil, err
	}

	select {
	case reply := <-response:
		return reply, nil
	case <-time.After(time.Second * 5):
		delete(s.inFlight, requestId)
		return nil, fmt.Errorf("Timeout sending")
	}
}

func (s *Subscription) Receive(message *api.CastMessage, headers *responses.Headers) bool {
	// Just skip the message if it isnt to this subscription

	//log.Println(message)
	if *message.SourceId != s.DestinationId || (*message.DestinationId != s.SourceId && *message.DestinationId != "*") || *message.Namespace != s.Urn {
		return false
	}

	//if this is a request we must send the response back to the pending request
	if headers.RequestId != nil && *headers.RequestId != 0 {
		if listener, ok := s.inFlight[*headers.RequestId]; ok {
			listener <- message
			delete(s.inFlight, *headers.RequestId)
		}
	}

	s.Handler.Unmarshal(message.GetPayloadUtf8())
	return true
}
