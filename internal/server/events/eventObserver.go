package events

import (
	"context"
	"errors"
	"github.com/quanxiang-cloud/process/internal/server/rpc"
	"github.com/quanxiang-cloud/process/pkg/config"
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
type ObserverHandler func(ctx context.Context, e *event, EventName string, EventData map[string]string) error

// event event observer
type event struct {
	client *rpc.Client
	pool   *WorkPool
}

// NewEventObserver new event observer
func NewEventObserver(conf *config.Configs) (Observer, error) {
	p := &event{
		client: rpc.NewClient(conf),
		pool:   NewPool(5, 20, 0),
	}
	return p, nil
}

// Name name
func (e event) Name() string {
	return "eventObserver"
}

// Update notify
func (e *event) Update(ctx context.Context, param Param) error {
	if em, ok := param.(*EventMessage); ok {
		switch em.Message.MessageSendMode {
		case SynchronizationMode:
			err := e.Handler(ctx, param)
			return err
		case AsynchronousMode:
			e.pool.Start(ctx).PushTaskFunc(ctx, param, e.Handler)
		}
	}
	return nil
}

// EventMessage event message
type EventMessage struct {
	Message   *Message          `json:"message"`
	EventName string            `json:"eventName"`
	EventData map[string]string `json:"eventData"`
}

// Eval Eval
func (e EventMessage) Eval() {}

// handler event message handler
func (e *event) Handler(ctx context.Context, param Param) error {
	em := param.(*EventMessage)
	if handler, ok := EventHandler[em.Message.MessageType]; ok {
		return handler(ctx, e, em.EventName, em.EventData)
	}
	return errors.New("invalid message name")
}

// MessagePublish message publish function
func MessagePublish(ctx context.Context, e *event, name string, d map[string]string) error {
	req := &rpc.EventPublishReq{
		EventName: name,
		EventType: ProcessEvent,
		EventData: d,
	}
	err := e.client.Publish(req)
	return err
}
