package options

import (
	listener "github.com/quanxiang-cloud/process/internal/server/events"
	"gorm.io/gorm"
)

// Opt options interface
type Opt interface {
	SetDB(db *gorm.DB)
	SetListener(l *listener.Listener)
}

// Options type options functions
type Options func(Opt)

// WithDB set db client to OPT
func WithDB(db *gorm.DB) Options {
	return func(o Opt) {
		o.SetDB(db)
	}
}

// WithListener set listener to OPT
func WithListener(l *listener.Listener) Options {
	return func(o Opt) {
		o.SetListener(l)
	}
}
