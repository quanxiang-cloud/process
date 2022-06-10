package models

import "gorm.io/gorm"

// Instance info
type Instance struct {
	ID         string
	ProcID     string
	Name       string
	PID        string
	Status     string // COMPLETED, ACTIVE
	AppStatus  string // ACTIVE,SUSPEND
	EndTime    string
	CreatorID  string
	CreateTime string
	ModifierID string
	ModifyTime string
	TenantID   string
}

// InstanceRepo interface
type InstanceRepo interface {
	Create(db *gorm.DB, entity *Instance) error
	FindByID(db *gorm.DB, id string) (*Instance, error)
	UpdateAppByProcDefKey(db *gorm.DB, procDefKey []string, appStatus string) error
	DeleteAppByProcDefKey(db *gorm.DB, procDefKey []string) error
	DeleteByID(db *gorm.DB, id string) error
	Update(db *gorm.DB, entity *Instance) error
	FindPageInstance(db *gorm.DB, page, limit int, processID, id, parentID []string, name, status string) ([]*Instance, int64)
	FindByUserID(db *gorm.DB, page, limit int, req *QueryTaskCondition) ([]*Instance, int64)
	FindAgencyByUserID(db *gorm.DB, page, limit int, req *QueryTaskCondition) ([]*Instance, int64)
	FindAllByUserID(db *gorm.DB, page, limit int, req *QueryTaskCondition) ([]*Instance, int64)
}
