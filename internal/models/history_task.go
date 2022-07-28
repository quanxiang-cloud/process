package models

import "gorm.io/gorm"

// HistoryTask info
type HistoryTask struct {
	ID             string `json:"id"`
	ProcID         string `json:"procId"`
	ProcInstanceID string `json:"procInstanceId"`
	ExecutionID    string `json:"executionId"`
	NodeID         string `json:"nodeId"`
	NodeDefKey     string `json:"nodeDefKey"`
	NextNodeDefKey string `json:"nextNodeDefKey"`
	Name           string `json:"name"`
	Desc           string `json:"desc"`
	TaskType       string `json:"taskType"` // Model模型任务、TempModel临时模型任务、NonModel非模型任务
	Assignee       string `json:"assignee"`
	Status         string `json:"status"` // COMPLETED, DELETED
	DueTime        string `json:"dueTime"`
	EndTime        string `json:"endTime"`
	Comments       string `json:"comments"`
	CreatorID      string `json:"creatorId"`
	CreateTime     string `json:"createTime"`
	ModifierID     string `json:"modifierId"`
	ModifyTime     string `json:"modifyTime"`
	TenantID       string `json:"tenantId"`
}

// HistoryTaskRepo interface
type HistoryTaskRepo interface {
	Create(db *gorm.DB, entity *HistoryTask) error
	Update(db *gorm.DB, entity *HistoryTask) error
	FindByInstanceID(db *gorm.DB, page, limit int, req *QueryTaskCondition) ([]*HistoryTask, int64)
	FindPreTask(db *gorm.DB, instanceID, executionID string) ([]*HistoryTask, error)
	FindByIDs(db *gorm.DB, ids []string) ([]*HistoryTask, error)
}
