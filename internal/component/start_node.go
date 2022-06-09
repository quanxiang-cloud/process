package component

import (
	"context"
	"github.com/quanxiang-cloud/process/rpc/pb"
	"gorm.io/gorm"
)

// StartNode StartNode
type StartNode struct {
	*Node
}

// Init init node
func (n *StartNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error {
	cr := &CompleteNodeReq{
		Node:      req.Node,
		UserID:    req.UserID,
		Instance:  req.Instance,
		Execution: req.Execution,
	}
	_, err := n.Complete(ctx, tx, cr)
	return err
}

// Complete complete node
func (n *StartNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
	// --------------------add component instance begin--------------------------
	initNodeInstanceReq := InitNodeInstanceReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      req.Node,
		TaskID:    "",
		UserID:    req.UserID,
	}
	nodeInstance, err := n.CreateNodeInstance(tx, &initNodeInstanceReq)
	if err != nil {
		return false, err
	}
	req.Execution.NodeInstanceID = nodeInstance.ID
	// --------------------add component instance end--------------------------

	// ---------------------update excution begin--------------------------
	req.Execution, err = n.UpdateExecution(tx, req, nodeInstance.ID)
	if err != nil {
		return false, err
	}
	// ---------------------update excution end--------------------------

	return true, nil
}
