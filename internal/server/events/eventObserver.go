package events

import (
	"context"
	"errors"
	"git.internal.yunify.com/qxp/process/pkg/config"
	rpc2 "git.internal.yunify.com/qxp/process/rpc"
	"git.internal.yunify.com/qxp/process/rpc/pb"
)

const (
	// MessageType MessageType
	MessageType = "eventMessage"
	// AsynchronousMode AsynchronousMode
	AsynchronousMode = "asynchronous"
	// SynchronizationMode SynchronizationMode
	SynchronizationMode = "synchronization"
	// MessagePublishEvent MessagePublishEvent
	MessagePublishEvent = "messagePublish"
	// ProcessEvent event type
	ProcessEvent = "processEvent"
)

var (
	// EventHandler event Handler define
	EventHandler = map[string]ObserverHandler{MessageType: MessagePublish}
)

// ObserverHandler handler function
type ObserverHandler func(ctx context.Context, e *event, eventType string, data *pb.NodeEventReqData) (*pb.NodeEventRespData, error)

// event event observer
type event struct {
	client *rpc2.Client
	pool   *WorkPool
}

// NewEventObserver new event observer
func NewEventObserver(conf *config.Configs) (Observer, error) {
	p := &event{
		client: rpc2.NewClient(conf),
		pool:   NewPool(5, 20, 0),
	}
	return p, nil
}

// Name name
func (e event) Name() string {
	return "eventObserver"
}

// Update notify
func (e *event) Update(ctx context.Context, param Param) (*pb.NodeEventRespData, error) {
	if em, ok := param.(*EventMessage); ok {
		switch em.Message.MessageSendMode {
		case SynchronizationMode:
			return e.Handler(ctx, param)
		case AsynchronousMode:
			e.pool.Start(ctx).PushTaskFunc(ctx, param, e.Handler)
		}
	}
	return nil, nil
}

// EventMessage event message
type EventMessage struct {
	Message   *Message             `json:"message"`
	EventName string               `json:"eventName"`
	EventData *pb.NodeEventReqData `json:"eventData"`
}

// Eval Eval
func (e EventMessage) Eval() {}

// handler event message handler
func (e *event) Handler(ctx context.Context, param Param) (*pb.NodeEventRespData, error) {
	em := param.(*EventMessage)
	if handler, ok := EventHandler[em.Message.MessageType]; ok {
		return handler(ctx, e, em.EventName, em.EventData)
	}
	return nil, errors.New("invalid message name")
}

// MessagePublish message publish function
func MessagePublish(ctx context.Context, e *event, eventType string, data *pb.NodeEventReqData) (*pb.NodeEventRespData, error) {
	req := &pb.NodeEventReq{
		EventType: eventType,
		Data:      data,
	}
	resp, err := e.client.NodeEventPublish(ctx, req)

	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
