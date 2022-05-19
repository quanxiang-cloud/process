package component

import (
	"context"
	"git.internal.yunify.com/qxp/process/internal"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/rpc/pb"
	"gorm.io/gorm"
)

// FreeNode FreeNode,后加签、跳转结束...
type FreeNode struct {
	*Node
	TaskRepo         models.TaskRepo
	VariablesRepo    models.VariablesRepo
	HistoryTaskRepo  models.HistoryTaskRepo
	IdentityLinkRepo models.IdentityLinkRepo
}

// Init init free component
func (n *FreeNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error {
	currentNode, err := n.NodeRepo.FindByDefKey(tx, req.Task.ProcID, req.NextNodes)
	if err != nil {
		return err
	}
	req.NextNodes = req.Task.NextNodeDefKey
	req.Node = currentNode
	req.TaskType = internal.TempModel
	cNode, err := NodeFactory(currentNode.NodeType)
	if err != nil {
		return err
	}
	// 会签节点、分支节点，需要特殊处理 todo
	if err = cNode.Init(ctx, tx, req, nil); err != nil {
		return err
	}
	return nil
}

// Complete complete task
func (n *FreeNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
	return false, nil
}
