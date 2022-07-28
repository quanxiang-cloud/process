package process

import (
	"context"
	"encoding/json"
	"github.com/quanxiang-cloud/process/internal/models"
	"github.com/quanxiang-cloud/process/internal/models/mysql"
	listener "github.com/quanxiang-cloud/process/internal/server/events"
	"github.com/quanxiang-cloud/process/internal/server/options"
	"github.com/quanxiang-cloud/process/pkg/config"
	"github.com/quanxiang-cloud/process/pkg/misc/id2"
	"github.com/quanxiang-cloud/process/pkg/misc/logger"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	"gorm.io/gorm"
)

// Process service
type Process interface {
	AddModel(ctx context.Context, req *AddModelReq) (*models.Model, error)
	FindProcessNode(ctx context.Context, req *QueryNodeReq) ([]*models.Node, error)
	SetDB(db *gorm.DB)
}

// NewProcess init
func NewProcess(conf *config.Configs, opts ...options.Options) (Process, error) {
	node, _ := NewNode(conf, opts...)
	p := &process{
		nodeRepo:  mysql.NewNodeRepo(),
		modelRepo: mysql.NewModelRepo(),
		node:      node,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

type process struct {
	db        *gorm.DB
	nodeRepo  models.NodeRepo
	modelRepo models.ModelRepo
	node      Node
	l         *listener.Listener
}

// SetDB set db
func (p *process) SetDB(db *gorm.DB) {
	p.db = db
}

func (p *process) SetListener(l *listener.Listener) {
	p.l = l
}

func (p *process) FindProcessNode(ctx context.Context, req *QueryNodeReq) ([]*models.Node, error) {
	return p.nodeRepo.FindByProcessID(p.db, req.ProcessID, req.NodeDefKey)
}

// AddModel add process model
func (p *process) AddModel(ctx context.Context, req *AddModelReq) (*models.Model, error) {
	modelData := &ModelData{}
	err := json.Unmarshal([]byte(req.Model), modelData)
	if err != nil {
		logger.Logger.Errorw(err.Error(), logger.STDRequestID(ctx))
		return nil, err
	}

	// Insert model data
	modelEntity := models.Model{
		ID:         id2.GenID(),
		Name:       modelData.Name,
		DefKey:     modelData.DefKey,
		CreatorID:  req.CreatorID,
		ModifierID: req.CreatorID,
		CreateTime: time2.Now(),
		ModifyTime: time2.Now(),
		TenantID:   req.TenantID,
	}

	tx := p.db.Begin()

	err = p.modelRepo.Create(tx, &modelEntity)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = p.node.AddNodes(tx, &modelEntity, modelData.Nodes)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &modelEntity, nil
}

// QueryNodeReq query node req
type QueryNodeReq struct {
	ProcessID  string `json:"processID" binding:"required"`
	NodeDefKey string `json:"nodeDefKey"`
}

// AddModelReq add model request params
type AddModelReq struct {
	Model     string `json:"model" binding:"required"`
	CreatorID string `json:"creatorId"`
	TenantID  string `json:"tenantId"`
}

// ModelData model json data
type ModelData struct {
	DefKey string      `json:"id"`
	Name   string      `json:"name"`
	Nodes  []*NodeData `json:"nodes"`
}

// NodeData component json data
type NodeData struct {
	DefKey     string          `json:"id" binding:"required"`
	Name       string          `json:"name"`
	Type       string          `json:"type" binding:"required"`
	SubModel   *ModelData      `json:"subModel"`
	Desc       string          `json:"desc"`
	UserIDs    []string        `json:"userIds"`
	GroupIDs   []string        `json:"groupIds"`
	Variable   string          `json:"variable"`
	NextNodes  []*NodeLinkData `json:"nextNodes"`
	InstanceID string          `json:"-"`
}

// NodeLinkData component link json data
type NodeLinkData struct {
	NodeID    string `json:"id"`
	Condition string `json:"condition"`
}

// InitNextNodeReq InitNextNodeReq
type InitNextNodeReq struct {
	ProcessID   string                 `json:"processID" binding:"required"`
	InstanceID  string                 `json:"instanceID" binding:"required"`
	NodeDefKey  string                 `json:"nodeDefKey" binding:"required"`
	ExecutionID string                 `json:"executionID" binding:"required"`
	NextNodes   string                 `json:"nextNodes"`
	Params      map[string]interface{} `json:"params"`
	UserID      string                 `json:"userID"`
}

// NodeInstanceListReq req
type NodeInstanceListReq struct {
	ProcInstanceID string `json:"procInstanceId"`
}
