package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/stampzilla/gocast/responses"
)

type Media struct {
	baseHandler
	currentStatus  *MediaStatus
	mediaSessionId int
}

func (m *Media) Connect() {
	// Request a new status update
	log.Println("Connecting to media")
	m.Send(&responses.Headers{Type: "GET_STATUS"})
}

func (m *Media) Disconnect() {
	//r.knownApplications = make(map[string]responses.ApplicationSession, 0)
	m.currentStatus = nil
}

func (m *Media) Unmarshal(message string) {
	log.Println("Media received: ", message)

	response := &MediaStatusResponse{}
	err := json.Unmarshal([]byte(message), response)

	if err != nil {
		fmt.Printf("Failed to unmarshal status message:%s - %s\n", err, message)
		return
	}

	//log.Println("MEDIA SESSION ID: ", response.MediaSessionID)
	if len(response.Status) > 0 {
		m.currentStatus = response.Status[0]
	}

	//r.Dispatch(events.Media{
	//Status: response.Status,
	//})
}
func (m *Media) Play() {
	if m.currentStatus != nil {
		m.Request(&MediaCommand{commandMediaPlay, m.currentStatus.MediaSessionID})
	}
}

func (m *Media) Pause() {
	if m.currentStatus != nil {
		m.Request(&MediaCommand{commandMediaPause, m.currentStatus.MediaSessionID})
	}
}

func (m *Media) Stop() {
	if m.currentStatus != nil {
		m.Request(&MediaCommand{commandMediaStop, m.currentStatus.MediaSessionID})
	}
}

func (m *Media) Seek(currentTime int) {
	if m.currentStatus != nil {
		m.Request(&SeekMediaCommand{commandMediaSeek, currentTime, m.currentStatus.MediaSessionID})
	}
}

var getMediaStatus = responses.Headers{Type: "GET_STATUS"}

var commandMediaPlay = responses.Headers{Type: "PLAY"}
var commandMediaPause = responses.Headers{Type: "PAUSE"}
var commandMediaStop = responses.Headers{Type: "STOP"}
var commandMediaLoad = responses.Headers{Type: "LOAD"}
var commandMediaSeek = responses.Headers{Type: "SEEK"}

//TODO move to responses package
type MediaCommand struct {
	responses.Headers
	MediaSessionID int `json:"mediaSessionId"`
}

type LoadMediaCommand struct {
	responses.Headers
	Media       MediaItem   `json:"media"`
	CurrentTime int         `json:"currentTime"`
	Autoplay    bool        `json:"autoplay"`
	CustomData  interface{} `json:"customData"`
}

type MediaItem struct {
	ContentId   string `json:"contentId"`
	StreamType  string `json:"streamType"`
	ContentType string `json:"contentType"`
}

type MediaStatusMedia struct {
	ContentId   string  `json:"contentId"`
	StreamType  string  `json:"streamType"`
	ContentType string  `json:"contentType"`
	Duration    float64 `json:"duration"`
}

type MediaStatusResponse struct {
	responses.Headers
	Status []*MediaStatus `json:"status,omitempty"`
}

type MediaStatus struct {
	responses.Headers
	MediaSessionID         int                    `json:"mediaSessionId"`
	PlaybackRate           float64                `json:"playbackRate"`
	PlayerState            string                 `json:"playerState"`
	CurrentTime            float64                `json:"currentTime"`
	SupportedMediaCommands int                    `json:"supportedMediaCommands"`
	Volume                 *responses.Volume      `json:"volume,omitempty"`
	Media                  *MediaStatusMedia      `json:"media"`
	CustomData             map[string]interface{} `json:"customData"`
	RepeatMode             string                 `json:"repeatMode"`
	IdleReason             string                 `json:"idleReason"`
}

type SeekMediaCommand struct {
	responses.Headers
	CurrentTime    int `json:"currentTime"`
	MediaSessionID int `json:"mediaSessionId"`
}

func (c *Media) LoadMedia(media MediaItem, currentTime int, autoplay bool, customData interface{}) error {
	_, err := c.Request(&LoadMediaCommand{
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
