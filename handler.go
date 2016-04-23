package gocast

import (
	"github.com/stampzilla/gocast/api"
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

type Handler interface {
	RegisterSend(func(responses.Payload) error)
	RegisterRequest(func(responses.Payload) (*api.CastMessage, error))
	RegisterDispatch(func(events.Event))
	Send(responses.Payload) error
	Request(responses.Payload) (*api.CastMessage, error)

	Connect()
	Disconnect()
	Unmarshal(string)
}
