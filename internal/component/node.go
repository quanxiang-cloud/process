package component

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	rc "github.com/go-redis/redis/v8"
	"github.com/quanxiang-cloud/process/internal"
	"github.com/quanxiang-cloud/process/internal/models"
	listener "github.com/quanxiang-cloud/process/internal/server/events"
	"github.com/quanxiang-cloud/process/pkg/misc/id2"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	"github.com/quanxiang-cloud/process/rpc/pb"
	"gorm.io/gorm"
	"strings"
)

// ExecuteType
const (
	PauseExecution = "pauseExecution"
	EndExecution   = "endExecution"
	EndProcess     = "endProcess"
)

// Node Node
type Node struct {
	Listener         *listener.Listener
	TaskIdentity     models.TaskIdentityRepo
	NodeInstanceRepo models.NodeInstanceRepo
	ExecutionRepo    models.ExecutionRepo
	NodeRepo         models.NodeRepo
	NodeLinkRepo     models.NodeLinkRepo
	RedisClient      *rc.ClusterClient
}

// INode node interface
type INode interface {
	Init(ctx context.Context, tx *gorm.DB, req *InitNodeReq, initParam *pb.NodeEventRespData) error
	Complete(ctx context.Context, tx *gorm.DB, req *CompleteNodeReq) (bool, error)
}

// CreateNodeInstance init component instance
func (n *Node) CreateNodeInstance(tx *gorm.DB, req *InitNodeInstanceReq) (*models.NodeInstance, error) {
	ni := &models.NodeInstance{
		ProcID:         req.Instance.ProcID,
		ProcInstanceID: req.Instance.ID,
		PID:            req.Execution.NodeInstanceID,
		ExecutionID:    req.Execution.ID,
		NodeDefKey:     req.Node.DefKey,
		NodeName:       req.Node.Name,
		NodeType:       req.Node.NodeType,
		TaskID:         req.TaskID,
		CreatorID:      req.UserID,
	}

	err := n.NodeInstanceRepo.Create(tx, ni)
	if err != nil {
		return nil, err
	}

	return ni, nil
}

