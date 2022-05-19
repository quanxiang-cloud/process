package mysql

import (
	"git.internal.yunify.com/qxp/misc/id2"
	"git.internal.yunify.com/qxp/misc/time2"
	"git.internal.yunify.com/qxp/process/internal/models"
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

func (ni *nodeInstanceRepo) FindByInstanceID(db *gorm.DB, instanceID string) ([]*models.NodeInstance, error) {
	datas := make([]*models.NodeInstance, 0)
	err := db.Table(ni.TableName()).
		Where("proc_instance_id = ?", instanceID).
		Find(&datas).
		Error
	if err != nil {
		return nil, err
	}

	return datas, err
}
