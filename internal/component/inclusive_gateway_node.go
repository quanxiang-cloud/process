package component

import (
	"context"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/rpc/pb"
	"gorm.io/gorm"
)

// InclusiveGatewayNode confluence gateway node
type InclusiveGatewayNode struct {
	*Node
	TaskRepo models.TaskRepo
}

// Init init node
func (inc *InclusiveGatewayNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error {
	// 设置进入合流网关的分支
	if err := inc.ExecutionRepo.SetActive(tx, req.Execution.ID, 0); err != nil {
		return err
	}
	nodes := make([]*models.Execution, 0)
	err := inc.findActiveExecution(tx, req.Instance.ID, req.Execution.PID, &nodes)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		entity := map[string]interface{}{
			"node_def_key":     req.Node.DefKey,
			"node_instance_id": req.Execution.NodeInstanceID,
			"is_active":        1,
			"modifier_id":      req.UserID,
		}
		// 把父executionID激活
		mExecution, err := inc.ExecutionRepo.Update(tx, req.Execution.PID, entity)
		if err != nil {
			return err
		}
		creq := &CompleteNodeReq{
			Execution: mExecution,
			Instance:  req.Instance,
			Node:      req.Node,
			UserID:    req.UserID,
		}
		_, err = inc.Complete(ctx, tx, creq)
		if err != nil {
			return err
		}
	}
	return nil
}

// Complete complete task
func (inc *InclusiveGatewayNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
	// create node instance
	initNodeInstanceReq := InitNodeInstanceReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      req.Node,
		TaskID:    "",
		UserID:    req.UserID,
	}
	ni, err := inc.CreateNodeInstance(tx, &initNodeInstanceReq)
	if err != nil {
		return false, err
	}
	req.Execution.NodeInstanceID = ni.ID
	initReq := &InitNodeReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      req.Node,
		UserID:    req.UserID,
	}
	if _, err := inc.InitNextNodes(ctx, tx, initReq); err != nil {
		return false, err
	}
	return true, nil
}

func (inc *InclusiveGatewayNode) findActiveExecution(tx *gorm.DB, instanceID, executionID string, res *[]*models.Execution) error {
	exps, err := inc.ExecutionRepo.FindByPID(tx, instanceID, executionID, 1)
	if err != nil {
		return err
	}
	if len(exps) == 0 {
		// 还需要递归查找同级分支的子分支是否有未完成的
		childEx := make([]*models.Execution, 0)
		expsa, err := inc.ExecutionRepo.FindAllByPID(tx, instanceID, executionID)
		if err != nil {
			return err
		}
		for i := 0; i < len(expsa); i++ {
			err = inc.findActiveExecution(tx, expsa[i].ProcInstanceID, expsa[i].ID, &childEx)
			if err != nil {
				return err
			}
		}
		exps = append(exps, childEx...)
	}
	*res = append(*res, exps...)
	return nil
}
