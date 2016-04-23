package events

import "github.com/stampzilla/gocast/responses"

type AppStopped struct {
	*responses.ApplicationSession
}
