package models

import "gorm.io/gorm"

// Execution info
type Execution struct {
	ID             string `json:"id"`
	ProcID         string `json:"procId"`
	ProcInstanceID string `json:"procInstanceId"`
	PID            string `json:"pId"`
	NodeDefKey     string `json:"nodeDefKey"`
	NodeInstanceID string `json:"nodeInstanceId"`
	IsActive       int8   `json:"isActive"` // 1,0
	CreatorID      string `json:"creatorId"`
	CreateTime     string `json:"createTime"`
	ModifierID     string `json:"modifierId"`
	ModifyTime     string `json:"modifyTime"`
	TenantID       string `json:"tenantId"`
}

// ExecutionRepo interface
type ExecutionRepo interface {
	Create(db *gorm.DB, entity *Execution) error
	Update(db *gorm.DB, ID string, updateMap map[string]interface{}) (*Execution, error)
	DeleteByID(db *gorm.DB, ID string) error
	DeleteByInstanceID(db *gorm.DB, instanceID string) error
	SetActive(db *gorm.DB, ID string, isActive int8) error
	FindByID(db *gorm.DB, ID string) (*Execution, error)
	FindByInstanceID(db *gorm.DB, instanceID string) ([]*Execution, error)
	FindByPID(db *gorm.DB, instanceID string, PID string, active int8) ([]*Execution, error)
	FindAllByPID(db *gorm.DB, instanceID string, PID string) ([]*Execution, error)
	FindInstanceParentExecution(db *gorm.DB, instanceID string) (*Execution, error)
}
