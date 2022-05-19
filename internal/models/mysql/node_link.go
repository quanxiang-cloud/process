package mysql

import (
	"git.internal.yunify.com/qxp/process/internal/models"
	"gorm.io/gorm"
)

type nodeLinkRepo struct{}

func (nl *nodeLinkRepo) TableName() string {
	return "proc_node_link"
}

// NewNodeLinkRepo new
func NewNodeLinkRepo() models.NodeLinkRepo {
	return &nodeLinkRepo{}
}

// Create create component link
func (nl *nodeLinkRepo) Create(db *gorm.DB, entity *models.NodeLink) error {
	err := db.Table(nl.TableName()).
		Create(entity).
		Error
	return err
}

// FindByNodeID find component link by ID
func (nl *nodeLinkRepo) FindByNodeID(db *gorm.DB, processID, nodeID string) ([]*models.NodeLink, error) {
	nls := make([]*models.NodeLink, 0)
	err := db.Table(nl.TableName()).
		Where(map[string]interface{}{
			// "proc_id": processID,
			"node_id": nodeID,
		}).
		Find(&nls).
		Error
	if err != nil {
		return nil, err
	}
	return nls, nil
}

// FindNextByNodeID find next component str by component ID
func (nl *nodeLinkRepo) FindNextByNodeID(db *gorm.DB, processID, nodeID string) (string, error) {
	nodes, err := nl.FindByNodeID(db, processID, nodeID)
	if err != nil {
		return "", err
	}

	if nodes != nil {
		nodeStr := ""
		for _, value := range nodes {
			if nodeStr == "" {
				nodeStr = value.NextNodeDefKey
			} else {
				nodeStr += "," + value.NextNodeDefKey
			}

		}
		return nodeStr, nil
	}
	return "", nil
}

// FindNextNodesByNodeID FindNextNodesByNodeID
func (nl *nodeLinkRepo) FindNextNodesByNodeID(db *gorm.DB, processID, nodeID string) ([]string, error) {
	nls := make([]*models.NodeLink, 0)
	err := db.Table(nl.TableName()).
		Where(map[string]interface{}{
			"proc_id": processID,
			"node_id": nodeID,
		}).
		Find(&nls).
		Error
	if err != nil {
		return nil, err
	}
	if len(nls) == 0 {
		return nil, nil
	}

	var nodeDefKeys []string
	for _, value := range nls {
		nodeDefKeys = append(nodeDefKeys, value.NextNodeDefKey)
	}
	return nodeDefKeys, nil
}

func (nl *nodeLinkRepo) FindByNodeDefKey(db *gorm.DB, processID, nodeDefKey string) ([]*models.NodeLink, error) {
	nls := make([]*models.NodeLink, 0)
	err := db.Table(nl.TableName()).
		Where(map[string]interface{}{
			"proc_id":           processID,
			"next_node_def_key": nodeDefKey,
		}).
		Find(&nls).
		Error
	if err != nil {
		return nil, err
	}
	return nls, nil
}
