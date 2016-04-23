package responses

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
