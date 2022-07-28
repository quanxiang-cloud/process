package events

import (
	"context"
	"github.com/quanxiang-cloud/process/pkg/config"
	"github.com/quanxiang-cloud/process/pkg/misc/logger"
	"github.com/quanxiang-cloud/process/rpc/pb"
	"math/rand"
	"time"
)

// Listener listen events
type Listener struct {
	observers []Observer
}

// RetryConf retry config
type RetryConf struct {
	Attempts int
	Sleep    time.Duration
	Func     func(ctx context.Context, param Param) (*pb.NodeEventRespData, error)
}

// Stop s
type Stop struct {
	error
}

// NewListener new listener
func NewListener(conf *config.Configs) (*Listener, error) {
	o, err := NewEventObserver(conf)
	if err != nil {
		return nil, err
	}
	l := &Listener{
		observers: make([]Observer, 0),
	}
	l.AddObserve(o)
	return l, nil
}

// AddObserve add listener
func (l *Listener) AddObserve(ob ...Observer) {
	l.observers = append(l.observers, ob...)
}

// RemoveObserve remove listener
func (l *Listener) RemoveObserve(ob Observer) {
	for i, s := range l.observers {
		if s == ob {
			l.observers = append(l.observers[:i], l.observers[i+1:]...)
		}
	}
}

// Notify notify message
func (l *Listener) Notify(ctx context.Context, param Param) (*pb.NodeEventRespData, error) {
	for _, s := range l.observers {
		//return s.Update(ctx, param)
		return l.NotifyRetry(ctx, &param, &RetryConf{
			Attempts: 10,
			Sleep:    time.Second * 5,
			Func:     s.Update,
		})

	}
	return nil, nil
}

// NotifyRetry 发布消息错误重试
func (l *Listener) NotifyRetry(ctx context.Context, param *Param, C *RetryConf) (*pb.NodeEventRespData, error) {
	resp, err := C.Func(ctx, *param)
	if err != nil {
		if C.Attempts--; C.Attempts > 0 {
			logger.Logger.Errorf("Rpc call error. retrying times: %d...\n", C.Attempts)
			// Add some randomness to prevent creating a Thundering Herd
			jitter := time.Duration(rand.Int63n(int64(C.Sleep)))
			C.Sleep = C.Sleep + jitter/2
			time.Sleep(C.Sleep)
			return l.NotifyRetry(ctx, param, C)
		}
		return nil, err
	}
	return resp, nil
}
