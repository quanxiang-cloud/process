package process

import (
	"context"
	"encoding/json"
	"fmt"
	"git.internal.yunify.com/qxp/misc/error2"
	"git.internal.yunify.com/qxp/misc/id2"
	"git.internal.yunify.com/qxp/misc/time2"
	"git.internal.yunify.com/qxp/process/internal"
	"git.internal.yunify.com/qxp/process/internal/component"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/internal/models/mysql"
	listener "git.internal.yunify.com/qxp/process/internal/server/events"
	"git.internal.yunify.com/qxp/process/internal/server/options"
	"git.internal.yunify.com/qxp/process/pkg/code"
	"git.internal.yunify.com/qxp/process/pkg/config"
	"git.internal.yunify.com/qxp/process/pkg/page"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"time"
)

// Instance process instance
type Instance interface {
	Start(ctx context.Context, req *StartProcessReq) (*StartProcessResp, error)
	InitInstance(ctx context.Context, req *InitInstanceReq) (*StartProcessResp, error)
	DeleteInstance(ctx context.Context, req *DeleteProcessReq) (*DeleteProcessResp, error)
	TerminatedInstance(ctx context.Context, req *DeleteProcessReq) (*DeleteProcessResp, error)
	InstanceList(ctx context.Context, req *ListProcessReq) (*ListProcessResp, error)
	DoneInstance(ctx context.Context, req *QueryTaskReq) (*page.Page, error)
	AgencyInstance(ctx context.Context, req *QueryTaskReq) (*page.Page, error)
	WholeInstance(ctx context.Context, req *QueryTaskReq) (*page.Page, error)
	SaveVariables(ctx context.Context, req *SaveVariablesReq) (*SaveVariablesResp, error)
	FindVariables(ctx context.Context, req *GetVariablesReq) (map[string]interface{}, error)
	AppDeleteHandler(ctx context.Context, req *AppDelReq) (*AppDelResp, error)
	CompleteNode(ctx context.Context, req *InitNextNodeReq) error

	NodeInstanceList(ctx context.Context, req *NodeInstanceListReq) ([]*models.NodeInstanceVO, error)
}

const saveTime = time.Hour * time.Duration(12)

// StartProcessReq start process request
type StartProcessReq struct {
	ProcessID string                 `json:"processID"`
	UserID    string                 `json:"userID"`
	Params    map[string]interface{} `json:"params"`
}

// InitInstanceReq init instance request
type InitInstanceReq struct {
	InstanceID string                 `json:"instanceID"`
	UserID     string                 `json:"userID"`
	Params     map[string]interface{} `json:"params"`
}

// DeleteProcessReq delete process request
type DeleteProcessReq struct {
	InstanceID string `json:"instanceID"`
	UserID     string `json:"userID"`
}

// SaveVariablesReq save variables request
type SaveVariablesReq struct {
	ProcessID  string      `json:"processID"`
	InstanceID string      `json:"instanceID"`
	NodeDefKey string      `json:"nodeDefKey"`
	UserID     string      `json:"userID"`
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
}

// GetVariablesReq get variables request
type GetVariablesReq struct {
	InstanceID string   `json:"instanceID"`
	Key        []string `json:"keys"`
}

// AppDelReq add delete update app status
type AppDelReq struct {
	AppDefKey []string `json:"defKey"`
	Action    string   `json:"action"`
}

// AppDelResp AppDelResp
type AppDelResp struct{}

// SaveVariablesResp save variables
type SaveVariablesResp struct{}

// ListProcessReq list process request
type ListProcessReq struct {
	Name             string   `json:"name"`
	ProcessID        []string `json:"processID"`
	InstanceID       []string `json:"instanceID"`
	ParentInstanceID []string `json:"parentInstanceID"`
	ProcessStatus    string   `json:"processStatus"`
}

// ListProcessResp list process request
type ListProcessResp struct {
	Instances []*models.Instance `json:"instances"`
}

// StartProcessResp start process resp
type StartProcessResp struct {
	InstanceID string `json:"instanceID"`
}

// DeleteProcessResp DeleteProcessResp
type DeleteProcessResp struct{}

type instance struct {
	db               *gorm.DB
	l                *listener.Listener
	conf             *config.Configs
	modalRepo        models.ModelRepo
	instanceRepo     models.InstanceRepo
	taskRepo         models.TaskRepo
	historyTaskRepo  models.HistoryTaskRepo
	nodeRepo         models.NodeRepo
	nodeLinkRepo     models.NodeLinkRepo
	executionRepo    models.ExecutionRepo
	variablesRepo    models.VariablesRepo
	nodeInstanceRepo models.NodeInstanceRepo
	task             Task
}

