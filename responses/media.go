package responses

type MediaCommand struct {
	Headers
	MediaSessionID int `json:"mediaSessionId"`
}

type LoadMediaCommand struct {
	Headers
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
	Headers
	Status []*MediaStatus `json:"status,omitempty"`
}

type MediaStatus struct {
	Headers
	MediaSessionID         int                    `json:"mediaSessionId"`
	PlaybackRate           float64                `json:"playbackRate"`
	PlayerState            string                 `json:"playerState"`
	CurrentTime            float64                `json:"currentTime"`
	SupportedMediaCommands int                    `json:"supportedMediaCommands"`
	Volume                 *Volume                `json:"volume,omitempty"`
	Media                  *MediaStatusMedia      `json:"media"`
	CustomData             map[string]interface{} `json:"customData"`
	RepeatMode             string                 `json:"repeatMode"`
	IdleReason             string                 `json:"idleReason"`
}

type SeekMediaCommand struct {
	Headers
	CurrentTime    int `json:"currentTime"`
	MediaSessionID int `json:"mediaSessionId"`
}
