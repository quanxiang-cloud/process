package component

import (
	"errors"
)

var (
	// NodeHandlers task handler define
	NodeHandlers = map[string]INode{}
	// ErrNodeType ErrNodeType
	ErrNodeType = errors.New("no exist node type")
)

// NodeFactory NodeFactory
func NodeFactory(name string) (INode, error) {
	if handler, ok := NodeHandlers[name]; ok {
		return handler, nil
	}
	return nil, ErrNodeType
}
