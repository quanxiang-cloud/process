package process

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs"
	rc "github.com/go-redis/redis/v8"
	"github.com/quanxiang-cloud/process/internal"
	"github.com/quanxiang-cloud/process/internal/component"
	"github.com/quanxiang-cloud/process/internal/models"
	"github.com/quanxiang-cloud/process/internal/models/mysql"
	listener "github.com/quanxiang-cloud/process/internal/server/events"
	"github.com/quanxiang-cloud/process/internal/server/options"
	"github.com/quanxiang-cloud/process/pkg"
	"github.com/quanxiang-cloud/process/pkg/client"
	"github.com/quanxiang-cloud/process/pkg/code"
	"github.com/quanxiang-cloud/process/pkg/config"
	"github.com/quanxiang-cloud/process/pkg/misc/error2"
	"github.com/quanxiang-cloud/process/pkg/misc/id2"
	"github.com/quanxiang-cloud/process/pkg/misc/logger"
	"github.com/quanxiang-cloud/process/pkg/misc/redis2"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	"github.com/quanxiang-cloud/process/pkg/page"
	"github.com/quanxiang-cloud/process/rpc/pb"
	"gorm.io/gorm"
	"strings"
)

const (
	identityUser     = "USER"
	identityVariable = "VARIABLE"
	user             = "User"
	multiUser        = "MultiUser"
	service          = "Service"
	inclusiveGateway = "InclusiveGateway"
	parallelGateway  = "ParallelGateway"
	start            = "Start"
	end              = "End"
	free             = "Free"
	nonModel         = "NonModel"
)

// Task task
type Task interface {
	AgencyTask(ctx context.Context, req *QueryTaskReq) (*page.Page, error)
	AgencyTaskTotal(ctx context.Context, req *QueryTaskReq) (*TotalTaskResp, error)
	WholeTask(ctx context.Context, req *QueryTaskReq) (*page.Page, error)
	DeleteTask(ctx context.Context, req *DeleteTaskReq) (*DeleteTaskResp, error)
	CompleteTask(ctx context.Context, req *CompleteTaskReq) (*CompleteTaskResp, error)
	BatchCompleteNonModelTask(ctx context.Context, req *CompleteNonModelTaskReq) (*CompleteTaskResp, error)
	CompleteExecution(ctx context.Context, req *CompleteExecutionReq) (*CompleteTaskResp, error)
	InitTask(ctx context.Context, tx *gorm.DB, req *component.CompleteNodeReq) (*pb.NodeEventRespData, error)
	InstanceTasks(ctx context.Context, req *InstanceTaskReq) ([]*models.Task, error)
	InstanceDoneTasks(ctx context.Context, req *QueryTaskReq) (*page.Page, error)
	AddNonNodeTask(ctx context.Context, req *AddTaskReq) (*AddTaskResp, error)
	AddNodeTask(ctx context.Context, req *AddTaskReq) (*AddTaskResp, error)
	AddBackNodeTask(ctx context.Context, req *AddTaskReq) (*AddTaskResp, error)
	BackReFill(ctx context.Context, req *AddTaskReq) (*AddTaskResp, error)
	FallbackTask(ctx context.Context, req *AddTaskReq) (*AddTaskResp, error)
	AddTaskCondition(ctx context.Context, req *AddTaskConditionReq) (*AddTaskResp, error)
	TransferTask(ctx context.Context, req *AddTaskConditionReq) (*AddTaskResp, error)
	AddHistoryTask(ctx context.Context, req *AddHistoryTaskReq) (*AddHistoryTaskResp, error)
	FindTaskPreNode(ctx context.Context, req *TaskPreNodeReq) (*TaskPreNodeResp, error)
	FindGateWayExecutions(ctx context.Context, req *GateWayExecutionReq) (*GateWayExecutionResp, error)
	FindParentExecutions(ctx context.Context, req *ParentExecutionReq) (*ParentExecutionResp, error)
	// inter function
	DoDeleteTask(tx *gorm.DB, task *models.Task) error
	QueryGroupIDByUserID(ctx context.Context, userID string) (string, error)
	R() *rc.ClusterClient
}

type task struct {
	db               *gorm.DB
	l                *listener.Listener
	node             Node
	conf             *config.Configs
	taskRepo         models.TaskRepo
	identity         client.Identity
	nodeRepo         models.NodeRepo
	nodeLinkRepo     models.NodeLinkRepo
	taskIdentity     models.TaskIdentityRepo
	instantRepo      models.InstanceRepo
	identityLinkRepo models.IdentityLinkRepo
	variablesRepo    models.VariablesRepo
	nodeInstanceRepo models.NodeInstanceRepo
	executionRepo    models.ExecutionRepo
	historyTaskRepo  models.HistoryTaskRepo
	redisClient      *rc.ClusterClient
}

