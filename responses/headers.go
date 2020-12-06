package responses

const (
	TypePing            = "PING"
	TypePong            = "PONG"
	TypeStatus          = "RECEIVER_STATUS"
	TypeAppAvailability = "GET_APP_AVAILABILITY"
	TypeInvalid         = "INVALID_REQUEST"
	TypeMediaStatus     = "MEDIA_STATUS"
	TypeClose           = "CLOSE"
	TypeLoadFailed      = "LOAD_FAILED"
	TypeLaunchError     = "LAUNCH_ERROR"
)

type Headers struct {
	Type      string `json:"type"`
	RequestId *int   `json:"requestId,omitempty"`
}

func (h *Headers) SetRequestId(id int) {
	h.RequestId = &id
}

func (h *Headers) GetRequestId() int {
	return *h.RequestId
}

type Payload interface {
	SetRequestId(id int)
	GetRequestId() int
}
