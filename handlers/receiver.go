package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast/api"
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
	r.knownApplications = make(map[string]responses.ApplicationSession)
}

func (r *Receiver) Unmarshal(message string) {
	logrus.Debug("Receiver received: ", message)

	response := &responses.ReceiverResponse{}
	err := json.Unmarshal([]byte(message), response)
	if err != nil {
		logrus.Errorf("Failed to unmarshal status message:%s - %s\n", err, message)
		return
	}

	if response.Type != responses.TypeStatus { // Probably an error like: {"reason":"CANCELLED","requestId":2,"type":"LAUNCH_ERROR"}
		logrus.Debugf("Type RECEIVER_STATUS expected got: %s", response.Type)
		return
	}

	prev := make(map[string]responses.ApplicationSession)
	if r.knownApplications == nil {
		r.knownApplications = make(map[string]responses.ApplicationSession)
	}

	// Make a copy of known applications
	for k, v := range r.knownApplications {
		prev[k] = v
	}

	for _, app := range response.Status.Applications {
		// App already running
		if _, ok := prev[app.AppID]; ok {
			// Remove it from the list of previous known apps
			delete(prev, app.AppID)
			continue
		}

		// New app, add it to the list
		r.knownApplications[app.AppID] = *app

		r.Dispatch(events.AppStarted{app})
	}

	// Loop thru all stopped apps
	for key, app := range prev {
		delete(r.knownApplications, key)

		r.Dispatch(events.AppStopped{&app})
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

var ErrAppAlreadyLaunched = fmt.Errorf("app already launched")

func (r *Receiver) LaunchApp(appId string) error {
	if app := r.GetSessionByAppId(appId); app != nil {
		return ErrAppAlreadyLaunched
	}

	_, err := r.Request(&LaunchRequest{
		Headers: responses.Headers{Type: "LAUNCH"},
		AppId:   appId,
	})
	return err
}

// TODO maybe do 0-100 instead of 0.0 to 1.0?
func (r *Receiver) SetVolume(volume float64) (*api.CastMessage, error) {
	return r.Request(&responses.ReceiverStatus{
		Headers: responses.Headers{Type: "SET_VOLUME"},
		Volume: &responses.Volume{
			Level: volume,
		},
	})
}
