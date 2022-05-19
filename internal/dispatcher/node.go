package dispatcher

import "context"

// ProcessNode component define ,eg serviceNode、crdNode、workflowNode
type ProcessNode interface {
	PreHandler(ctx context.Context, pre Pre, handlers ...PreFunc) error
	Handler(ctx context.Context, req *TaskReq) error
	PostHandler(ctx context.Context, post Post, handlers ...PostFunc) error
}

// Pre pre handler
type Pre map[string]interface{}

// Post post handler
type Post map[string]interface{}

// TaskReq taskReq
type TaskReq struct {
	Node     *NodeReq
	Pre      Pre
	Post     Post
	PreFunc  []PreFunc
	PostFunc []PostFunc
}

// NodeReq component request
type NodeReq struct {
	TaskID            string
	UserID            string
	ProcessInstanceID string
	HandleTask
}

// HandleTask task detail
type HandleTask struct {
	HandleType     string
	HandleDesc     string
	TaskDefKey     string
	HandleUserIds  []string
	CorrelationIds []string
}

// PreFunc pre function
type PreFunc func(ctx context.Context, pre Pre) error

// PostFunc post function
type PostFunc func(ctx context.Context, post Post) error
