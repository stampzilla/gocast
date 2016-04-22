package gocast

import (
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

type Handler interface {
	RegisterSend(func(responses.Headers) error)
	RegisterDispatch(func(events.Event))

	Connect()
	Disconnect()
	Unmarshal(string)
}
