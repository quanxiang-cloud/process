package models

import "gorm.io/gorm"

// IdentityLink info
type IdentityLink struct {
	ID           string `json:"id"`
	NodeID       string `json:"nodeId"`
	UserID       string `json:"userId"`
	GroupID      string `json:"groupId"`
	Variable     string `json:"variable"`
	IdentityType string `json:"identityType"` // USER、VARIABLE,GROUP
	CreatorID    string `json:"creatorId"`
	CreateTime   string `json:"createTime"`
	ModifierID   string `json:"modifierId"`
	ModifyTime   string `json:"modifyTime"`
	TenantID     string `json:"tenantId"`
}

// IdentityLinkRepo interface
type IdentityLinkRepo interface {
	Create(db *gorm.DB, entity *IdentityLink) error
	QueryByNodeID(db *gorm.DB, nodeID string) ([]*IdentityLink, error)
}
