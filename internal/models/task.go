package models

import "gorm.io/gorm"

// Task info
type Task struct {
	ID             string `json:"id"`
	ProcID         string `json:"procId"`
	ProcInstanceID string `json:"procInstanceId"`
	ExecutionID    string `json:"executionId"`
	NodeID         string `json:"nodeId"`
	NodeDefKey     string `json:"nodeDefKey"`
	NextNodeDefKey string `json:"nextNodeDefKey"`
	Name           string `json:"name"`
	Desc           string `json:"desc"`     // 例如：审批、填写、抄送、阅示
	TaskType       string `json:"taskType"` // Model模型任务、TempModel临时模型任务、NonModel非模型任务
	Assignee       string `json:"assignee"`
	Status         string `json:"status"` // COMPLETED, ACTIVE
	DueTime        string `json:"dueTime"`
	EndTime        string `json:"endTime"`
	Comments       string `json:"comments"`
	CreatorID      string `json:"creatorId"`
	CreateTime     string `json:"createTime"`
	ModifierID     string `json:"modifierId"`
	ModifyTime     string `json:"modifyTime"`
	TenantID       string `json:"tenantId"`
}

// TaskVO TaskVO
type TaskVO struct {
	Task
	NodeInstanceID  string `json:"tenantId"`
	NodeInstancePid string `json:"tenantId"`
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
