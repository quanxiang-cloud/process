package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Variables entity
type Variables struct {
	ID             string         `json:"id"`
	ProcInstanceID string         `json:"procInstanceId"`
	NodeID         string         `json:"nodeId"`
	VarScope       string         `json:"varScope"` // LOCAL
	Name           string         `json:"name"`
	VarType        string         `json:"varType"` // string, []string
	Value          string         `json:"value"`
	BytesValue     []byte         `json:"bytesValue"`
	ComplexValue   datatypes.JSON `json:"complexValue"`
	CreatorID      string         `json:"creatorId"`
	CreateTime     string         `json:"createTime"`
	ModifierID     string         `json:"modifierId"`
	ModifyTime     string         `json:"modifyTime"`
	TenantID       string         `json:"tenantId"`
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

	GetInstanceValueAndParams(db *gorm.DB, instanceID string, params map[string]interface{}) (map[string]interface{}, error)
}
