package handlers

import (
	"log"

	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

type Media struct {
	Dispatch func(events.Event)
	Send     func(responses.Headers) error

	//knownApplications map[string]responses.ApplicationSession
}

func (r *Media) RegisterDispatch(dispatch func(events.Event)) {
	r.Dispatch = dispatch
}
func (r *Media) RegisterSend(send func(responses.Headers) error) {
	r.Send = send
}

func (r *Media) Connect() {
	// Request a new status update
	log.Println("Connecting to media")
	r.Send(responses.Headers{Type: "GET_STATUS"})
}

func (r *Media) Disconnect() {
	//r.knownApplications = make(map[string]responses.ApplicationSession, 0)
}

func (r *Media) Unmarshal(message string) {
	log.Println("Media received: ", message)

	//response := &responses.Media{}
	//err := json.Unmarshal([]byte(message), response)

	//if err != nil {
	//fmt.Printf("Failed to unmarshal status message:%s - %s\n", err, message)
	//return
	//}

	//r.Dispatch(events.Media{
	//Status: response.Status,
	//})
}
