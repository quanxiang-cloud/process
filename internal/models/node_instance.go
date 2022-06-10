package models

import "gorm.io/gorm"

// NodeInstance info
type NodeInstance struct {
	ID             string
	ProcID         string
	ProcInstanceID string
	PID            string
	ExecutionID    string
	NodeDefKey     string
	NodeName       string
	NodeType       string
	TaskID         string
	CreatorID      string
	CreateTime     string
	ModifierID     string
	ModifyTime     string
	TenantID       string
}

// NodeInstanceRepo interface
type NodeInstanceRepo interface {
	Create(db *gorm.DB, entity *NodeInstance) error
	FindByID(db *gorm.DB, id string) (*NodeInstance, error)
	FindByTaskID(db *gorm.DB, taskID string) (*NodeInstance, error)
}
