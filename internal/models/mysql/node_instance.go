package mysql

import (
	"github.com/quanxiang-cloud/process/internal/models"
	"github.com/quanxiang-cloud/process/pkg/misc/id2"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	"gorm.io/gorm"
)

type nodeInstanceRepo struct{}

func (ni *nodeInstanceRepo) TableName() string {
	return "proc_node_instance"
}

// NewNodeInstanceRepo new
func NewNodeInstanceRepo() models.NodeInstanceRepo {
	return &nodeInstanceRepo{}
}

func (ni *nodeInstanceRepo) Create(db *gorm.DB, entity *models.NodeInstance) error {
	entity.ID = id2.GenID()
	entity.CreateTime = time2.Now()
	err := db.Table(ni.TableName()).
		Create(entity).
		Error
	return err
}

func (ni *nodeInstanceRepo) FindByID(db *gorm.DB, id string) (*models.NodeInstance, error) {
	modal := new(models.NodeInstance)
	err := db.Table(ni.TableName()).
		Where("id = ?", id).
		Find(modal).
		Error
	if err != nil {
		return nil, err
	}
	if modal.ID == "" {
		return nil, nil
	}
	return modal, err
}

func (ni *nodeInstanceRepo) FindByTaskID(db *gorm.DB, id string) (*models.NodeInstance, error) {
	modal := new(models.NodeInstance)
	err := db.Table(ni.TableName()).
		Where("task_id = ?", id).
		Find(modal).
		Error
	if err != nil {
		return nil, err
	}
	if modal.ID == "" {
		return nil, nil
	}
	return modal, err
}
