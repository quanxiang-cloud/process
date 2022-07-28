package models

import "gorm.io/gorm"

// Node info
type Node struct {
	ID             string `json:"id"`
	ProcID         string `json:"procId"`
	ProcInstanceID string `json:"procInstanceId"`
	Name           string `json:"name"`
	DefKey         string `json:"defKey"`
	NodeType       string `json:"nodeType"`   // Start、End、User、MultiUser、Service、Script、ParallelGateway、InclusiveGateway、SubProcess
	SubProcID      string `json:"subProcId"`  // Type is SubProcess
	PairDefKey     string `json:"pairDefKey"` // ParallelGateway < == > InclusiveGateway
	Desc           string `json:"desc"`
	CreatorID      string `json:"creatorId"`
	CreateTime     string `json:"createTime"`
	ModifierID     string `json:"modifierId"`
	ModifyTime     string `json:"modifyTime"`
	TenantID       string `json:"tenantId"`
}

// NodeRepo interface
type NodeRepo interface {
	Create(db *gorm.DB, entity *Node) error
	FindStartNode(db *gorm.DB, processID string) (*Node, error)
	FindByID(db *gorm.DB, id string) (*Node, error)
	FindByProcessID(db *gorm.DB, processID, nodeDefKey string) ([]*Node, error)
	FindByDefKey(db *gorm.DB, processID, defKey string) (*Node, error)
	FindInstanceNode(db *gorm.DB, instanceID, defKey string) (*Node, error)
}