// NewInstance NewInstance
func NewInstance(conf *config.Configs, opts ...options.Options) (Instance, error) {
	task, err := NewTask(conf, opts...)
	if err != nil {
		return nil, err
	}
	inst := &instance{
		conf:             conf,
		task:             task,
		modalRepo:        mysql.NewModelRepo(),
		instanceRepo:     mysql.NewInstanceRepo(),
		taskRepo:         mysql.NewTaskRepo(),
		historyTaskRepo:  mysql.NewHistoryTaskRepo(),
		nodeRepo:         mysql.NewNodeRepo(),
		nodeLinkRepo:     mysql.NewNodeLinkRepo(),
		executionRepo:    mysql.NewExecutionRepo(),
		variablesRepo:    mysql.NewVariablesRepo(),
		nodeInstanceRepo: mysql.NewNodeInstanceRepo(),
	}
	for _, opt := range opts {
		opt(inst)
	}
	return inst, nil
}

func (i *instance) Transaction() *gorm.DB {
	return i.db.Begin()
}

func (i *instance) SetDB(db *gorm.DB) {
	i.db = db
}

func (i *instance) SetListener(l *listener.Listener) {
	i.l = l
}

func (i *instance) InitInstance(ctx context.Context, req *InitInstanceReq) (*StartProcessResp, error) {
	tx := i.db.Begin()
	tks, err := i.taskRepo.FindByInstanceID(tx, req.InstanceID)
	if err != nil {
		return nil, err
	}
	if len(tks) > 0 {
		return nil, error2.NewError(code.InstanceInitError)
	}
	inst, err := i.instanceRepo.FindByID(tx, req.InstanceID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	// ---------------------create excution begin--------------------------
	execution := &models.Execution{
		ProcID:         inst.ProcID,
		ProcInstanceID: req.InstanceID,
		PID:            "",
		IsActive:       1,
		CreatorID:      req.UserID,
	}
	err = i.executionRepo.Create(tx, execution)
	if err != nil {
		return nil, err
	}
	// ---------------------create excution end--------------------------

	startNode, err := i.nodeRepo.FindStartNode(tx, inst.ProcID)
	if err != nil {
		return nil, err
	}
	// --------------------complete start component begin-------------------
	cNode, err := component.NodeFactory("Start")
	if err != nil {
		return nil, err
	}
	nextNodeStr, err := i.nodeLinkRepo.FindNextByNodeID(tx, startNode.ProcID, startNode.ID)
	if err != nil {
		return nil, err
	}
	completeReq := &component.CompleteNodeReq{
		Execution: execution,
		Instance:  inst,
		Node:      startNode,
		UserID:    req.UserID,
		NextNodes: nextNodeStr,
		Params:    req.Params,
	}
	_, err = cNode.Complete(ctx, tx, completeReq)
	if err != nil {
		return nil, err
	}
	// --------------------complete start component end-------------------

	// init next task
	_, err = i.task.InitTask(ctx, tx, completeReq)
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return &StartProcessResp{
		InstanceID: req.InstanceID,
	}, nil
}

func (i *instance) SaveVariables(ctx context.Context, req *SaveVariablesReq) (*SaveVariablesResp, error) {
	node, err := i.nodeRepo.FindByDefKey(i.db, req.ProcessID, req.NodeDefKey)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, error2.NewError(code.InvalidParams)
	}
	variables, err := i.variablesRepo.FindVariablesByName(i.db, req.InstanceID, node.ID, req.Key)
	if err != nil {
		return nil, err
	}
	if variables == nil {
		variables = &models.Variables{
			ID:             id2.GenID(),
			Name:           req.Key,
			NodeID:         node.ID,
			CreatorID:      req.UserID,
			CreateTime:     time2.Now(),
			ProcInstanceID: req.InstanceID,
		}
		if err = i.packValue(req, variables); err != nil {
			return nil, err
		}
		if err = i.variablesRepo.Create(i.db, variables); err != nil {
			return nil, err
		}
		if err = i.saveVariablesRedis(ctx, variables, saveTime); err != nil {
			return nil, err
		}
		return &SaveVariablesResp{}, nil
	}
	if err = i.packValue(req, variables); err != nil {
		return nil, err
	}
	if err = i.variablesRepo.Update(i.db, variables); err != nil {
		return nil, err
	}
	if err = i.saveVariablesRedis(ctx, variables, saveTime); err != nil {
		return nil, err
	}
	return &SaveVariablesResp{}, nil
}

func (i *instance) saveVariablesRedis(ctx context.Context, variables *models.Variables, ttl time.Duration) error {
	entityJSON, err := json.Marshal(variables)
	if err != nil {
		return err
	}
	key := internal.RedisPreKey + variables.ProcInstanceID + ":" + variables.NodeID + ":" + variables.Name
	return i.task.R().Set(ctx, key, entityJSON, ttl).Err()
}