// CreateExecution create execution
func (n *Node) CreateExecution(tx *gorm.DB, req *InitNodeReq) (*models.Execution, error) {
	entity := &models.Execution{
		ProcID:         req.Instance.ProcID,
		ProcInstanceID: req.Instance.ID,
		PID:            "",
		IsActive:       1,
		CreatorID:      req.UserID,
	}
	err := n.ExecutionRepo.Create(tx, entity)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// UpdateExecution update execution
func (n *Node) UpdateExecution(tx *gorm.DB, req *CompleteNodeReq, nodeInstanceID string) (*models.Execution, error) {
	entity := map[string]interface{}{
		"node_def_key":     req.Node.DefKey,
		"node_instance_id": nodeInstanceID,
		"is_active":        1,
		"modifier_id":      req.UserID,
	}

	return n.ExecutionRepo.Update(tx, req.Execution.ID, entity)
}

// InitNextNodes InitNextNodes
func (n *Node) InitNextNodes(ctx context.Context, tx *gorm.DB, req *InitNodeReq) (*pb.NodeEventRespData, error) {
	nextNodeDefKeys, err := n.NodeLinkRepo.FindNextNodesByNodeID(tx, req.Instance.ProcID, req.Node.ID)
	if err != nil {
		return nil, err
	}
	if req.NextNodes != "" {
		nextNodeDefKeys = strings.Split(req.NextNodes, ",")
	}
	// init next node
	for i := 0; i < len(nextNodeDefKeys); i++ {
		d := &pb.NodeEventReqData{
			ProcessInstanceID: req.Instance.ID,
			ProcessID:         req.Instance.ProcID,
			// "userID":            req.UserID,
			NodeDefKey:  nextNodeDefKeys[i],
			ExecutionID: req.Execution.ID,
		}

		nodeInitBeginResp, err := n.PublishMessage(ctx, internal.SynchronizationMode, internal.NodeInitBeginEvent, d)
		if err != nil {
			return nil, err
		}
		node, err := n.NodeRepo.FindByDefKey(tx, req.Instance.ProcID, nextNodeDefKeys[i])
		if err != nil {
			return nil, err
		}
		cNode, err := NodeFactory(node.NodeType)
		if err != nil {
			return nil, err
		}
		initNodeReq := &InitNodeReq{
			Execution: req.Execution,
			Instance:  req.Instance,
			Node:      node,
			UserID:    req.UserID,
			Params:    req.Params,
		}
		err = cNode.Init(ctx, tx, initNodeReq, nodeInitBeginResp)
		if err != nil {
			return nil, err
		}
		d.TaskID = strings.Split(initNodeReq.InitResp.EventTaskID, ",")
		return n.PublishMessage(ctx, internal.AsynchronousMode, internal.NodeInitEndEvent, d)
	}
	return nil, nil
}

// PublishMessage PublishMessage
func (n *Node) PublishMessage(ctx context.Context, mode, name string, data *pb.NodeEventReqData) (*pb.NodeEventRespData, error) {
	ms := &listener.EventMessage{
		EventName: name,
		EventData: data,
		Message: &listener.Message{
			MessageType:     "eventMessage",
			MessageSendMode: mode,
		},
	}

	return n.Listener.Notify(ctx, ms)
}

// UserTaskIdentity UserTaskIdentity
func (n *Node) UserTaskIdentity(tx *gorm.DB, instanceID, taskID, userID, createID string) error {
	if userID == "" {
		return nil
	}
	ti := &models.TaskIdentity{
		ID:           id2.GenID(),
		TaskID:       taskID,
		CreateTime:   time2.Now(),
		InstanceID:   instanceID,
		CreatorID:    createID,
		UserID:       userID,
		IdentityType: internal.IdentityUser,
	}
	if err := n.TaskIdentity.Create(tx, ti); err != nil {
		return err
	}
	return nil
}

// InitNodeReq InitNodeReq
type InitNodeReq struct {
	Task      *models.Task
	Execution *models.Execution
	Instance  *models.Instance
	Node      *models.Node
	TaskType  string
	NextNodes string
	Assignee  string
	UserID    string
	Name      string
	Desc      string
	Params    map[string]interface{}
	InitResp
}

// InitResp sometime return init response
type InitResp struct {
	Tasks       *models.Task
	EventTaskID string
}

// CompleteNodeReq CompleteNodeReq
type CompleteNodeReq struct {
	Task      *models.Task
	Execution *models.Execution
	Instance  *models.Instance
	Node      *models.Node
	NextNodes string
	UserID    string
	Params    map[string]interface{}
	Comments  string
}

// InitNodeInstanceReq InitNodeInstanceReq
type InitNodeInstanceReq struct {
	Execution *models.Execution
	Instance  *models.Instance
	Node      *models.Node
	Assignee  string
	TaskID    string
	UserID    string
}

func (n *Node) getVariablesFromRedis(ctx context.Context, instanceID, nodeID, name string) ([]string, error) {
	key := internal.RedisPreKey + instanceID + ":" + nodeID + ":" + name
	entityByte, err := n.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			return nil, nil
		}
		return nil, err
	}
	vv := make([]string, 0)
	entity := new(models.Variables)
	err = json.Unmarshal(entityByte, entity)
	if entity.ID != "" {
		err := json.Unmarshal([]byte(entity.ComplexValue.String()), &vv)
		if err != nil {
			return nil, err
		}
	}
	return vv, err
}

// FindChildExecution FindChildExecution
func (n *Node) FindChildExecution(tx *gorm.DB, instanceID, executionID string, res *[]*models.Execution) error {
	exs, err := n.ExecutionRepo.FindByPID(tx, instanceID, executionID, 1)
	if err != nil {
		return err
	}
	if len(exs) > 0 {
		for _, ex := range exs {
			*res = append(*res, ex)
			err := n.FindChildExecution(tx, ex.ProcInstanceID, ex.ID, res)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
