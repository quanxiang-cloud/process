package mysql

import (
	"github.com/quanxiang-cloud/process/internal/models"
	"gorm.io/gorm"
)

type nodeRepo struct{}

func (n *nodeRepo) TableName() string {
	return "proc_node"
}

// NewNodeRepo new
func NewNodeRepo() models.NodeRepo {
	return &nodeRepo{}
}

func (n *nodeRepo) Create(db *gorm.DB, entity *models.Node) error {
	err := db.Table(n.TableName()).
		Create(entity).
		Error
	return err
}

func (n *nodeRepo) FindStartNode(db *gorm.DB, processID string) (*models.Node, error) {
	node := new(models.Node)
	err := db.Table(n.TableName()).
		Where(map[string]interface{}{
			"proc_id":     processID,
			"node_type":   "Start",
			"sub_proc_id": "",
		}).
		Find(node).
		Error
	if err != nil {
		return nil, err
	}
	if node.ID == "" {
		return nil, nil
	}
	return node, nil
}

func (n *nodeRepo) FindByID(db *gorm.DB, id string) (*models.Node, error) {
	node := new(models.Node)
	err := db.Table(n.TableName()).
		Where("id = ?", id).
		Find(node).
		Error
	if err != nil {
		return nil, err
	}
	if node.ID == "" {
		return nil, nil
	}
	return node, nil
}

func (n *nodeRepo) FindByDefKey(db *gorm.DB, processID, defKey string) (*models.Node, error) {
	node := new(models.Node)
	err := db.Table(n.TableName()).
		Where("proc_id = ? and def_key = ?", processID, defKey).
		Find(node).
		Error
	if err != nil {
		return nil, err
	}
	if node.ID == "" {
		return nil, nil
	}
	return node, nil
}

func (n *nodeRepo) FindInstanceNode(db *gorm.DB, instanceID, defKey string) (*models.Node, error) {
	node := new(models.Node)
	err := db.Table(n.TableName()).
		Where("proc_instance_id = ? and def_key = ?", instanceID, defKey).
		Find(node).
		Error
	if err != nil {
		return nil, err
	}
	if node.ID == "" {
		return nil, nil
	}
	return node, nil
}

func (n *nodeRepo) FindByProcessID(db *gorm.DB, processID, nodeDefKey string) ([]*models.Node, error) {
	node := make([]*models.Node, 0)
	db = db.Table(n.TableName()).
		Where("proc_id = ?", processID)
	if nodeDefKey != "" {
		db = db.Where("def_key = ?", nodeDefKey)
	}
	err := db.Find(&node).
		Error
	if err != nil {
		return nil, err
	}
	return node, nil
}
