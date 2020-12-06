package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

type Media struct {
	baseHandler
	currentStatus  *responses.MediaStatus
	mediaSessionId int
}

func (m *Media) Connect() {
	logrus.Info("Connecting to media")
	// Request a new status update
	m.Send(&responses.Headers{Type: "GET_STATUS"})
}

func (m *Media) Disconnect() {
	// r.knownApplications = make(map[string]responses.ApplicationSession, 0)
	m.currentStatus = nil
}

func (m *Media) Unmarshal(message string) {
	logrus.Debug("Media received: ", message)

	response := &responses.MediaStatusResponse{}
	err := json.Unmarshal([]byte(message), response)
	if err != nil {
		logrus.Errorf("Failed to unmarshal status message:%s - %s\n", err, message)
		return
	}

	if len(response.Status) > 0 {
		m.currentStatus = response.Status[0]
		m.Dispatch(events.Media{m.currentStatus})
	}
}

func (m *Media) Play() {
	if m.currentStatus != nil {
		m.Request(&responses.MediaCommand{commandMediaPlay, m.currentStatus.MediaSessionID})
	}
}

func (m *Media) Pause() {
	if m.currentStatus != nil {
		m.Request(&responses.MediaCommand{commandMediaPause, m.currentStatus.MediaSessionID})
	}
}

func (m *Media) Stop() {
	if m.currentStatus != nil {
		m.Request(&responses.MediaCommand{commandMediaStop, m.currentStatus.MediaSessionID})
	}
}

func (m *Media) Seek(currentTime int) {
	if m.currentStatus != nil {
		m.Request(&responses.SeekMediaCommand{commandMediaSeek, currentTime, m.currentStatus.MediaSessionID})
	}
}

var getMediaStatus = responses.Headers{Type: "GET_STATUS"}

var (
	commandMediaPlay  = responses.Headers{Type: "PLAY"}
	commandMediaPause = responses.Headers{Type: "PAUSE"}
	commandMediaStop  = responses.Headers{Type: "STOP"}
	commandMediaLoad  = responses.Headers{Type: "LOAD"}
	commandMediaSeek  = responses.Headers{Type: "SEEK"}
)

func (c *Media) LoadMedia(media responses.MediaItem, currentTime int, autoplay bool, customData interface{}) error {
	_, err := c.Request(&responses.LoadMediaCommand{
		Headers:     commandMediaLoad,
		Media:       media,
		CurrentTime: currentTime,
		Autoplay:    autoplay,
		CustomData:  customData,
	})
	if err != nil {
		return fmt.Errorf("Failed to send load command: %s", err)
	}
	return nil
}
