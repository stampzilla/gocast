package responses

const TypePing = "PING"
const TypePong = "PONG"
const TypeStatus = "RECEIVER_STATUS"
const TypeAppAvailability = "GET_APP_AVAILABILITY"
const TypeInvalid = "INVALID_REQUEST"
const TypeMediaStatus = "MEDIA_STATUS"
const TypeClose = "CLOSE"
const TypeLoadFailed = "LOAD_FAILED"
const TypeLaunchError = "LAUNCH_ERROR"

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
