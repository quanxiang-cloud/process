package events

import (
	"context"
	"git.internal.yunify.com/qxp/process/rpc/pb"
)

// Observed Observed
type Observed interface {
	Notify(ctx context.Context)
	AddObserve(ob ...Observer)
	RemoveObserve(ob Observer)
}

// Observer Observer
type Observer interface {
	Update(ctx context.Context, param Param) (*pb.NodeEventRespData, error)
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