func (i *instance) packValue(req *SaveVariablesReq, variables *models.Variables) error {
	v := reflect.ValueOf(req.Value)
	switch reflect.TypeOf(req.Value).Kind() {
	case reflect.String:
		variables.Value = v.String()
		variables.VarType = "string"
	case reflect.Slice, reflect.Array:
		entity, err := json.Marshal(v.Interface())
		if err != nil {
			return err
		}
		variables.ComplexValue = entity
		variables.VarType = "[]string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		variables.Value = fmt.Sprintf("%d", v.Int())
		variables.VarType = "int"
	case reflect.Float32, reflect.Float64:
		variables.Value = fmt.Sprintf("%f", v.Float())
		variables.VarType = "float"
	}
	return nil
}

func (i *instance) FindVariables(ctx context.Context, req *GetVariablesReq) (map[string]interface{}, error) {
	rest := make(map[string]interface{}, 0)
	res, err := i.variablesRepo.GetInstanceValueByName(i.db, req.InstanceID, req.Key)
	if err != nil {
		return rest, err
	}
	return res, nil
}

func (i *instance) DoneInstance(ctx context.Context, req *QueryTaskReq) (*page.Page, error) {
	cd := packTaskCondition(req)
	list, total := i.instanceRepo.FindByUserID(i.db, req.Page, req.Limit, cd)
	pages := page.NewPage(req.Page, req.Limit, total)
	pages.Data = list
	return pages, nil
}

func (i *instance) AgencyInstance(ctx context.Context, req *QueryTaskReq) (*page.Page, error) {
	cd := packTaskCondition(req)
	if req.Assignee != "" {
		depID, err := i.task.QueryGroupIDByUserID(ctx, req.Assignee)
		if err != nil {
			return nil, err
		}
		cd.GroupID = depID
	}
	list, total := i.instanceRepo.FindAgencyByUserID(i.db, req.Page, req.Limit, cd)
	pages := page.NewPage(req.Page, req.Limit, total)
	pages.Data = list
	return pages, nil
}

func (i *instance) WholeInstance(ctx context.Context, req *QueryTaskReq) (*page.Page, error) {
	cd := packTaskCondition(req)
	if req.Assignee != "" {
		depID, err := i.task.QueryGroupIDByUserID(ctx, req.Assignee)
		if err != nil {
			return nil, err
		}
		cd.GroupID = depID
	}
	list, total := i.instanceRepo.FindAllByUserID(i.db, req.Page, req.Limit, cd)
	pages := page.NewPage(req.Page, req.Limit, total)
	pages.Data = list
	return pages, nil
}

func packTaskCondition(req *QueryTaskReq) *models.QueryTaskCondition {
	cd := &models.QueryTaskCondition{
		ProcessID:  req.ProcessID,
		InstanceID: req.InstanceID,
		NodeDefKey: req.NodeDefKey,
		Name:       req.Name,
		Des:        strings.Join(req.Des, "|"),
		UserID:     req.Assignee,
		DueTime:    req.DueTime,
		Order:      req.Order,
	}
	return cd
}

// Start start a instance
func (i *instance) Start(ctx context.Context, req *StartProcessReq) (resp *StartProcessResp, err error) {
	tx := i.Transaction()
	resp = &StartProcessResp{}
	// ---------------------create instance begin--------------------------
	model, err := i.modalRepo.FindByID(tx, req.ProcessID)
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, error2.NewError(code.InvalidProcessID)
	}
	inst := &models.Instance{
		ID:         id2.GenID(),
		ProcID:     req.ProcessID,
		PID:        "",
		Name:       model.Name,
		CreatorID:  req.UserID,
		CreateTime: time2.Now(),
		Status:     internal.Active,
	}
	err = i.instanceRepo.Create(tx, inst)
	if err != nil {
		return nil, err
	}
	resp.InstanceID = inst.ID
	tx.Commit()
	return resp, nil
}

func (i *instance) TerminatedInstance(ctx context.Context, req *DeleteProcessReq) (*DeleteProcessResp, error) {
	tx := i.db.Begin()
	if err := i.clearInstance(ctx, tx, req); err != nil {
		tx.Rollback()
		return nil, err
	}
	ins := &models.Instance{
		ID:         req.InstanceID,
		ModifierID: req.UserID,
		Status:     internal.Terminated,
	}
	err := i.instanceRepo.Update(tx, ins)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &DeleteProcessResp{}, nil
}

func (i *instance) DeleteInstance(ctx context.Context, req *DeleteProcessReq) (*DeleteProcessResp, error) {
	tx := i.db.Begin()
	if err := i.clearInstance(ctx, tx, req); err != nil {
		tx.Rollback()
		return nil, err
	}
	ins := &models.Instance{
		ID:         req.InstanceID,
		ModifierID: req.UserID,
		Status:     internal.Deleted,
	}
	err := i.instanceRepo.Update(tx, ins)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &DeleteProcessResp{}, nil
}

