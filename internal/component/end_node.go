package component

import (
	"context"
	"github.com/quanxiang-cloud/process/internal"
	"github.com/quanxiang-cloud/process/internal/models"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	"gorm.io/gorm"
)

// EndNode EndNode
type EndNode struct {
	*Node
	TaskRepo        models.TaskRepo
	HistoryTaskRepo models.HistoryTaskRepo
	InstanceRepo    models.InstanceRepo
}

// Init init user component
func (n *EndNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq) error {
	cr := &CompleteNodeReq{
		Node:      req.Node,
		UserID:    req.UserID,
		Instance:  req.Instance,
		Execution: req.Execution,
	}
	_, err := n.Complete(ctx, tx, cr)
	return err
}

// Complete complete task
func (n *EndNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
	// --------------------add component instance begin--------------------------
	initNodeInstanceReq := InitNodeInstanceReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      req.Node,
		TaskID:    "",
		UserID:    req.UserID,
	}
	_, err := n.CreateNodeInstance(tx, &initNodeInstanceReq)
	if err != nil {
		return false, err
	}
	// update instance status
	instance := &models.Instance{
		ID:      req.Instance.ID,
		Status:  internal.Completed,
		EndTime: time2.Now(),
	}

	err = n.InstanceRepo.Update(tx, instance)
	if err != nil {
		return false, err
	}
	// 检查节点资源是否都已经销毁
	tks, err := n.TaskRepo.FindByInstanceID(tx, req.Instance.ID)
	if err != nil {
		return false, err
	}
	for i := 0; i < len(tks); i++ {
		if tks[i].TaskType != internal.NonModel {
			if err = n.TaskRepo.DeleteByID(tx, tks[i].ID); err != nil {
				return false, err
			}
			if err = n.TaskIdentity.DeleteByTaskID(tx, tks[i].ID); err != nil {
				return false, err
			}
		}
	}
	// if err = n.ExecutionRepo.DeleteByInstanceID(tx, req.Instance.ID); err != nil {
	// 	return false, err
	// }
	return true, nil
}
