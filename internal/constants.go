package internal

const (
	// Dep department
	Dep = "DEP"
	// AsynchronousMode AsynchronousMode
	AsynchronousMode = "asynchronous"
	// SynchronizationMode SynchronizationMode
	SynchronizationMode = "synchronization"
	// RequestID RequestID
	RequestID = "Request-Id"
	// RedisPreKey redis key
	RedisPreKey = "process:"
)

// flow status
const (
	Completed  = "COMPLETED"
	Active     = "ACTIVE"
	Deleted    = "DELETED"
	Terminated = "TERMINATED"
	Suspend    = "SUSPEND"
)

// task event
const (
	NodeInitBeginEvent   = "nodeInitBeginEvent"
	NodeInitEndEvent     = "nodeInitEndEvent"
	TaskCompleted        = "taskCompletedEvent"
	InclusiveGatewayInit = "inclusiveGatewayInitEvent"
)

// IdentityUser type
const (
	IdentityUser     = "USER"
	IdentityGroup    = "Group"
	IdentityVariable = "VARIABLE"
)

// task type
const (
	ModelTask = "MODEL"
	TempModel = "TEMP_MODEL"
	NonModel  = "NON_MODEL"
)

// app delete or reactive
const (
	DoSuspend  = "suspend"
	DoReactive = "reactive"
	DoDump     = "dump"
)
