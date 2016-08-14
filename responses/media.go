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
	ContentId   string        `json:"contentId"`
	StreamType  string        `json:"streamType"`
	ContentType string        `json:"contentType"`
	MetaData    MediaItemMeta `json:"metadata"`
}

type MediaItemMeta struct {
	Title    string               `json:"title"`
	SubTitle string               `json:"subtitle"`
	Images   []MediaItemMetaImage `json:"images"`
}

type MediaItemMetaImage struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type MediaStatusMedia struct {
	ContentId   string        `json:"contentId"`
	StreamType  string        `json:"streamType"`
	ContentType string        `json:"contentType"`
	Duration    float64       `json:"duration"`
	MetaData    MediaItemMeta `json:"metadata"`
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
