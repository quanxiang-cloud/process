package models

import "gorm.io/gorm"

// HistoryTask info
type HistoryTask struct {
	ID             string
	ProcID         string
	ProcInstanceID string
	ExecutionID    string
	NodeID         string
	NodeDefKey     string
	NextNodeDefKey string
	Name           string
	Desc           string
	TaskType       string // Model模型任务、TempModel临时模型任务、NonModel非模型任务
	Assignee       string
	Status         string // COMPLETED, DELETED
	DueTime        string
	EndTime        string
	CreatorID      string
	CreateTime     string
	ModifierID     string
	ModifyTime     string
	TenantID       string
	Comments       string
}

// HistoryTaskRepo interface
type HistoryTaskRepo interface {
	Create(db *gorm.DB, entity *HistoryTask) error
	Update(db *gorm.DB, entity *HistoryTask) error
	FindByInstanceID(db *gorm.DB, page, limit int, req *QueryTaskCondition) ([]*HistoryTask, int64)
	FindPreTask(db *gorm.DB, instanceID, executionID string) ([]*HistoryTask, error)
}
