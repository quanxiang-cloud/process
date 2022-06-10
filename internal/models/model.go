package models

import (
	"gorm.io/gorm"
)

// Model info
type Model struct {
	ID         string
	Name       string
	DefKey     string
	CreatorID  string
	CreateTime string
	ModifierID string
	ModifyTime string
	TenantID   string
}

// ModelRepo interface
type ModelRepo interface {
	Create(db *gorm.DB, entity *Model) error
	FindByID(db *gorm.DB, id string) (*Model, error)
	FindByDefKey(db *gorm.DB, defKey string) ([]*Model, error)
}
