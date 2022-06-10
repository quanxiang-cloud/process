package models

import "gorm.io/gorm"

// IdentityLink info
type IdentityLink struct {
	ID           string
	NodeID       string
	UserID       string
	GroupID      string
	Variable     string
	IdentityType string // USER、VARIABLE,GROUP
	CreatorID    string
	CreateTime   string
	ModifierID   string
	ModifyTime   string
	TenantID     string
}

// IdentityLinkRepo interface
type IdentityLinkRepo interface {
	Create(db *gorm.DB, entity *IdentityLink) error
	QueryByNodeID(db *gorm.DB, nodeID string) ([]*IdentityLink, error)
}
