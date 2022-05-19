package models

import "gorm.io/gorm"

// NodeLink info
type NodeLink struct {
	ID             string `json:"id"`
	ProcID         string `json:"procId"`
	NodeID         string `json:"nodeId"`
	NextNodeDefKey string `json:"nextNodeDefKey"` // next component def key
	Condition      string `json:"condition"`
	CreatorID      string `json:"creatorId"`
	CreateTime     string `json:"createTime"`
	ModifierID     string `json:"modifierId"`
	ModifyTime     string `json:"modifyTime"`
	TenantID       string `json:"tenantId"`
}

// NodeLinkRepo interface
type NodeLinkRepo interface {
	Create(db *gorm.DB, entity *NodeLink) error
	FindByNodeID(db *gorm.DB, processID, nodeID string) ([]*NodeLink, error)
	FindByNodeDefKey(db *gorm.DB, processID, nodeDefKey string) ([]*NodeLink, error)
	FindNextByNodeID(db *gorm.DB, processID, nodeID string) (string, error)
	FindNextNodesByNodeID(db *gorm.DB, processID, nodeID string) ([]string, error)
}
