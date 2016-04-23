package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

type Receiver struct {
	baseHandler

	knownApplications map[string]responses.ApplicationSession
	status            *responses.ReceiverStatus
}

func (r *Receiver) Connect() {
	// Request a new status update
	r.Send(&responses.Headers{Type: "GET_STATUS"})
}

func (r *Receiver) Disconnect() {
	r.knownApplications = make(map[string]responses.ApplicationSession, 0)
}

func (r *Receiver) Unmarshal(message string) {
	fmt.Println("Receiver received: ", message)

	response := &responses.ReceiverResponse{}
	err := json.Unmarshal([]byte(message), response)

	if err != nil {
		fmt.Printf("Failed to unmarshal status message:%s - %s\n", err, message)
		return
	}

	prev := make(map[string]responses.ApplicationSession, 0)
	if r.knownApplications == nil {
		r.knownApplications = make(map[string]responses.ApplicationSession, 0)
	}

	// Make a copy of known applications
	for k, v := range r.knownApplications {
		prev[k] = v
	}

	for _, app := range response.Status.Applications {
		// App allready running
		if _, ok := prev[app.AppID]; ok {
			// Remove it from the list of previous known apps
			delete(prev, app.AppID)
			continue
		}

		// New app, add it to the list
		r.knownApplications[app.AppID] = *app

		r.Dispatch(events.AppStarted{app})
		//AppID:       app.AppID,
		//DisplayName: app.DisplayName,
		//TransportId: app.TransportId,
		//})
	}

	// Loop thru all stopped apps
	for key, app := range prev {
		delete(r.knownApplications, key)

		r.Dispatch(events.AppStopped{&app})
		//AppID:       app.AppID,
		//DisplayName: app.DisplayName,
		//TransportId: app.TransportId,
		//})
	}

	r.Dispatch(events.ReceiverStatus{
		Status: response.Status,
	})
	r.status = response.Status
}
func (r *Receiver) GetSessionByAppId(appId string) *responses.ApplicationSession {
	for _, app := range r.knownApplications {
		if app.AppID == appId {
			return &app
		}
	}
	return nil
}

type LaunchRequest struct {
	responses.Headers
	AppId string `json:"appId"`
}

func (r *Receiver) LaunchApp(appId string) error {
	//already launched?
	if app := r.GetSessionByAppId(appId); app != nil {
		return nil
	}

	_, err := r.Request(&LaunchRequest{
		Headers: responses.Headers{Type: "LAUNCH"},
		AppId:   appId,
	})
	return err
}
