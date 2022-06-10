package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Variables entity
type Variables struct {
	ID             string
	ProcInstanceID string
	NodeID         string
	VarScope       string // LOCAL
	Name           string
	VarType        string // string, []string
	Value          string
	BytesValue     []byte
	ComplexValue   datatypes.JSON
	CreatorID      string
	CreateTime     string
	ModifierID     string
	ModifyTime     string
	TenantID       string
}

// VariablesRepo interface
type VariablesRepo interface {
	Create(db *gorm.DB, entity *Variables) error
	Update(db *gorm.DB, entity *Variables) error
	GetStringValue(db *gorm.DB, instanceID string, name string) (string, error)
	GetStringArrayValue(db *gorm.DB, instanceID, nodeID, name string) ([]string, error)
	GetInstanceValue(db *gorm.DB, instanceID string) (map[string]interface{}, error)
	GetInstanceValueByName(db *gorm.DB, instanceID string, names []string) (map[string]interface{}, error)
	FindVariablesByName(db *gorm.DB, instanceID, nodeID, name string) (*Variables, error)
}
