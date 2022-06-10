package models

import "gorm.io/gorm"

// Execution info
type Execution struct {
	ID             string
	ProcID         string
	ProcInstanceID string
	PID            string
	NodeDefKey     string
	NodeInstanceID string
	IsActive       int8 // 1,0
	CreatorID      string
	CreateTime     string
	ModifierID     string
	ModifyTime     string
	TenantID       string
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
