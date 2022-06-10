package models

import "gorm.io/gorm"

// Node info
type Node struct {
	ID             string
	ProcID         string
	ProcInstanceID string
	Name           string
	DefKey         string
	NodeType       string // Start、End、User、MultiUser、Service、Script、ParallelGateway、InclusiveGateway、SubProcess
	SubProcID      string // Type is SubProcess
	PairDefKey     string // ParallelGateway < == > InclusiveGateway
	Desc           string
	CreatorID      string
	CreateTime     string
	ModifierID     string
	ModifyTime     string
	TenantID       string
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
