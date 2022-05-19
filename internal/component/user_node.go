package component

import (
	"context"
	"git.internal.yunify.com/qxp/misc/id2"
	"git.internal.yunify.com/qxp/misc/time2"
	"git.internal.yunify.com/qxp/process/internal"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/pkg"
	"git.internal.yunify.com/qxp/process/rpc/pb"
	"gorm.io/gorm"
	"strings"
)

// UserNode UserNode
type UserNode struct {
	*Node
	TaskRepo         models.TaskRepo
	VariablesRepo    models.VariablesRepo
	HistoryTaskRepo  models.HistoryTaskRepo
	IdentityLinkRepo models.IdentityLinkRepo
	InstanceRepo     models.InstanceRepo
}

// Init init user component
func (n *UserNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error {
	il, err := n.IdentityLinkRepo.QueryByNodeID(tx, req.Node.ID)
	if err != nil {
		return err
	}
	if req.NextNodes == "" {
		nextNodeStr, err := n.NodeLinkRepo.FindNextByNodeID(tx, req.Instance.ProcID, req.Node.ID)
		if err != nil {
			return err
		}
		req.NextNodes = nextNodeStr
	}
	taskType := internal.ModelTask
	if req.TaskType != "" {
		taskType = req.TaskType
	}
	t := &models.Task{
		ProcID:         req.Instance.ProcID,
		ProcInstanceID: req.Instance.ID,
		ExecutionID:    req.Execution.ID,
		NodeID:         req.Node.ID,
		NodeDefKey:     req.Node.DefKey,
		NextNodeDefKey: req.NextNodes,
		Name:           req.Node.Name,
		Desc:           req.Node.Desc,
		Assignee:       req.Assignee,
		TaskType:       taskType,
		Status:         internal.Active,
		CreatorID:      req.UserID,
	}
	if err = n.TaskRepo.Create(tx, t); err != nil {
		return err
	}
	req.InitResp.Tasks = t
	// 针对rpc消息通知
	req.InitResp.EventTaskID = t.ID
	// create task_identity
	if err = n.CreateTaskIdentity(ctx, tx, t, il, req.Instance.ID, req.Node.ID, req.UserID); err != nil {
		return err
	}
	// --------------------add component instance begin--------------------------
	initNodeInstanceReq := InitNodeInstanceReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      req.Node,
		TaskID:    t.ID,
		Assignee:  req.Assignee,
		UserID:    req.UserID,
	}
	nodeInstance, err := n.CreateNodeInstance(tx, &initNodeInstanceReq)
	if err != nil {
		return err
	}
	// --------------------add component instance end--------------------------

	// ---------------------update excution begin--------------------------
	entity := map[string]interface{}{
		"node_def_key":     req.Node.DefKey,
		"node_instance_id": nodeInstance.ID,
		"is_active":        1,
		"modifier_id":      req.UserID,
	}

	_, err = n.ExecutionRepo.Update(tx, req.Execution.ID, entity)
	if err != nil {
		return err
	}
	// ---------------------update excution end--------------------------

	return nil
}

// Complete complete task
func (n *UserNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
	t, err := n.TaskRepo.FindByID(tx, req.Task.ID)
	if err != nil {
		return false, err
	}

	ht := &models.HistoryTask{}

	err = pkg.CopyProperties(ht, t)
	if err != nil {
		return false, err
	}

	ht.Status = internal.Completed
	ht.Assignee = req.UserID
	ht.EndTime = time2.Now()
	ht.Comments = req.Comments

	err = n.TaskRepo.DeleteByID(tx, req.Task.ID)
	if err != nil {
		return false, err
	}
	// 更新instance的时间
	if err = n.InstanceRepo.Update(tx, req.Instance); err != nil {
		return false, err
	}
	err = n.HistoryTaskRepo.Create(tx, ht)
	if err != nil {
		return false, err
	}
	// if next node is gateway,set this node execution no active
	if req.Execution.PID != "" {
		err = n.ExecutionRepo.SetActive(tx, t.ExecutionID, 0)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// CreateTaskIdentity CreateTaskIdentity
func (n *UserNode) CreateTaskIdentity(ctx context.Context, tx *gorm.DB, t *models.Task, il []*models.IdentityLink, instanceID, nodeID, userID string) error {
	if t.Assignee != "" {
		if err := n.UserTaskIdentity(tx, t.ProcInstanceID, t.ID, t.Assignee, userID); err != nil {
			return err
		}
		return nil
	}
	for k := 0; k < len(il); k++ {
		switch il[k].IdentityType {
		case internal.IdentityUser:
			if err := n.UserTaskIdentity(tx, t.ProcInstanceID, t.ID, il[k].UserID, userID); err != nil {
				return err
			}
		case internal.IdentityGroup:
			if err := n.GroupTaskIdentity(tx, t.ProcInstanceID, t.ID, il[k].GroupID, userID); err != nil {
				return err
			}
		case internal.IdentityVariable:
			vv, err := n.getVariablesFromRedis(ctx, instanceID, nodeID, il[k].Variable)
			if err != nil {
				return err
			}
			if len(vv) == 0 {
				vv, err = n.VariablesRepo.GetStringArrayValue(tx, instanceID, nodeID, il[k].Variable)
				if err != nil {
					return err
				}
			}
			for _, v := range vv {
				if strings.Contains(v, internal.Dep) {
					if err = n.GroupTaskIdentity(tx, t.ProcInstanceID, t.ID, v, userID); err != nil {
						return err
					}
				} else {
					if err = n.UserTaskIdentity(tx, t.ProcInstanceID, t.ID, v, userID); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// GroupTaskIdentity GroupTaskIdentity
func (n *UserNode) GroupTaskIdentity(tx *gorm.DB, instanceID, taskID, groupID, createID string) error {
	if groupID == "" {
		return nil
	}
	ti := &models.TaskIdentity{
		ID:           id2.GenID(),
		TaskID:       taskID,
		InstanceID:   instanceID,
		CreateTime:   time2.Now(),
		CreatorID:    createID,
		GroupID:      internal.Dep + "_" + groupID,
		IdentityType: internal.IdentityGroup,
	}
	if err := n.TaskIdentity.Create(tx, ti); err != nil {
		return err
	}
	return nil
}
