package component

import (
	"context"
	"github.com/quanxiang-cloud/process/internal"
	"github.com/quanxiang-cloud/process/internal/models"
	"github.com/quanxiang-cloud/process/pkg"
	"github.com/quanxiang-cloud/process/pkg/client"
	"github.com/quanxiang-cloud/process/pkg/config"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	"github.com/quanxiang-cloud/process/rpc/pb"
	"gorm.io/gorm"
	"strings"
)

// MultiUserNode MultiUserNode
type MultiUserNode struct {
	*Node
	HistoryTaskRepo models.HistoryTaskRepo
	Identity        client.Identity
	Conf            *config.Configs
	UserNode
}

// Init init node
func (n *MultiUserNode) Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error {
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
		return err
	}
	// --------------------add component instance end--------------------------

	// ---------------------update excution begin--------------------------
	entity := map[string]interface{}{
		"node_def_key":     req.Node.DefKey,
		"node_instance_id": nodeInstance.ID,
		"is_active":        0,
		"modifier_id":      req.UserID,
	}

	req.Execution, err = n.ExecutionRepo.Update(tx, req.Execution.ID, entity)
	if err != nil {
		return err
	}
	// ---------------------update excution end--------------------------
	userIDs, err := n.packUserID(ctx, tx, req)
	if err != nil {
		return err
	}

	ids := removeRepeatedElement(userIDs)
	if len(ids) == 0 {
		initNodeReq := &InitNodeReq{
			Execution: req.Execution,
			Assignee:  "",
		}
		if err := n.createSingleNode(ctx, tx, req, initNodeReq); err != nil {
			return err
		}
		return nil
	}
	if err = n.createNode(ctx, tx, req, ids, nodeInstance.ID); err != nil {
		return err
	}
	return nil
}

// Complete complete task
func (n *MultiUserNode) Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error) {
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

	err = n.HistoryTaskRepo.Create(tx, ht)
	if err != nil {
		return false, err
	}

	err = n.ExecutionRepo.SetActive(tx, t.ExecutionID, 0)
	if err != nil {
		return false, err
	}
	err = n.TaskRepo.DeleteByID(tx, req.Task.ID)
	if err != nil {
		return false, err
	}
	// 更新instance的时间
	if err = n.InstanceRepo.Update(tx, req.Instance); err != nil {
		return false, err
	}
	nodes, err := n.ExecutionRepo.FindByPID(tx, req.Instance.ID, req.Execution.PID, 1)
	if err != nil {
		return false, err
	}
	if len(nodes) == 0 {
		err = n.ExecutionRepo.SetActive(tx, req.Execution.PID, 1)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	// 所有会签都完成了，才能算成功，可以初始化下个节点
	return false, err
}

func (n *MultiUserNode) createSingleNode(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initNodeReq *InitNodeReq) error {
	initNodeReq.Instance = req.Instance
	initNodeReq.Node = req.Node
	initNodeReq.UserID = req.UserID
	initNodeReq.NextNodes = req.NextNodes
	initNodeReq.TaskType = req.TaskType
	err := n.UserNode.Init(ctx, tx, initNodeReq, nil)
	if err != nil {
		return err
	}
	return nil
}

func (n *MultiUserNode) createNode(ctx context.Context, tx *gorm.DB, req *InitNodeReq, userIDs []string, nodeInstanceID string) error {
	for i := 0; i < len(userIDs); i++ {
		// ---------------------create excution begin--------------------------
		nExecution := &models.Execution{
			ProcID:         req.Instance.ProcID,
			NodeInstanceID: nodeInstanceID,
			ProcInstanceID: req.Instance.ID,
			PID:            req.Execution.ID,
			IsActive:       1,
			CreatorID:      req.UserID,
		}
		err := n.ExecutionRepo.Create(tx, nExecution)
		if err != nil {
			return err
		}
		// ---------------------create excution end--------------------------
		// --------------------component begin-------------------
		initNodeReq := &InitNodeReq{
			Execution: nExecution,
			Assignee:  userIDs[i],
		}
		if err := n.createSingleNode(ctx, tx, req, initNodeReq); err != nil {
			return err
		}
		req.InitResp.Tasks = initNodeReq.InitResp.Tasks
		if req.InitResp.EventTaskID == "" {
			req.InitResp.EventTaskID = initNodeReq.InitResp.Tasks.ID
		} else {
			req.InitResp.EventTaskID = strings.Join([]string{req.InitResp.EventTaskID, initNodeReq.InitResp.Tasks.ID}, ",")
		}
		// --------------------component end-------------------
	}
	return nil
}

func (n *MultiUserNode) packGroup(ctx context.Context, groupIDs []string, userIDs *[]string) error {
	if len(groupIDs) > 0 {
		gUserIDs, err := n.Identity.FindUserIDsByGroups(ctx, groupIDs)
		if err != nil {
			return err
		}
		*userIDs = append(*userIDs, gUserIDs...)
	}
	return nil
}

func (n *MultiUserNode) queryVariable(ctx context.Context, tx *gorm.DB, req *InitNodeReq, variables string) ([]string, error) {
	vv, err := n.getVariablesFromRedis(ctx, req.Instance.ID, req.Node.ID, variables)
	if err != nil {
		return nil, err
	}
	if vv == nil {
		vv, err = n.VariablesRepo.GetStringArrayValue(tx, req.Instance.ID, req.Node.ID, variables)
		if err != nil {
			return nil, err
		}
	}
	return vv, nil
}

func (n *MultiUserNode) packVariable(ctx context.Context, tx *gorm.DB, req *InitNodeReq, variables string, userIDs *[]string) error {
	if variables == "" {
		return nil
	}
	vv, err := n.queryVariable(ctx, tx, req, variables)
	if err != nil {
		return err
	}
	for _, v := range vv {
		if strings.Contains(v, internal.Dep) {
			users, err := n.Identity.FindUsersByGroup(ctx, v[4:])
			if err != nil {
				return err
			}
			for _, us := range users.Users {
				*userIDs = append(*userIDs, us.ID)
			}
		}
		*userIDs = append(*userIDs, v)
	}
	return nil
}

func (n *MultiUserNode) packUserID(ctx context.Context, tx *gorm.DB, req *InitNodeReq) ([]string, error) {
	idens, err := n.IdentityLinkRepo.QueryByNodeID(tx, req.Node.ID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0)
	groupIDs := make([]string, 0)
	var variables string
	for i := 0; i < len(idens); i++ {
		switch idens[i].IdentityType {
		case internal.IdentityUser:
			if idens[i].UserID != "" {
				userIDs = append(userIDs, idens[i].UserID)
			}
		case internal.IdentityGroup:
			if idens[i].GroupID != "" {
				groupIDs = append(groupIDs, idens[i].GroupID)
			}
		case internal.IdentityVariable:
			variables = idens[i].Variable
		}
	}
	// get group users
	if err := n.packGroup(ctx, groupIDs, &userIDs); err != nil {
		return nil, err
	}
	// get variable users,if is groupID,should get group users
	if err := n.packVariable(ctx, tx, req, variables, &userIDs); err != nil {
		return nil, err
	}
	return userIDs, nil
}

func removeRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}
