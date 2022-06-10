package component

import (
	"context"
	"github.com/quanxiang-cloud/process/internal/models"
	"gorm.io/gorm"
)

// ServiceNode non-artificial node
type ServiceNode struct {
	*Node
	TaskRepo models.TaskRepo
}

// Init init node
func (s *ServiceNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq) error {
	// init node instance
	initNodeInstanceReq := InitNodeInstanceReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      req.Node,
		TaskID:    "",
		UserID:    req.UserID,
	}
	if _, err := s.CreateNodeInstance(tx, &initNodeInstanceReq); err != nil {
		return err
	}
	if err := s.InitNextNodes(ctx, tx, req); err != nil {
		return err
	}
	return nil
}

// Complete complete task
func (s *ServiceNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
	// nothing
	return false, nil
}
