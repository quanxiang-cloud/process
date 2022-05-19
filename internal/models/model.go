package models

import (
	"gorm.io/gorm"
)

// Model info
type Model struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	DefKey     string `json:"defKey"`
	CreatorID  string `json:"creatorId"`
	CreateTime string `json:"createTime"`
	ModifierID string `json:"modifierId"`
	ModifyTime string `json:"modifyTime"`
	TenantID   string `json:"tenantId"`
}

// ModelRepo interface
type ModelRepo interface {
	Create(db *gorm.DB, entity *Model) error
	FindByID(db *gorm.DB, id string) (*Model, error)
	FindByDefKey(db *gorm.DB, defKey string) ([]*Model, error)
}
