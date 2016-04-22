package responses

type ReceiverResponse struct {
	Headers
	Status *ReceiverStatus `json:"status,omitempty"`
}

type ReceiverStatus struct {
	Headers
	Applications  []*ApplicationSession `json:"applications"`
	Volume        *Volume               `json:"volume,omitempty"`
	IsStandBy     bool                  `json:"isStandby,omitempty"`
	IsActiveInput bool                  `json:"isActiveInput,omitempty"`
}

type ApplicationSession struct {
	AppID       string      `json:"appId,omitempty"`
	DisplayName string      `json:"displayName,omitempty"`
	Namespaces  []Namespace `json:"namespaces"`
	SessionID   string      `json:"sessionId,omitempty"`
	StatusText  string      `json:"statusText,omitempty"`
	TransportId string      `json:"transportId,omitempty"`
}

type Namespace struct {
	Name string `json:"name"`
}

type Volume struct {
	Level float64 `json:"level,omitempty"`
	Muted bool    `json:"muted,omitempty"`
}
