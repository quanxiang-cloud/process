package component

import (
	"context"
	"git.internal.yunify.com/qxp/misc/error2"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/pkg/client"
	"git.internal.yunify.com/qxp/process/pkg/code"
	"git.internal.yunify.com/qxp/process/rpc/pb"
	"gorm.io/gorm"
)

const defaultBranch = "defaultBranch"

// ParallelGatewayNode shunt gateway node
type ParallelGatewayNode struct {
	*Node
	Condition        client.Condition
	TaskRepo         models.TaskRepo
	IdentityLinkRepo models.IdentityLinkRepo
	VariablesRepo    models.VariablesRepo
}

// Init init node
func (p *ParallelGatewayNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error {
	nextNodeDefKeys, err := p.NodeLinkRepo.FindNextNodesByNodeID(tx, req.Instance.ProcID, req.Node.ID)
	if err != nil {
		return err
	}
	// init node instance
	initNodeInstanceReq := InitNodeInstanceReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      req.Node,
		TaskID:    "",
		UserID:    req.UserID,
	}
	nodeInstance, err := p.CreateNodeInstance(tx, &initNodeInstanceReq)
	if err != nil {
		return err
	}
	// check next node condition
	flag := false
	defaultNode := &models.Node{}
	for i := 0; i < len(nextNodeDefKeys); i++ {
		node, err := p.NodeRepo.FindByDefKey(tx, req.Instance.ProcID, nextNodeDefKeys[i])
		if err != nil {
			return err
		}
		nls, err := p.NodeLinkRepo.FindByNodeDefKey(tx, node.ProcID, node.DefKey)
		if err != nil {
			return err
		}
		if len(nls) > 0 {
			for _, nl := range nls {
				// 参数封装 todo
				vv, err := p.VariablesRepo.GetInstanceValue(tx, req.Instance.ID)
				if err != nil {
					return err
				}
				param := req.Params
				for k, v := range vv {
					param[k] = v
				}
				b := false
				if nl.Condition == defaultBranch {
					defaultNode = node
					continue
				}
				if nl.Condition != "" {
					b, err = p.Condition.GetConditionResult(ctx, nl.Condition, param)
					if err != nil {
						return err
					}
				}
				if b {
					if err := p.createBranch(ctx, tx, req, nodeInstance, node); err != nil {
						return err
					}
					flag = true
				}
			}
		}
	}
	// 所有节点的condition都没有满足,抛异常
	if !flag {
		if defaultNode.NodeType == "" {
			return error2.NewError(code.AllConditionMissMatch)
		}
		if err := p.createBranch(ctx, tx, req, nodeInstance, defaultNode); err != nil {
			return err
		}
	}
	return nil
}

// Complete complete task
func (p *ParallelGatewayNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
	// nothing
	return false, nil
}

func (p *ParallelGatewayNode) createBranch(ctx context.Context, tx *gorm.DB, req *InitNodeReq,
	nodeInstance *models.NodeInstance, node *models.Node) error {
	// init execution
	execution := &models.Execution{
		ProcID:         req.Execution.ProcID,
		ProcInstanceID: req.Execution.ProcInstanceID,
		NodeInstanceID: nodeInstance.ID,
		PID:            req.Execution.ID,
		IsActive:       1,
		CreatorID:      req.UserID,
	}
	err := p.ExecutionRepo.Create(tx, execution)
	if err != nil {
		return err
	}
	// update main process execution
	entity := map[string]interface{}{
		"node_def_key":     req.Node.DefKey,
		"node_instance_id": nodeInstance.ID,
		"is_active":        0,
		"modifier_id":      req.UserID,
	}
	_, err = p.ExecutionRepo.Update(tx, req.Execution.ID, entity)
	if err != nil {
		return err
	}

	initNodeReq := &InitNodeReq{
		Execution: execution,
		Instance:  req.Instance,
		Node:      node,
		NextNodes: node.DefKey,
		UserID:    req.UserID,
		Params:    req.Params,
	}
	// err = cNode.Init(ctx, tx, initNodeReq)
	_, err = p.InitNextNodes(ctx, tx, initNodeReq)
	if err != nil {
		return err
	}
	return nil
}