// NewTask new
func NewTask(conf *config.Configs, opts ...options.Options) (Task, error) {
	identity, err := client.NewIdentity(conf)
	if err != nil {
		return nil, err
	}
	redisClient, err := redis2.NewClient(conf.Redis)
	if err != nil {
		return nil, err
	}
	node, _ := NewNode(conf, opts...)
	task := &task{
		conf:             conf,
		identity:         identity,
		node:             node,
		taskRepo:         mysql.NewTaskRepo(),
		nodeRepo:         mysql.NewNodeRepo(),
		nodeLinkRepo:     mysql.NewNodeLinkRepo(),
		instantRepo:      mysql.NewInstanceRepo(),
		taskIdentity:     mysql.NewTaskIdentityRepo(),
		identityLinkRepo: mysql.NewIdentityLinkRepo(),
		variablesRepo:    mysql.NewVariablesRepo(),
		nodeInstanceRepo: mysql.NewNodeInstanceRepo(),
		executionRepo:    mysql.NewExecutionRepo(),
		historyTaskRepo:  mysql.NewHistoryTaskRepo(),
		redisClient:      redisClient,
	}

	for _, opt := range opts {
		opt(task)
	}
	err = InitNodeHandler(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// InitNodeHandler InitNodeHandler
func InitNodeHandler(task *task) error {
	node := &component.Node{
		Listener:         task.l,
		TaskIdentity:     task.taskIdentity,
		NodeInstanceRepo: task.nodeInstanceRepo,
		ExecutionRepo:    task.executionRepo,
		NodeLinkRepo:     task.nodeLinkRepo,
		NodeRepo:         task.nodeRepo,
		RedisClient:      task.redisClient,
	}
	nodePack(withStart(node), withMultiUser(task, node), withEnd(task, node), withNonModel(task, node),
		withFree(task, node), withService(task, node), withParallelGateway(task, node), withInclusiveGateway(task, node))
	return nil
}

func (t *task) AddTaskCondition(ctx context.Context, req *AddTaskConditionReq) (*AddTaskResp, error) {
	task := &models.Task{
		ID:         req.TaskID,
		DueTime:    req.DueTime,
		ModifierID: req.UserID,
		ModifyTime: time2.Now(),
	}
	if err := t.taskRepo.Update(t.db, task); err != nil {
		return nil, err
	}
	return &AddTaskResp{}, nil
}

func (t *task) userTransfer(tx *gorm.DB, req *AddTaskConditionReq, ti *models.TaskIdentity) error {
	if ti == nil {
		tn := &models.TaskIdentity{
			ID:           id2.GenID(),
			TaskID:       req.TaskID,
			InstanceID:   req.InstanceID,
			UserID:       req.Assignee,
			IdentityType: internal.IdentityUser,
			CreatorID:    req.UserID,
			CreateTime:   time2.Now(),
		}
		if err := t.taskIdentity.Create(tx, tn); err != nil {
			return err
		}
		return nil
	}
	ti.UserID = req.Assignee
	ti.GroupID = ""
	ti.ModifierID = req.UserID
	ti.ModifyTime = time2.Now()
	if err := t.taskIdentity.Update(tx, ti); err != nil {
		return err
	}
	return nil
}

func (t *task) nonUserTransfer(tx *gorm.DB, req *AddTaskConditionReq) error {
	if err := t.taskIdentity.DeleteByTaskID(tx, req.TaskID); err != nil {
		return err
	}
	tn := &models.TaskIdentity{
		ID:           id2.GenID(),
		TaskID:       req.TaskID,
		InstanceID:   req.InstanceID,
		UserID:       req.Assignee,
		IdentityType: internal.IdentityUser,
		CreatorID:    req.UserID,
		CreateTime:   time2.Now(),
	}
	if err := t.taskIdentity.Create(tx, tn); err != nil {
		return err
	}
	return nil
}

func (t *task) TransferTask(ctx context.Context, req *AddTaskConditionReq) (resp *AddTaskResp, err error) {
	tx := t.db.Begin()
	defer func() {
		if err != nil {
			logger.Logger.Errorw(err.Error(), logger.STDRequestID(ctx))
			tx.Rollback()
		}
	}()
	resp = &AddTaskResp{}
	depID, err := t.QueryGroupIDByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	tk, err := t.taskRepo.FindByID(tx, req.TaskID)
	if err != nil {
		return nil, err
	}
	if tk == nil {
		return nil, error2.NewError(code.NoResult)
	}
	ti, err := t.taskIdentity.FindUserInstanceTask(t.db, req.InstanceID, req.TaskID, req.UserID, depID)
	if err != nil {
		return nil, err
	}
	nd, err := t.nodeRepo.FindByID(tx, tk.NodeID)
	if err != nil {
		return nil, err
	}
	if nd.NodeType == user {
		if err := t.userTransfer(tx, req, ti); err != nil {
			return nil, err
		}
	} else {
		if err := t.nonUserTransfer(tx, req); err != nil {
			return nil, err
		}
	}
	tx.Commit()
	return resp, nil
}

func (t *task) DeleteTask(ctx context.Context, req *DeleteTaskReq) (*DeleteTaskResp, error) {
	tx := t.db.Begin()
	tk, err := t.taskRepo.FindByID(t.db, req.TaskID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err = t.DoDeleteTask(tx, tk); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return nil, nil
}

// DoDeleteTask do delete
func (t *task) DoDeleteTask(tx *gorm.DB, task *models.Task) error {
	ht := &models.HistoryTask{}
	err := pkg.CopyProperties(ht, task)
	if err != nil {
		tx.Rollback()
		return err
	}
	ht.Status = internal.Deleted
	ht.EndTime = time2.Now()

	err = t.taskRepo.DeleteByID(tx, task.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = t.historyTaskRepo.Create(tx, ht)
	if err != nil {
		tx.Rollback()
		return err
	}

	// delete task identity
	err = t.taskIdentity.DeleteByTaskID(tx, task.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (t *task) AddNonNodeTask(ctx context.Context, req *AddTaskReq) (resp *AddTaskResp, err error) {
	tx := t.db.Begin()
	defer func() {
		if err != nil {
			logger.Logger.Errorw(err.Error(), logger.STDRequestID(ctx))
			tx.Rollback()
		}
	}()
	resp = &AddTaskResp{}
	cNode, err := component.NodeFactory(nonModel)
	if err != nil {
		return nil, err
	}
	exs, err := t.executionRepo.FindByInstanceID(tx, req.InstanceID)
	if err != nil {
		return nil, err
	}
	if len(exs) == 0 {
		return nil, error2.NewError(code.NoResult)
	}
	tasks := make([]*models.Task, 0)
	for i := 0; i < len(req.Assignee); i++ {
		initNodeReq := &component.InitNodeReq{
			Instance: &models.Instance{
				ProcID: exs[0].ProcID,
				ID:     req.InstanceID,
			},
			// 这里只能用instance的主executeID，如果是分支节点触发的，或许executeID对不上
			Execution: &models.Execution{
				ID: exs[0].ID,
			},
			Assignee:  req.Assignee[i],
			Name:      req.Name,
			Desc:      req.Desc,
			UserID:    req.UserID,
			NextNodes: req.NodeDefKey,
		}
		if err := cNode.Init(ctx, tx, initNodeReq, nil); err != nil {
			return nil, err
		}
		tasks = append(tasks, initNodeReq.InitResp.Tasks)
	}
	resp.Tasks = tasks
	tx.Commit()
	return resp, nil
}

// 前加签
func (t *task) AddNodeTask(ctx context.Context, req *AddTaskReq) (*AddTaskResp, error) {
	task, err := t.taskRepo.FindByID(t.db, req.TaskID)
	if err != nil {
		return nil, err
	}
	req.Node.InstanceID = task.ProcInstanceID
	nd, err := t.genInstanceNode(t.db, req)
	if err != nil {
		return nil, err
	}
	inst, err := t.instantRepo.FindByID(t.db, task.ProcInstanceID)
	if err != nil {
		return nil, err
	}
	execution, err := t.executionRepo.FindByID(t.db, task.ExecutionID)
	if err != nil {
		return nil, err
	}
	cNode, err := component.NodeFactory(nd.NodeType)
	if err != nil {
		return nil, err
	}
	initNodeReq := packInitNodeReq(nd, inst, execution, req.UserID, task.NodeDefKey, internal.TempModel)
	tx := t.db.Begin()
	if err = cNode.Init(ctx, tx, initNodeReq, nil); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err = t.DoDeleteTask(tx, task); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &AddTaskResp{
		Tasks: []*models.Task{initNodeReq.InitResp.Tasks},
	}, nil
}

// 后加签，加签不结束本任务，可以加签多次
func (t *task) AddBackNodeTask(ctx context.Context, req *AddTaskReq) (*AddTaskResp, error) {
	tx := t.db.Begin()
	task, err := t.taskRepo.FindByID(t.db, req.TaskID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	req.Node.InstanceID = task.ProcInstanceID
	if len(req.Node.NextNodes) > 0 {
		req.Node.NextNodes[0].NodeID = task.NextNodeDefKey
	} else {
		req.Node.NextNodes = []*NodeLinkData{{NodeID: task.NextNodeDefKey}}
	}
	nd, err := t.genInstanceNode(tx, req)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	task.NextNodeDefKey = nd.DefKey
	task.ModifierID = req.UserID
	task.ModifyTime = time2.Now()
	if err := t.taskRepo.Update(tx, task); err != nil {
		return nil, err
	}
	tx.Commit()
	return &AddTaskResp{}, nil
}

func (t *task) genInstanceNode(tx *gorm.DB, req *AddTaskReq) (*models.Node, error) {
	model := &models.Model{
		CreatorID: req.UserID,
	}
	nodes := []*NodeData{req.Node}
	res, err := t.node.AddNodes(tx, model, nodes)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res[0], nil
	}
	return nil, nil
}

func packInitNodeReq(node *models.Node, inst *models.Instance, execution *models.Execution, userID, nodeDefKey, taskType string) *component.InitNodeReq {
	initNodeReq := &component.InitNodeReq{
		Node:      node,
		Instance:  inst,
		UserID:    userID,
		Execution: execution,
		NextNodes: nodeDefKey,
		TaskType:  taskType,
	}
	return initNodeReq
}

// 打回重填
func (t *task) BackReFill(ctx context.Context, req *AddTaskReq) (resp *AddTaskResp, err error) {
	tx := t.db.Begin()
	defer func() {
		if err != nil {
			logger.Logger.Errorw(err.Error(), logger.STDRequestID(ctx))
			tx.Rollback()
		}
	}()
	resp = &AddTaskResp{}
	task, err := t.taskRepo.FindByID(tx, req.TaskID)
	if err != nil {
		return nil, err
	}
	ins, err := t.instantRepo.FindByID(tx, task.ProcInstanceID)
	if err != nil {
		return nil, err
	}
	tks, err := t.taskRepo.FindByInstanceID(tx, task.ProcInstanceID)
	if err != nil {
		return nil, err
	}
	node, err := t.nodeRepo.FindStartNode(tx, task.ProcID)
	if err != nil {
		return nil, err
	}
	nextNodeStr, err := t.nodeLinkRepo.FindNextByNodeID(tx, task.ProcID, node.ID)
	if err != nil {
		return nil, err
	}
	tk := &models.Task{
		ProcID:         task.ProcID,
		ProcInstanceID: task.ProcInstanceID,
		ExecutionID:    task.ExecutionID,
		NodeID:         node.ID,
		NodeDefKey:     node.DefKey,
		NextNodeDefKey: nextNodeStr,
		Name:           req.Name,
		Desc:           req.Desc,
		Assignee:       ins.CreatorID,
		TaskType:       internal.TempModel,
		Status:         internal.Active,
		CreatorID:      req.UserID,
	}
	if err = t.taskRepo.Create(tx, tk); err != nil {
		return nil, err
	}
	ti := &models.TaskIdentity{
		ID:           id2.GenID(),
		TaskID:       tk.ID,
		CreateTime:   time2.Now(),
		InstanceID:   tk.ProcInstanceID,
		CreatorID:    req.UserID,
		UserID:       tk.Assignee,
		IdentityType: internal.IdentityUser,
	}
	if err := t.taskIdentity.Create(tx, ti); err != nil {
		return nil, err
	}
	for i := 0; i < len(tks); i++ {
		if err = t.DoDeleteTask(tx, tks[i]); err != nil {
			return nil, err
		}
	}
	resp.Tasks = []*models.Task{tk}
	tx.Commit()
	return resp, nil
}

// 回退
func (t *task) FallbackTask(ctx context.Context, req *AddTaskReq) (resp *AddTaskResp, err error) {
	tx := t.db.Begin()
	defer func() {
		if err != nil {
			logger.Logger.Errorw(err.Error(), logger.STDRequestID(ctx))
			tx.Rollback()
		}
	}()
	resp = &AddTaskResp{}
	task, err := t.taskRepo.FindByID(t.db, req.TaskID)
	if err != nil {
		return nil, err
	}
	nd, err := t.nodeRepo.FindByDefKey(tx, task.ProcID, req.NodeDefKey)
	if err != nil {
		return nil, err
	}
	inst, err := t.instantRepo.FindByID(t.db, task.ProcInstanceID)
	if err != nil {
		return nil, err
	}
	execution, err := t.executionRepo.FindByID(t.db, task.ExecutionID)
	if err != nil {
		return nil, err
	}
	cNode, err := component.NodeFactory(nd.NodeType)
	if err != nil {
		return nil, err
	}
	initNodeReq := packInitNodeReq(nd, inst, execution, req.UserID, "", "")
	if err = cNode.Init(ctx, tx, initNodeReq, nil); err != nil {
		return nil, err
	}
	if err = t.DoDeleteTask(tx, task); err != nil {
		return nil, err
	}
	resp.Tasks = []*models.Task{initNodeReq.InitResp.Tasks}
	tx.Commit()
	return resp, nil
}

// 增加一个已办任务
func (t *task) AddHistoryTask(ctx context.Context, req *AddHistoryTaskReq) (*AddHistoryTaskResp, error) {
	inst, err := t.instantRepo.FindByID(t.db, req.InstanceID)
	if err != nil {
		return nil, err
	}
	node, err := t.nodeRepo.FindByDefKey(t.db, inst.ProcID, req.NodeDefKey)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, error2.NewError(code.InvalidParams)
	}
	nextNodeStr, err := t.nodeLinkRepo.FindNextByNodeID(t.db, node.ProcID, node.ID)
	if err != nil {
		return nil, err
	}
	ht := &models.HistoryTask{
		ID:             id2.GenID(),
		ProcID:         inst.ProcID,
		NodeID:         node.ID,
		ProcInstanceID: req.InstanceID,
		NodeDefKey:     node.DefKey,
		NextNodeDefKey: nextNodeStr,
		Assignee:       req.Assignee,
		Name:           req.Name,
		Desc:           req.Desc,
		ExecutionID:    req.ExecutionID,
		CreateTime:     time2.Now(),
		CreatorID:      req.UserID,
		TaskType:       internal.TempModel,
		Status:         internal.Completed,
	}
	if err = t.historyTaskRepo.Create(t.db, ht); err != nil {
		return nil, err
	}
	if err = t.updateTaskInstanceModifyTime(req.InstanceID); err != nil {
		return nil, err
	}
	return &AddHistoryTaskResp{
		Tasks: []*models.HistoryTask{ht},
	}, nil
}

// InitTask init next task by component info
func (t *task) InitTask(ctx context.Context, tx *gorm.DB, req *component.CompleteNodeReq) (*pb.NodeEventRespData, error) {
	nextNodeDefKeys := strings.Split(req.NextNodes, ",")
	if len(nextNodeDefKeys) > 0 {
		for _, value := range nextNodeDefKeys {
			d := &pb.NodeEventReqData{
				ProcessInstanceID: req.Instance.ID,
				ProcessID:         req.Instance.ProcID,
				NodeDefKey:        value,
				ExecutionID:       req.Execution.ID,
			}

			nodeInitBeginResp, err := t.PublishMessage(ctx, internal.SynchronizationMode, internal.NodeInitBeginEvent, d)

			if err != nil {
				return nil, err
			}
			fmt.Println("rpc返回完成", d, time2.NowUnixMill())
			nextNode, err := t.nodeRepo.FindByDefKey(tx, req.Instance.ProcID, value)
			if err != nil {
				return nil, err
			}
			if nextNode == nil {
				nextNode, err = t.nodeRepo.FindInstanceNode(tx, req.Instance.ID, value)
				if err != nil {
					return nil, err
				}
			}
			cNode, err := component.NodeFactory(nextNode.NodeType)
			if err != nil {
				return nil, err
			}
			initNodeReq := &component.InitNodeReq{
				Execution: req.Execution,
				Instance:  req.Instance,
				Node:      nextNode,
				UserID:    req.UserID,
				Params:    req.Params,
			}
			err = cNode.Init(ctx, tx, initNodeReq, nodeInitBeginResp)
			if err != nil {
				return nil, err
			}
			d.TaskID = strings.Split(initNodeReq.InitResp.EventTaskID, ",")
			return t.PublishMessage(ctx, internal.AsynchronousMode, internal.NodeInitEndEvent, d)
		}
	}
	return nil, nil
}

func (t *task) CompleteExecution(ctx context.Context, req *CompleteExecutionReq) (resp *CompleteTaskResp, err error) {
	tx := t.db.Begin()
	defer func() {
		if err != nil {
			logger.Logger.Errorw(err.Error(), logger.STDRequestID(ctx))
			tx.Rollback()
		}
	}()
	resp = &CompleteTaskResp{}
	// 用于判断分支是否都已经结束，是就到合流网关去
	for _, exID := range req.ExecutionID {
		ex, err := t.executionRepo.FindByID(tx, exID)
		if err != nil {
			return nil, err
		}
		if ex == nil {
			return nil, error2.NewError(code.InvalidExecutionID)
		}
		if ex.PID == "" {
			return nil, error2.NewError(code.NoBranchExecution)
		}
		// 该分支可能存在的下级分支的任务也结束掉,递归查找
		ids := []string{ex.ID}
		childEx := make([]*models.Execution, 0)
		err = t.findChildExecution(ex.ProcInstanceID, ex.ID, &childEx)
		if err != nil {
			return nil, err
		}
		for _, cex := range childEx {
			ids = append(ids, cex.ID)
		}
		tks, err := t.taskRepo.FindByExecutionIDs(tx, ids)
		if err != nil {
			return nil, err
		}
		for j := 0; j < len(tks); j++ {
			if err := t.DoDeleteTask(tx, tks[j]); err != nil {
				return nil, err
			}
			if err = t.taskIdentity.DeleteByTaskID(tx, tks[j].ID); err != nil {
				return nil, err
			}
		}
		// 需要递归结束掉所有子分支
		if err := t.setBranchExecutionActive(tx, ex.ProcInstanceID, ex.ID); err != nil {
			return nil, err
		}
		// 需要递归查询分支的子分支,子分支还没有完成的情况下，主分支也应该是未完成状态
		exps := make([]*models.Execution, 0)
		err = t.findActiveExecution(tx, ex.ProcInstanceID, ex.PID, &exps)
		if err != nil {
			return nil, err
		}
		// 所有分支都结束了，跳转到指定节点
		if len(exps) == 0 && req.NodeDefKey != "" {
			nd, err := t.nodeRepo.FindByDefKey(tx, ex.ProcID, req.NodeDefKey)
			if err != nil {
				return nil, err
			}
			if nd == nil {
				return nil, error2.NewError(code.InvalidParams)
			}
			cNode, err := component.NodeFactory(nd.NodeType)
			if err != nil {
				return nil, err
			}
			exto := ex
			// 如果是跳到合流网关，需要找到该网关的execution
			if nd.NodeType == inclusiveGateway {
				freq := &ParentExecutionReq{
					TaskID: req.TaskID,
					DefKey: req.NodeDefKey,
				}
				exsp, err := t.FindParentExecutions(ctx, freq)
				if err != nil {
					return nil, err
				}
				exto, err = t.executionRepo.FindByID(tx, exsp.ExecutionID)
				if err != nil {
					return nil, err
				}
			}
			initNodeReq := &component.InitNodeReq{
				Execution: exto,
				Instance: &models.Instance{
					ID:     ex.ProcInstanceID,
					ProcID: ex.ProcID,
				},
				Node: nd,
			}
			err = cNode.Init(ctx, tx, initNodeReq, nil)
			if err != nil {
				return nil, err
			}
		}
	}
	// 为结束的任务增加已办记录
	if req.TaskID != "" {
		entity := &models.HistoryTask{
			ID:       req.TaskID,
			Assignee: req.UserID,
			Status:   internal.Completed,
		}
		err := t.historyTaskRepo.Update(tx, entity)
		if err != nil {
			return nil, err
		}
	}
	tx.Commit()
	return resp, nil
}

func (t *task) setBranchExecutionActive(tx *gorm.DB, instanceID, executionID string) error {
	if err := t.executionRepo.SetActive(tx, executionID, 0); err != nil {
		return err
	}
	exs, err := t.executionRepo.FindByPID(tx, instanceID, executionID, 1)
	if err != nil {
		return err
	}
	for i := 0; i < len(exs); i++ {
		if err := t.setBranchExecutionActive(tx, instanceID, exs[i].ID); err != nil {
			return err
		}
	}
	return nil
}

func (t *task) CompleteTask(ctx context.Context, req *CompleteTaskReq) (resp *CompleteTaskResp, err error) {
	tx := t.db.Begin()
	defer func() {
		if err != nil {
			logger.Logger.Errorw(err.Error(), logger.STDRequestID(ctx))
			tx.Rollback()
		}
	}()
	resp = &CompleteTaskResp{}
	comReq, err := t.packCompleteReq(tx, req)
	if err != nil {
		return nil, err
	}
	if comReq.Task.TaskType == internal.NonModel {
		comReq.Node = &models.Node{
			NodeType: nonModel,
		}
	}
	cNode, err := component.NodeFactory(comReq.Node.NodeType)
	if err != nil {
		return nil, err
	}
	// U := models.Task{
	//	ID:       comReq.Task.ID,
	//	Comments: comReq.Comments,
	// }
	// if err := t.taskRepo.Update(tx, &U); err != nil {
	//	return nil, err
	// }
	completeStatus, err := cNode.Complete(ctx, tx, comReq)
	if err != nil {
		return nil, err
	}
	if completeStatus {
		commentsObj, err := gabs.ParseJSON([]byte(req.Comments))
		if (commentsObj.Path("reviewResult").Data().(string)) != "REFUSE" {
			// 如果调用方指定了nextNode来初始化，优先
			if req.NextNodeDefKey != "" {
				fNode, err := component.NodeFactory(free)
				if err != nil {
					return nil, err
				}
				initNodeReq := &component.InitNodeReq{
					Task:      comReq.Task,
					Execution: comReq.Execution,
					Instance:  comReq.Instance,
					NextNodes: req.NextNodeDefKey,
					UserID:    req.UserID,
					Params:    req.Params,
				}
				if err = fNode.Init(ctx, tx, initNodeReq, nil); err != nil {
					return nil, err
				}
				tx.Commit()
				return resp, nil
			}
			switch comReq.Node.NodeType {
			case multiUser:
				ex, err := t.executionRepo.FindByID(tx, comReq.Execution.PID)
				if err != nil {
					return nil, err
				}
				comReq.Execution = ex
			}
			_, err = t.InitTask(ctx, tx, comReq)
			if err != nil {
				return nil, err
			}
		}
		// delete complete task identity
		err = t.taskIdentity.DeleteByTaskID(tx, req.TaskID)
		if err != nil {
			return nil, err
		}
		// temp-model task need delete
		if comReq.Task.TaskType == internal.TempModel {
			err = t.taskRepo.DeleteByID(tx, comReq.Task.ID)
			if err != nil {
				return nil, err
			}
		}
	}
	tx.Commit()
	return resp, nil
}

func (t *task) packCompleteReq(tx *gorm.DB, req *CompleteTaskReq) (*component.CompleteNodeReq, error) {
	task, err := t.taskRepo.FindByID(tx, req.TaskID)
	if err != nil {
		return nil, err
	}
	inst, err := t.instantRepo.FindByID(tx, task.ProcInstanceID)
	if err != nil {
		return nil, err
	}
	execution, err := t.executionRepo.FindByID(tx, task.ExecutionID)
	if err != nil {
		return nil, err
	}
	currentNode, err := t.nodeRepo.FindByID(tx, task.NodeID)
	if err != nil {
		return nil, err
	}
	completeNodeReq := &component.CompleteNodeReq{
		Task:      task,
		Execution: execution,
		Instance:  inst,
		NextNodes: task.NextNodeDefKey,
		Node:      currentNode,
		UserID:    req.UserID,
		Params:    req.Params,
		Comments:  req.Comments,
	}
	return completeNodeReq, nil
}

func (t *task) BatchCompleteNonModelTask(ctx context.Context, req *CompleteNonModelTaskReq) (*CompleteTaskResp, error) {
	tx := t.db.Begin()
	cNode, err := component.NodeFactory(nonModel)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	for i := 0; i < len(req.TaskID); i++ {
		comReq := &component.CompleteNodeReq{
			Task:   &models.Task{ID: req.TaskID[i]},
			UserID: req.UserID,
		}
		_, err = cNode.Complete(ctx, tx, comReq)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		err = t.taskIdentity.DeleteByTaskID(tx, req.TaskID[i])
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	tx.Commit()
	return &CompleteTaskResp{}, nil
}

func (t *task) SetDB(db *gorm.DB) {
	t.db = db
}

func (t *task) SetListener(l *listener.Listener) {
	t.l = l
}

// PublishMessage PublishMessage
// func (t *task) PublishMessage(ctx context.Context, mode, name string, data map[string]string) error {
// 	ms := &listener.EventMessage{
// 		EventName: name,
// 		EventData: data,
// 		Message: &listener.Message{
// 			MessageType:     "eventMessage",
// 			MessageSendMode: mode,
// 		},
// 	}
// 	err := t.l.Notify(ctx, ms)
// 	return err
// }

func (t *task) PublishMessage(ctx context.Context, mode, name string, data *pb.NodeEventReqData) (*pb.NodeEventRespData, error) {
	ms := &listener.EventMessage{
		EventName: name,
		EventData: data,
		Message: &listener.Message{
			MessageType:     "eventMessage",
			MessageSendMode: mode,
		},
	}

	return t.l.Notify(ctx, ms)
}

func (t *task) updateTaskInstanceModifyTime(instanceID string) error {
	et, err := t.instantRepo.FindByID(t.db, instanceID)
	if err != nil {
		return err
	}
	if et != nil {
		if err = t.instantRepo.Update(t.db, et); err != nil {
			return err
		}
	}
	return nil
}

func (t *task) R() *rc.ClusterClient {
	return t.redisClient
}

// TaskPreNodeReq TaskPreNodeReq
type TaskPreNodeReq struct {
	TaskID string `json:"taskID" binding:"required"`
	UserID string `json:"userID"`
}

// TaskPreNodeResp TaskPreNodeResp
type TaskPreNodeResp struct {
	Nodes []*nodeInfo `json:"nodes"`
}

type nodeInfo struct {
	Name       string `json:"name"`
	NodeDefKey string `json:"nodeDefKey"`
	NodeType   string `json:"nodeType"`
}

// GateWayExecutionReq GateWayExecutionReq
type GateWayExecutionReq struct {
	TaskID string `json:"taskID" binding:"required"`
	// UserID string `json:"userID"`
}

// GateWayExecutionResp GateWayExecutionResp
type GateWayExecutionResp struct {
	Executions         []string `json:"executionIds"`
	CurrentExecutionID string   `json:"currentExecutionID"`
}

// ParentExecutionReq ParentExecutionReq
type ParentExecutionReq struct {
	TaskID string `json:"taskID" binding:"required"`
	DefKey string `json:"defKey"`
}

// ParentExecutionResp ParentExecutionResp
type ParentExecutionResp struct {
	ExecutionID string `json:"executionID"`
}

// InitTaskReq InitTaskReq
type InitTaskReq struct {
	Instance    *models.Instance
	Node        *models.Node
	ExecutionID string
	Assignee    string
	UserID      string
}

// CompleteTaskReq CompleteTaskReq
type CompleteTaskReq struct {
	InstanceID     string                 `json:"instanceID"`
	TaskID         string                 `json:"taskID" binding:"required"`
	UserID         string                 `json:"userID"`
	NextNodeDefKey string                 `json:"nextNodeDefKey"`
	Params         map[string]interface{} `json:"params"`
	Comments       string                 `json:"comments"`
}

// CompleteNonModelTaskReq CompleteNonModelTaskReq
type CompleteNonModelTaskReq struct {
	TaskID []string `json:"tasks" binding:"required"`
	UserID string   `json:"userID"`
}

// CompleteTaskResp CompleteTaskResp
type CompleteTaskResp struct{}

// CompleteExecutionReq CompleteExecutionReq
type CompleteExecutionReq struct {
	// InstanceID     string   `json:"instanceID"`
	TaskID      string   `json:"taskID"`
	ExecutionID []string `json:"executionID"`
	UserID      string   `json:"userID"`
	NodeDefKey  string   `json:"nextDefKey"`
	Comments    string   `json:"comments"`
}

// DeleteTaskReq DeleteTaskReq
type DeleteTaskReq struct {
	TaskID string `json:"taskID"`
	UserID string `json:"userID"`
}

// InstanceTaskReq InstanceTaskReq
type InstanceTaskReq struct {
	InstanceID string `json:"instanceID"`
}

// DeleteTaskResp DeleteTaskResp
type DeleteTaskResp struct{}

// QueryTaskReq QueryTaskReq
type QueryTaskReq struct {
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	Des        []string            `json:"desc"`
	Name       string              `json:"taskName"`
	NodeDefKey string              `json:"nodeDefKey"`
	Order      []models.QueryOrder `json:"orders"`
	ProcessID  []string            `json:"processID"`
	InstanceID []string            `json:"instanceID"`
	TaskID     []string            `json:"taskID"`
	Assignee   string              `json:"assignee"`
	DueTime    string              `json:"dueTime"`
	Status     string              `json:"status"`
}

// AddTaskReq AddTaskReq
type AddTaskReq struct {
	Name       string    `json:"name"`
	Desc       string    `json:"desc"`
	UserID     string    `json:"userID"`
	Assignee   []string  `json:"assignee"`
	TaskID     string    `json:"taskID"`
	NodeDefKey string    `json:"nodeDefKey"`
	InstanceID string    `json:"instanceID"`
	Node       *NodeData `json:"node"`
}

// AddHistoryTaskReq AddHistoryTaskReq
type AddHistoryTaskReq struct {
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	UserID      string `json:"userID"`
	Assignee    string `json:"assignee"`
	TaskID      string `json:"taskID"`
	NodeDefKey  string `json:"nodeDefKey"`
	InstanceID  string `json:"instanceID"`
	ExecutionID string `json:"executionID"`
}

// AddTaskConditionReq AddTaskConditionReq
type AddTaskConditionReq struct {
	UserID     string `json:"userID"`
	Assignee   string `json:"assignee"`
	TaskID     string `json:"taskID"`
	InstanceID string `json:"instanceID"`
	DueTime    string `json:"dueTime"`
}

// AddTaskResp AddTaskResp
type AddTaskResp struct {
	Tasks []*models.Task `json:"task"`
}

// AddHistoryTaskResp AddHistoryTaskResp
type AddHistoryTaskResp struct {
	Tasks []*models.HistoryTask `json:"task"`
}

// TotalTaskResp TotalTaskResp
type TotalTaskResp struct {
	Total int64 `json:"total"`
}
