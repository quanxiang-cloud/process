package mysql

import (
	"github.com/quanxiang-cloud/process/internal/models"
	"gorm.io/gorm"
)

type modelRepo struct{}

func (m *modelRepo) TableName() string {
	return "proc_model"
}

// NewModelRepo new
func NewModelRepo() models.ModelRepo {
	return &modelRepo{}
}

func (m *modelRepo) Create(db *gorm.DB, entity *models.Model) error {
	err := db.Table(m.TableName()).
		Create(entity).
		Error
	return err
}

func (m *modelRepo) FindByID(db *gorm.DB, id string) (*models.Model, error) {
	modal := new(models.Model)
	err := db.Table(m.TableName()).
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

func (m *modelRepo) FindByDefKey(db *gorm.DB, defKey string) ([]*models.Model, error) {
	modals := make([]*models.Model, 0)
	err := db.Table(m.TableName()).
		Where("def_key = ?", defKey).
		Find(&modals).
		Error
	if err != nil {
		return nil, err
	}
	return modals, err
}
