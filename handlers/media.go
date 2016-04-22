package handlers

import (
	"log"

	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/responses"
)

type Media struct {
	Dispatch func(events.Event)
	send     func(interface{}) error

	//knownApplications map[string]responses.ApplicationSession
}

func (r *Media) RegisterDispatch(dispatch func(events.Event)) {
	r.Dispatch = dispatch
}
func (r *Media) RegisterSend(send func(interface{}) error) {
	r.send = send
}

func (m *Media) Send(p interface{}) error {
	return m.send(p)
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

var getMediaStatus = responses.Headers{Type: "GET_STATUS"}

var commandMediaPlay = responses.Headers{Type: "PLAY"}
var commandMediaPause = responses.Headers{Type: "PAUSE"}
var commandMediaStop = responses.Headers{Type: "STOP"}
var commandMediaLoad = responses.Headers{Type: "LOAD"}

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

// launch app:
/*
func (c *Client) launchMediaApp(ctx context.Context) (string, error) {
	// get transport id
	status, err := c.receiver.GetStatus(ctx)
	if err != nil {
		return "", err
	}
	app := status.GetSessionByAppId(AppMedia)
	if app == nil {
		// needs launching
		status, err = c.receiver.LaunchApp(ctx, AppMedia)
		if err != nil {
			return "", err
		}
		app = status.GetSessionByAppId(AppMedia)
	}

	if app == nil {
		return "", errors.New("Failed to get media transport")
	}
	return *app.TransportId, nil
}
func (c *MediaController) LoadMedia(ctx context.Context, media MediaItem, currentTime int, autoplay bool, customData interface{}) (*api.CastMessage, error) {
	message, err := c.channel.Request(ctx, &LoadMediaCommand{
		PayloadHeaders: commandMediaLoad,
		Media:          media,
		CurrentTime:    currentTime,
		Autoplay:       autoplay,
		CustomData:     customData,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to send load command: %s", err)
	}

	response := &net.PayloadHeaders{}
	err = json.Unmarshal([]byte(*message.PayloadUtf8), response)
	if err != nil {
		return nil, err
	}
	if response.Type == "LOAD_FAILED" {
		return nil, errors.New("Load media failed")
	}

	return message, nil
}

func (c *MediaController) Play(ctx context.Context) (*api.CastMessage, error) {
	message, err := c.channel.Request(ctx, &MediaCommand{commandMediaPlay, c.MediaSessionID})
	if err != nil {
		return nil, fmt.Errorf("Failed to send play command: %s", err)
	}
	return message, nil
}

func (c *MediaController) Pause(ctx context.Context) (*api.CastMessage, error) {
	message, err := c.channel.Request(ctx, &MediaCommand{commandMediaPause, c.MediaSessionID})
	if err != nil {
		return nil, fmt.Errorf("Failed to send pause command: %s", err)
	}
	return message, nil
}

func (c *MediaController) Stop(ctx context.Context) (*api.CastMessage, error) {
	if c.MediaSessionID == 0 {
		// no current session to stop
		return nil, nil
	}
	message, err := c.channel.Request(ctx, &MediaCommand{commandMediaStop, c.MediaSessionID})
	if err != nil {
		return nil, fmt.Errorf("Failed to send stop command: %s", err)
	}
	return message, nil
}
*/
