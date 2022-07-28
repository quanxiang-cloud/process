package component

import (
	"context"
	"github.com/quanxiang-cloud/process/internal"
	"github.com/quanxiang-cloud/process/internal/models"
	"github.com/quanxiang-cloud/process/pkg"
	"github.com/quanxiang-cloud/process/pkg/client"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	"github.com/quanxiang-cloud/process/rpc/pb"
	"gorm.io/gorm"
	"strings"
)

// NonModelNode temp node,read show ,copy for
type NonModelNode struct {
	*Node
	Identity        client.Identity
	TaskRepo        models.TaskRepo
	HistoryTaskRepo models.HistoryTaskRepo
	InstanceRepo    models.InstanceRepo
}

// Init init user component
func (n *NonModelNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error {
	if strings.Contains(req.Assignee, internal.Dep) {
		users, err := n.Identity.FindUsersByGroup(ctx, req.Assignee[4:])
		if err != nil {
			return err
		}
		for _, us := range users.Users {
			req.Assignee = us.ID
			if err = n.initSingleUser(tx, req); err != nil {
				return err
			}
		}
		return nil
	}
	if err := n.initSingleUser(tx, req); err != nil {
		return err
	}
	return nil
}

func (n *NonModelNode) initSingleUser(tx *gorm.DB, req *InitNodeReq) error {
	t := genTask(req)
	req.InitResp.Tasks = t
	if err := n.TaskRepo.Create(tx, t); err != nil {
		return err
	}
	// 这种节点是否创建nodeInstance
	initNodeInstanceReq := InitNodeInstanceReq{
		Execution: req.Execution,
		Instance:  req.Instance,
		Node:      &models.Node{},
		TaskID:    t.ID,
		Assignee:  req.Assignee,
		UserID:    req.UserID,
	}
	_, err := n.CreateNodeInstance(tx, &initNodeInstanceReq)
	if err != nil {
		return err
	}
	if err := n.UserTaskIdentity(tx, t.ProcInstanceID, t.ID, t.Assignee, req.UserID); err != nil {
		return err
	}
	return nil
}

// Complete complete task
func (n *NonModelNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
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
	inst, err := n.InstanceRepo.FindByID(tx, t.ProcInstanceID)
	if err != nil {
		return false, err
	}
	// 更新instance的时间
	if err = n.InstanceRepo.Update(tx, inst); err != nil {
		return false, err
	}

	err = n.HistoryTaskRepo.Create(tx, ht)
	if err != nil {
		return false, err
	}
	// need not init next node,so return false
	return false, nil
}

func genTask(req *InitNodeReq) *models.Task {
	t := &models.Task{
		ProcID:         req.Instance.ProcID,
		ProcInstanceID: req.Instance.ID,
		ExecutionID:    req.Execution.ID,
		NodeDefKey:     req.NextNodes,
		Name:           req.Name,
		Desc:           req.Desc,
		Assignee:       req.Assignee,
		TaskType:       internal.NonModel,
		Status:         internal.Active,
		CreatorID:      req.UserID,
	}
	return t
}
