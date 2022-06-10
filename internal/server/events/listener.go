package events

import (
	"context"
	"github.com/quanxiang-cloud/process/pkg/config"
)

// Listener listen events
type Listener struct {
	observers []Observer
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
func (l *Listener) Notify(ctx context.Context, param Param) error {
	for _, s := range l.observers {
		err := s.Update(ctx, param)
		if err != nil {
			return err
		}
	}
	return nil
}
