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
	AppID        string      `json:"appId,omitempty"`
	DisplayName  string      `json:"displayName,omitempty"`
	IsIdleScreen bool        `json:"isIdleScreen"`
	Namespaces   []Namespace `json:"namespaces"`
	SessionID    string      `json:"sessionId,omitempty"`
	StatusText   string      `json:"statusText,omitempty"`
	TransportId  string      `json:"transportId,omitempty"`
}

func (as *ApplicationSession) HasNamespace(ns string) bool {
	for _, v := range as.Namespaces {
		if v.Name == ns {
			return true
		}
	}
	return false
}

type Namespace struct {
	Name string `json:"name"`
}

type Volume struct {
	Level float64 `json:"level,omitempty"`
	Muted bool    `json:"muted,omitempty"`
}
