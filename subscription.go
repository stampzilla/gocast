package gocast

import (
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/responses"
)

type Subscription struct {
	Urn           string
	SourceId      string
	DestinationId string
	Handler       Handler
	Device        *Device
}

func (s *Subscription) Send(payload interface{}) error {
	return s.Device.Send(s.Urn, s.SourceId, s.DestinationId, payload)
}

func (s *Subscription) Receive(message *api.CastMessage, headers *responses.Headers) bool {
	// Just skip the message if it isnt to this subscription

	//log.Println(message)
	if *message.SourceId != s.DestinationId || (*message.DestinationId != s.SourceId && *message.DestinationId != "*") || *message.Namespace != s.Urn {
		return false
	}

	s.Handler.Unmarshal(message.GetPayloadUtf8())
	return true
}
