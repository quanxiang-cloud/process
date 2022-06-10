package models

import (
	"gorm.io/gorm"
)

// TaskIdentity info
type TaskIdentity struct {
	ID           string
	TaskID       string
	UserID       string
	GroupID      string
	IdentityType string // USER、GROUP
	InstanceID   string
	CreatorID    string
	CreateTime   string
	ModifierID   string
	ModifyTime   string
	TenantID     string
}

// TaskIdentityRepo interface
type TaskIdentityRepo interface {
	Create(db *gorm.DB, entity *TaskIdentity) error
	Update(db *gorm.DB, entity *TaskIdentity) error
	FindPageByUserID(db *gorm.DB, page, limit int, instanceID []string, userID, groupID string) ([]*TaskIdentity, int64)
	FindByUserID(db *gorm.DB, userID, groupID string) ([]string, error)
	DeleteByTaskID(db *gorm.DB, taskID string) error
	DeleteByID(db *gorm.DB, id string) error
	FindUserInstanceTask(db *gorm.DB, instanceID, taskID, userID, groupID string) (*TaskIdentity, error)
	DeleteByInstanceID(db *gorm.DB, instanceID string) error
}
