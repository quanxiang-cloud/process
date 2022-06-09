package mysql

import (
	"github.com/quanxiang-cloud/process/internal/models"
	"gorm.io/gorm"
)

type identityLinkRepo struct{}

func (il *identityLinkRepo) TableName() string {
	return "proc_identity_link"
}

// NewIdentityLinkRepo new
func NewIdentityLinkRepo() models.IdentityLinkRepo {
	return &identityLinkRepo{}
}

func (il *identityLinkRepo) Create(db *gorm.DB, entity *models.IdentityLink) error {
	err := db.Table(il.TableName()).
		Create(entity).
		Error
	return err
}

// QueryByNodeID query identity links by nodeID
func (il *identityLinkRepo) QueryByNodeID(db *gorm.DB, nodeID string) ([]*models.IdentityLink, error) {
	ils := make([]*models.IdentityLink, 0)
	err := db.Table(il.TableName()).
		Where("node_id = ?", nodeID).
		Find(&ils).
		Error

	if err != nil {
		return nil, err
	}
	return ils, nil
}
