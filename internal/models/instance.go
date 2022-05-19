package models

import "gorm.io/gorm"

// Instance info
type Instance struct {
	ID         string `json:"id"`
	ProcID     string `json:"procId"`
	Name       string `json:"name"`
	PID        string `json:"pId"`
	Status     string `json:"status"`    // COMPLETED, ACTIVE
	AppStatus  string `json:"appStatus"` // ACTIVE,SUSPEND
	EndTime    string `json:"endTime"`
	CreatorID  string `json:"creatorId"`
	CreateTime string `json:"createTime"`
	ModifierID string `json:"modifierId"`
	ModifyTime string `json:"modifyTime"`
	TenantID   string `json:"tenantId"`
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
