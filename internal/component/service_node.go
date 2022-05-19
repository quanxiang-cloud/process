package component

import (
	"context"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/rpc/pb"
	"gorm.io/gorm"
)

// ServiceNode non-artificial node
type ServiceNode struct {
	*Node
	TaskRepo models.TaskRepo
}

// Init init node
func (s *ServiceNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error {
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
	if initParam == nil || initParam.ExecuteType != PauseExecution {
		completeNodeReq := &CompleteNodeReq{
			Execution: req.Execution,
			Instance:  req.Instance,
			Node:      req.Node,
			NextNodes: req.NextNodes,
			UserID:    req.UserID,
			Params:    req.Params,
		}
		_, err := s.Complete(ctx, tx, completeNodeReq)
		return err
	}

	return nil
}

// Complete complete task
func (s *ServiceNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
	initNodeReq := &InitNodeReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      req.Node,
		NextNodes: req.NextNodes,
		UserID:    req.UserID,
		Params:    req.Params,
	}
	if _, err := s.InitNextNodes(ctx, tx, initNodeReq); err != nil {
		return false, err
	}

	// nothing
	return false, nil
}
