package models

import "gorm.io/gorm"

// Task info
type Task struct {
	ID             string
	ProcID         string
	ProcInstanceID string
	ExecutionID    string
	NodeID         string
	NodeDefKey     string
	NextNodeDefKey string
	Name           string
	Desc           string // 例如：审批、填写、抄送、阅示
	TaskType       string // Model模型任务、TempModel临时模型任务、NonModel非模型任务
	Assignee       string
	Status         string // COMPLETED, ACTIVE
	DueTime        string
	EndTime        string
	CreatorID      string
	CreateTime     string
	ModifierID     string
	ModifyTime     string
	TenantID       string
	Comments       string
}

// TaskVO TaskVO
type TaskVO struct {
	Task
	NodeInstanceID  string
	NodeInstancePid string
}

// QueryTaskCondition condition
type QueryTaskCondition struct {
	Des        string       `json:"desc"`
	Name       string       `json:"taskName"`
	UserID     string       `json:"userID"`
	GroupID    string       `json:"groupID"`
	Order      []QueryOrder `json:"orders"`
	NodeDefKey string       `json:"nodeDefKey"`
	ProcessID  []string     `json:"processID"`
	InstanceID []string     `json:"instanceID"`
	TaskID     []string     `json:"taskID"`
	DueTime    string       `json:"dueTime"`
	Status     string       `json:"status"`
	Assignee   string       `json:"assignee"`
}

// QueryOrder QueryOrder
type QueryOrder struct {
	OrderType string `json:"orderType"`
	Column    string `json:"column"`
}

// TaskRepo interface
type TaskRepo interface {
	Create(db *gorm.DB, entity *Task) error
	Update(db *gorm.DB, entity *Task) error
	DeleteByID(db *gorm.DB, id string) error
	FindByID(db *gorm.DB, id string) (*Task, error)
	FindByIDs(db *gorm.DB, ids []string) ([]*Task, error)
	DeleteByInstanceID(db *gorm.DB, instanceID string) error
	FindByExecutionIDs(db *gorm.DB, ids []string) ([]*Task, error)
	FindPageByInstanceID(db *gorm.DB, page, limit int, req *QueryTaskCondition) ([]*Task, int64)
	FindByInstanceID(db *gorm.DB, instanceID string) ([]*Task, error)
	FindPageByCondition(db *gorm.DB, page, limit int, req *QueryTaskCondition) ([]*TaskVO, int64)
	FindAllPageByCondition(db *gorm.DB, page, limit int, req *QueryTaskCondition) ([]*Task, int64)
}
