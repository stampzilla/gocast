package events

import "github.com/stampzilla/gocast/responses"

type AppStarted struct {
	*responses.ApplicationSession
}
