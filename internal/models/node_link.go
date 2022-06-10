package models

import "gorm.io/gorm"

// NodeLink info
type NodeLink struct {
	ID             string
	ProcID         string
	NodeID         string
	NextNodeDefKey string // next component def key
	Condition      string
	CreatorID      string
	CreateTime     string
	ModifierID     string
	ModifyTime     string
	TenantID       string
}

// NodeLinkRepo interface
type NodeLinkRepo interface {
	Create(db *gorm.DB, entity *NodeLink) error
	FindByNodeID(db *gorm.DB, processID, nodeID string) ([]*NodeLink, error)
	FindByNodeDefKey(db *gorm.DB, processID, nodeDefKey string) ([]*NodeLink, error)
	FindNextByNodeID(db *gorm.DB, processID, nodeID string) (string, error)
	FindNextNodesByNodeID(db *gorm.DB, processID, nodeID string) ([]string, error)
}