func (i *instance) clearInstance(ctx context.Context, tx *gorm.DB, req *DeleteProcessReq) error {
	taskReq := &InstanceTaskReq{
		InstanceID: req.InstanceID,
	}
	res, err := i.task.InstanceTasks(ctx, taskReq)
	if err != nil {
		return err
	}
	for _, tk := range res {
		if err := i.task.DoDeleteTask(tx, tk); err != nil {
			tx.Rollback()
			return err
		}
	}
	err = i.executionRepo.DeleteByInstanceID(tx, req.InstanceID)
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (i *instance) InstanceList(ctx context.Context, req *ListProcessReq) (*ListProcessResp, error) {
	res, _ := i.instanceRepo.FindPageInstance(i.db, 0, 1000, req.ProcessID, req.InstanceID, req.ParentInstanceID, req.Name, req.ProcessStatus)

	return &ListProcessResp{
		Instances: res,
	}, nil
}

// func (i *instance) PublishMessage(ctx context.Context, name string, data map[string]string) error {
// 	ms := &listener.EventMessage{
// 		EventName: name,
// 		EventData: data,
// 		Message: &listener.Message{
// 			MessageType:     "processEvent",
// 			MessageSendMode: "synchronization",
// 		},
// 	}
// 	err := i.l.Notify(ctx, ms)
// 	return err
// }

func (i *instance) AppDeleteHandler(ctx context.Context, req *AppDelReq) (*AppDelResp, error) {
	switch req.Action {
	case internal.DoSuspend:
		if err := i.instanceRepo.UpdateAppByProcDefKey(i.db, req.AppDefKey, internal.Suspend); err != nil {
			return nil, err
		}
	case internal.DoReactive:
		if err := i.instanceRepo.UpdateAppByProcDefKey(i.db, req.AppDefKey, internal.Active); err != nil {
			return nil, err
		}
	case internal.DoDump:
		if err := i.instanceRepo.DeleteAppByProcDefKey(i.db, req.AppDefKey); err != nil {
			return nil, err
		}
	}
	return &AppDelResp{}, nil
}

func (i *instance) CompleteNode(ctx context.Context, req *InitNextNodeReq) error {
	node, err := i.nodeRepo.FindByDefKey(i.db, req.ProcessID, req.NodeDefKey)
	if err != nil {
		return err
	}

	cNode, err := component.NodeFactory(node.NodeType)
	if err != nil {
		return err
	}

	execution, err := i.executionRepo.FindByID(i.db, req.ExecutionID)
	if err != nil {
		return err
	}

	instance, err := i.instanceRepo.FindByID(i.db, req.InstanceID)
	if err != nil {
		return err
	}

	completeNodeReq := &component.CompleteNodeReq{
		Execution: execution,
		Instance:  instance,
		Node:      node,
		NextNodes: req.NextNodes,
		UserID:    req.UserID,
		Params:    req.Params,
	}

	_, err = cNode.Complete(ctx, i.db, completeNodeReq)
	return err
}

// NodeInstanceList get node instance list
func (i *instance) NodeInstanceList(ctx context.Context, req *NodeInstanceListReq) ([]*models.NodeInstanceVO, error) {
	nodeInstances, err := i.nodeInstanceRepo.FindByInstanceID(i.db, req.ProcInstanceID)
	// 关联查询task信息
	if err != nil {
		return nil, err
	}

	taskIDs := make([]string, 0)
	for _, v := range nodeInstances {
		if len(v.TaskID) > 0 {
			taskIDs = append(taskIDs, v.TaskID)
		}
	}

	tasks, err := i.taskRepo.FindByIDs(i.db, taskIDs)
	if err != nil {
		return nil, err
	}
	historyTasks, err := i.historyTaskRepo.FindByIDs(i.db, taskIDs)
	if err != nil {
		return nil, err
	}

	taskMap := make(map[string]string, 0)
	for _, task := range tasks {
		taskMap[task.ID] = task.Assignee
	}
	for _, task := range historyTasks {
		taskMap[task.ID] = task.Assignee
	}

	nodeInstanceVOs := make([]*models.NodeInstanceVO, 0)
	for _, v := range nodeInstances {
		nodeInstanceVO := &models.NodeInstanceVO{
			NodeInstance: v,
		}
		if len(v.TaskID) > 0 {
			nodeInstanceVO.Assignee = taskMap[v.TaskID]
		}

		nodeInstanceVOs = append(nodeInstanceVOs, nodeInstanceVO)
	}

	return nodeInstanceVOs, nil
}
