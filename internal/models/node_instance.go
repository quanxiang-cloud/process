package models

import "gorm.io/gorm"

// NodeInstance info
type NodeInstance struct {
	ID             string `json:"id"`
	ProcID         string `json:"procId"`
	ProcInstanceID string `json:"procInstanceId"`
	PID            string `json:"pId"`
	ExecutionID    string `json:"executionId"`
	NodeDefKey     string `json:"nodeDefKey"`
	NodeName       string `json:"nodeName"`
	NodeType       string `json:"nodeType"`
	TaskID         string `json:"taskId"`
	Comments       string `json:"comments"`
	CreatorID      string `json:"creatorId"`
	CreateTime     string `json:"createTime"`
	ModifierID     string `json:"modifierId"`
	ModifyTime     string `json:"modifyTime"`
	TenantID       string `json:"tenantId"`
}

// NodeInstanceVO vo
type NodeInstanceVO struct {
	*NodeInstance
	Assignee string `json:"assignee"`
}

// NodeInstanceRepo interface
type NodeInstanceRepo interface {
	Create(db *gorm.DB, entity *NodeInstance) error
	FindByID(db *gorm.DB, id string) (*NodeInstance, error)
	FindByTaskID(db *gorm.DB, taskID string) (*NodeInstance, error)
	FindByInstanceID(db *gorm.DB, instanceID string) ([]*NodeInstance, error)
}
