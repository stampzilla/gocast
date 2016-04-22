package gocast

import "github.com/stampzilla/gocast/events"

type Handler interface {
	RegisterSend(func(interface{}) error)
	RegisterDispatch(func(events.Event))
	Send(interface{}) error

	Connect()
	Disconnect()
	Unmarshal(string)
}
