package events

import (
	"context"
)

// Observed Observed
type Observed interface {
	Notify(ctx context.Context)
	AddObserve(ob ...Observer)
	RemoveObserve(ob Observer)
}

// Observer Observer
type Observer interface {
	Update(context.Context, Param) error
}

// Param Param
type Param interface {
	Eval()
}

// Message message base info
type Message struct {
	MessageType     string `json:"messageType"`
	MessageSendMode string `json:"messageSendMode"`
}
