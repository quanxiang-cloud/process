package mysql

import (
	"fmt"
	"git.internal.yunify.com/qxp/process/internal"
	"git.internal.yunify.com/qxp/process/internal/models"
	page2 "git.internal.yunify.com/qxp/process/pkg/page"
	"gorm.io/gorm"
)

type historyTaskRepo struct{}

func (ht *historyTaskRepo) TableName() string {
	return "proc_history_task"
}

// NewHistoryTaskRepo new
func NewHistoryTaskRepo() models.HistoryTaskRepo {
	return &historyTaskRepo{}
}

func (ht *historyTaskRepo) Create(db *gorm.DB, entity *models.HistoryTask) error {
	err := db.Table(ht.TableName()).
		Create(entity).
		Error
	return err
}

func (ht *historyTaskRepo) Update(db *gorm.DB, entity *models.HistoryTask) error {
	err := db.Table(ht.TableName()).
		// Where("id=?", entity.ID).
		Updates(entity).Error
	return err
}

func (ht *historyTaskRepo) FindByInstanceID(db *gorm.DB, page, limit int, req *models.QueryTaskCondition) ([]*models.HistoryTask, int64) {
	db = db.Table(ht.TableName()).
		Where("proc_instance_id in (?)", req.InstanceID)
	if req.NodeDefKey != "" {
		db = db.Where("node_def_key = ?", req.NodeDefKey)
	}
	if req.Name != "" {
		db = db.Where("`name` like ?", fmt.Sprintf("%s%s%s", "%", req.Name, "%"))
	}
	if req.Des != "" {
		db = db.Where("`desc` REGEXP ?", req.Des)
	}
	if req.Assignee != "" {
		db = db.Where("assignee = ?", req.Assignee)
	}
	if len(req.Order) > 0 {
		for _, od := range req.Order {
			db = db.Order(fmt.Sprintf("%s%s%s", od.Column, " ", od.OrderType))
		}
	}
	res := make([]*models.HistoryTask, 0)
	var num int64
	db.Model(&models.HistoryTask{}).Count(&num)
	newPage := page2.NewPage(page, limit, num)
	db = db.Limit(newPage.PageSize).Offset(newPage.StartIndex)

	affected := db.Find(&res).RowsAffected
	if affected > 0 {
		return res, num
	}
	return nil, 0
}

func (ht *historyTaskRepo) FindPreTask(db *gorm.DB, instanceID, executionID string) ([]*models.HistoryTask, error) {
	rest := make([]*models.HistoryTask, 0)
	err := db.Table(ht.TableName()).
		Where("proc_instance_id = ? and execution_id = ? and task_type = ?", instanceID, executionID, internal.ModelTask).
		Find(&rest).
		Error
	return rest, err
}

func (ht *historyTaskRepo) FindByIDs(db *gorm.DB, ids []string) ([]*models.HistoryTask, error) {
	rest := make([]*models.HistoryTask, 0)
	err := db.Table(ht.TableName()).
		Where("id in (?)", ids).
		Find(&rest).
		Error
	if err != nil {
		return nil, err
	}
	return rest, nil
}
