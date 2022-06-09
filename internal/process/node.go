package process

import (
	"errors"
	"github.com/quanxiang-cloud/process/internal/models"
	"github.com/quanxiang-cloud/process/internal/models/mysql"
	"github.com/quanxiang-cloud/process/internal/server/options"
	"github.com/quanxiang-cloud/process/pkg"
	"github.com/quanxiang-cloud/process/pkg/config"
	"github.com/quanxiang-cloud/process/pkg/misc/id2"
	"github.com/quanxiang-cloud/process/pkg/misc/logger"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	"gorm.io/gorm"
)

// Node service
type Node interface {
	AddNodes(db *gorm.DB, model *models.Model, nodes []*NodeData) ([]*models.Node, error)
	AddNode(db *gorm.DB, req *AddNodeReq) (*models.Node, error)
}

type node struct {
	nodeRepo         models.NodeRepo
	nodeLinkRepo     models.NodeLinkRepo
	identityLinkRepo models.IdentityLinkRepo
}

// NewNode init
func NewNode(conf *config.Configs, opts ...options.Options) (Node, error) {
	n := &node{
		nodeRepo:         mysql.NewNodeRepo(),
		nodeLinkRepo:     mysql.NewNodeLinkRepo(),
		identityLinkRepo: mysql.NewIdentityLinkRepo(),
	}
	return n, nil
}

// AddNodes add nodes
func (n *node) AddNodes(db *gorm.DB, model *models.Model, nodes []*NodeData) ([]*models.Node, error) {
	if nodes == nil || len(nodes) == 0 {
		err := errors.New("component datas is null")
		logger.Logger.Errorw(err.Error())
		return nil, err
	}
	res := make([]*models.Node, 0)
	for _, node := range nodes {
		// Insert component data
		nodeEntity := models.Node{
			ID:             id2.GenID(),
			ProcID:         model.ID,
			Name:           node.Name,
			DefKey:         node.DefKey,
			NodeType:       node.Type,
			ProcInstanceID: node.InstanceID,
			// SubProcID string `json:"subProcId"`
			Desc:       node.Desc,
			CreatorID:  model.CreatorID,
			ModifierID: model.CreatorID,
			CreateTime: time2.Now(),
			ModifyTime: time2.Now(),
			TenantID:   model.TenantID,
		}
		err := n.nodeRepo.Create(db, &nodeEntity)
		if err != nil {
			return nil, err
		}

		// Insert component link data
		err = n.addNodeLink(db, &nodeEntity, node)
		if err != nil {
			return nil, err
		}

		// Insert identity link data
		err = n.addIdentityLink(db, &nodeEntity, node)
		if err != nil {
			return nil, err
		}
		res = append(res, &nodeEntity)
	}

	return res, nil
}

// addNodeLink add component link data
func (n *node) addNodeLink(db *gorm.DB, nodeEntity *models.Node, nodeData *NodeData) error {
	if len(nodeData.NextNodes) > 0 {
		for _, nextNode := range nodeData.NextNodes {
			nodeLinkEntity := models.NodeLink{
				ID:             id2.GenID(),
				ProcID:         nodeEntity.ProcID,
				NodeID:         nodeEntity.ID,
				NextNodeDefKey: nextNode.NodeID,
				Condition:      nextNode.Condition,
				CreatorID:      nodeEntity.CreatorID,
				ModifierID:     nodeEntity.CreatorID,
				CreateTime:     time2.Now(),
				ModifyTime:     time2.Now(),
				TenantID:       nodeEntity.TenantID,
			}
			err := n.nodeLinkRepo.Create(db, &nodeLinkEntity)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// addIdentityLink add component identity link data
func (n *node) addIdentityLink(db *gorm.DB, nodeEntity *models.Node, nodeData *NodeData) error {
	if len(nodeData.UserIDs) > 0 {
		for _, userID := range nodeData.UserIDs {
			identityLinkEntity := models.IdentityLink{
				ID:           id2.GenID(),
				NodeID:       nodeEntity.ID,
				IdentityType: "USER",
				UserID:       userID,
				CreatorID:    nodeEntity.CreatorID,
				ModifierID:   nodeEntity.CreatorID,
				CreateTime:   time2.Now(),
				ModifyTime:   time2.Now(),
				TenantID:     nodeEntity.TenantID,
			}

			err := n.identityLinkRepo.Create(db, &identityLinkEntity)
			if err != nil {
				return err
			}
		}
	}

	if len(nodeData.GroupIDs) > 0 {
		for _, groupID := range nodeData.GroupIDs {
			identityLinkEntity := models.IdentityLink{
				ID:           id2.GenID(),
				NodeID:       nodeEntity.ID,
				IdentityType: "GROUP",
				GroupID:      groupID,
				// GroupID   string `json:"groupId"`
				// Variable  string `json:"variable"`
				CreatorID:  nodeEntity.CreatorID,
				ModifierID: nodeEntity.CreatorID,
				CreateTime: time2.Now(),
				ModifyTime: time2.Now(),
				TenantID:   nodeEntity.TenantID,
			}

			err := n.identityLinkRepo.Create(db, &identityLinkEntity)
			if err != nil {
				return err
			}
		}
	}

	if len(nodeData.Variable) > 0 {
		identityLinkEntity := models.IdentityLink{
			ID:           id2.GenID(),
			NodeID:       nodeEntity.ID,
			IdentityType: "VARIABLE",
			Variable:     nodeData.Variable,
			CreatorID:    nodeEntity.CreatorID,
			ModifierID:   nodeEntity.CreatorID,
			CreateTime:   time2.Now(),
			ModifyTime:   time2.Now(),
			TenantID:     nodeEntity.TenantID,
		}

		err := n.identityLinkRepo.Create(db, &identityLinkEntity)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddNode insert component data
func (n *node) AddNode(db *gorm.DB, req *AddNodeReq) (*models.Node, error) {
	nodeEntity := models.Node{}
	pkg.CopyProperties(&nodeEntity, req)

	nodeEntity.ID = id2.GenID()
	nodeEntity.ModifierID = req.CreatorID
	nodeEntity.CreateTime = time2.Now()
	nodeEntity.ModifyTime = time2.Now()

	err := n.nodeRepo.Create(db, &nodeEntity)
	if err != nil {
		return nil, err
	}

	return &nodeEntity, nil
}

// AddNodeReq add component request params
type AddNodeReq struct {
	ProcID    string `json:"procId"`
	Name      string `json:"name"`
	DefKey    string `json:"defKey"`
	Type      string `json:"type"`
	SubProcID string `json:"subProcId"`
	Desc      string `json:"desc"`
	CreatorID string `json:"creatorId"`
	TenantID  string `json:"tenantId"`
}
