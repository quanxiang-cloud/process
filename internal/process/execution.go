package process

import (
	"context"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/internal/models/mysql"
	"git.internal.yunify.com/qxp/process/pkg/config"
	"gorm.io/gorm"
)

// Execution service
type Execution interface {
	// AddExecution(ctx context.Context, req *AddExecutionReq) (*models.Execution, error)
	UpdateExecution(ctx context.Context, req *UpdateExecutionReq) (*models.Execution, error)
	SetActive(ctx context.Context, ID string, isActive int8) error
	DeleteByID(ctx context.Context, ID string) error
}

// NewExecution init
func NewExecution(conf *config.Configs) (Execution, error) {
	e := &execution{
		executionRepo: mysql.NewExecutionRepo(),
	}
	return e, nil
}

type execution struct {
	db            *gorm.DB
	executionRepo models.ExecutionRepo
}

// SetDB set db
func (e *execution) SetDB(db *gorm.DB) {
	e.db = db
}

// // AddExecution add execution
// func (e *execution) AddExecution(ctx context.Context, req *AddExecutionReq) (*models.Execution, error) {
// 	entity := models.Execution{
// 		ProcID:         req.ProcID,
// 		ProcInstanceID: req.ProcInstanceID,
// 		PID:            req.PID,
// 		NodeDefKey:     req.NodeDefKey,
// 		NodeInstanceID: req.NodeInstanceID,
// 		IsActive:       req.IsActive,
// 		CreatorID:      req.UserID,
// 		TenantID:       req.TenantID,
// 	}
//
// 	err := e.executionRepo.Create(e.db, &entity)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &entity, nil
// }

// AddExecution update execution
func (e *execution) UpdateExecution(ctx context.Context, req *UpdateExecutionReq) (*models.Execution, error) {
	entity := map[string]interface{}{
		"node_def_key":     req.NodeDefKey,
		"node_instance_id": req.NodeInstanceID,
		"is_active":        req.IsActive,
		"modifier_id":      req.UserID,
	}

	return e.executionRepo.Update(e.db, req.ID, entity)
}

// SetActive set active status
func (e *execution) SetActive(ctx context.Context, ID string, isActive int8) error {
	err := e.executionRepo.SetActive(e.db, ID, isActive)
	return err
}

// DeleteByID delete by ID
func (e *execution) DeleteByID(ctx context.Context, ID string) error {
	err := e.executionRepo.DeleteByID(e.db, ID)
	return err
}

// AddExecutionReq add execution request params
type AddExecutionReq struct {
	ProcID         string
	ProcInstanceID string
	PID            string
	NodeDefKey     string
	NodeInstanceID string
	IsActive       int8
	TenantID       string
	UserID         string
}

// UpdateExecutionReq update execution request params
type UpdateExecutionReq struct {
	ID             string
	NodeDefKey     string
	NodeInstanceID string
	IsActive       int8
	UserID         string
}
