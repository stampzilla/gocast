package gocast

import "github.com/stampzilla/gocast/handlers"

type Handler interface {
	SendCallback(func(handlers.Headers) error)

	Connect()
	Disconnect()
	Unmarshal(string)
}
