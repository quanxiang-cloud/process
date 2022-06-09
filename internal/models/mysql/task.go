package mysql

import (
	"fmt"
	"github.com/quanxiang-cloud/process/internal"
	"github.com/quanxiang-cloud/process/internal/models"
	"github.com/quanxiang-cloud/process/pkg/misc/id2"
	"github.com/quanxiang-cloud/process/pkg/misc/time2"
	page2 "github.com/quanxiang-cloud/process/pkg/page"
	"gorm.io/gorm"
)

type taskRepo struct{}

func (t *taskRepo) TableName() string {
	return "proc_task"
}

// NewTaskRepo new
func NewTaskRepo() models.TaskRepo {
	return &taskRepo{}
}

// Create create task
func (t *taskRepo) Create(db *gorm.DB, entity *models.Task) error {
	entity.ID = id2.GenID()
	entity.CreateTime = time2.Now()
	err := db.Table(t.TableName()).
		Create(entity).
		Error
	return err
}

// Update update task
func (t *taskRepo) Update(db *gorm.DB, entity *models.Task) error {
	entity.ModifyTime = time2.Now()
	err := db.Table(t.TableName()).Where("id=?", entity.ID).Updates(entity).Error
	return err
}

// FindByID find task by ID
func (t *taskRepo) FindByID(db *gorm.DB, id string) (*models.Task, error) {
	modal := new(models.Task)
	err := db.Table(t.TableName()).
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

// DeleteByID delete by ID
func (t *taskRepo) DeleteByID(db *gorm.DB, id string) error {
	err := db.Table(t.TableName()).
		Where("id = ?", id).
		Delete(&models.Task{}).
		Error
	return err
}

func (t *taskRepo) DeleteByInstanceID(db *gorm.DB, instanceID string) error {
	err := db.Table(t.TableName()).
		Where("proc_instance_id = ?", instanceID).
		Delete(&models.Task{}).
		Error
	return err
}

// FindByIDs find task by IDs
func (t *taskRepo) FindByIDs(db *gorm.DB, ids []string) ([]*models.Task, error) {
	rest := make([]*models.Task, 0)
	err := db.Table(t.TableName()).
		Where("id in (?)", ids).
		Find(&rest).
		Error
	if err != nil {
		return nil, err
	}
	return rest, nil
}

// FindByExecutionIDs find task by IDs
func (t *taskRepo) FindByExecutionIDs(db *gorm.DB, ids []string) ([]*models.Task, error) {
	rest := make([]*models.Task, 0)
	err := db.Table(t.TableName()).
		Where("execution_id in (?)", ids).
		Find(&rest).
		Error
	if err != nil {
		return nil, err
	}
	return rest, nil
}

func (t *taskRepo) FindByInstanceID(db *gorm.DB, instanceID string) ([]*models.Task, error) {
	rest := make([]*models.Task, 0)
	err := db.Table(t.TableName()).
		Where("proc_instance_id = ?", instanceID).
		Find(&rest).
		Error
	if err != nil {
		return nil, err
	}
	return rest, nil
}

func (t *taskRepo) FindPageByInstanceID(db *gorm.DB, page, limit int, req *models.QueryTaskCondition) ([]*models.Task, int64) {
	db = db.Table(t.TableName()).
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
	if len(req.Order) > 0 {
		for _, od := range req.Order {
			db = db.Order(fmt.Sprintf("%s%s%s", od.Column, " ", od.OrderType))
		}
	}
	res := make([]*models.Task, 0)
	var num int64
	db.Model(&models.Task{}).Count(&num)
	newPage := page2.NewPage(page, limit, num)
	db = db.Limit(newPage.PageSize).Offset(newPage.StartIndex)

	affected := db.Find(&res).RowsAffected
	if affected > 0 {
		return res, num
	}
	return nil, 0
}

func (t *taskRepo) FindPageByCondition(db *gorm.DB, page, limit int, req *models.QueryTaskCondition) ([]*models.TaskVO, int64) {
	db1 := db.Table(t.TableName())
	if req.UserID != "" && req.GroupID != "" {
		db1 = db1.Where("proc_task.id in (select distinct task_id from proc_task_identity where user_id = ? or group_id = ?)", req.UserID, req.GroupID)
	} else {
		db1 = db1.Where("proc_task.id in (select distinct task_id from proc_task_identity )")
	}
	if req.NodeDefKey != "" {
		db1 = db1.Where("proc_task.node_def_key = ?", req.NodeDefKey)
	}
	if len(req.InstanceID) > 0 {
		db1 = db1.Where("proc_task.proc_instance_id in (?)", req.InstanceID)
	}
	if len(req.ProcessID) > 0 {
		db1 = db1.Where("proc_task.proc_id in (?)", req.ProcessID)
	}
	if len(req.TaskID) > 0 {
		db1 = db1.Where("proc_task.id in (?)", req.TaskID)
	}
	if req.Name != "" {
		db1 = db1.Where("proc_task.name like ?", "%"+req.Name+"%")
	}
	if req.Des != "" {
		db1 = db1.Where("proc_task.desc REGEXP ?", req.Des)
	}
	if len(req.Order) > 0 {
		for _, od := range req.Order {
			db1 = db1.Order(fmt.Sprintf("%s%s%s%s", "proc_task.", od.Column, " ", od.OrderType))
		}
	} else {
		db1 = db1.Order("proc_task.create_time desc")
	}
	if req.DueTime != "" {
		db1 = db1.Where("proc_task.due_time < ? and proc_task.due_time != ''", req.DueTime)
	}
	var num int64
	db1.Model(&models.Task{}).
		Joins("left join proc_instance on proc_task.proc_instance_id = proc_instance.id").
		Where("proc_instance.app_status != ?", internal.Suspend).
		Count(&num)
	newPage := page2.NewPage(page, limit, num)
	db1 = db1.Limit(newPage.PageSize).Offset(newPage.StartIndex)

	db1.Select("proc_node_instance.id as node_instance_id,proc_node_instance.p_id as node_instance_pid,proc_task.*").
		Joins("left join proc_node_instance on proc_node_instance.task_id = proc_task.id")

	var res []*models.TaskVO
	db1.Scan(&res)
	return res, num
}

func (t *taskRepo) FindAllPageByCondition(db *gorm.DB, page, limit int, req *models.QueryTaskCondition) ([]*models.Task, int64) {
	db1 := db.Table(t.TableName())
	if req.UserID != "" && req.GroupID != "" {
		db1 = db1.Where("proc_task.id in (select distinct task_id from proc_task_identity where user_id = ? or group_id = ?)", req.UserID, req.GroupID)
	} else {
		db1 = db1.Where("proc_task.id in (select distinct task_id from proc_task_identity )")
	}
	if req.NodeDefKey != "" {
		db1 = db1.Where("proc_task.node_def_key = ?", req.NodeDefKey)
	}
	if len(req.InstanceID) > 0 {
		db1 = db1.Where("proc_task.proc_instance_id in (?)", req.InstanceID)
	}
	if len(req.ProcessID) > 0 {
		db1 = db1.Where("proc_task.proc_id in (?)", req.ProcessID)
	}
	if len(req.TaskID) > 0 {
		db1 = db1.Where("proc_task.id in (?)", req.TaskID)
	}
	if req.Name != "" {
		db1 = db1.Where("`name` like ?", "%"+req.Name+"%")
	}
	if req.Des != "" {
		db1 = db1.Where("`desc` REGEXP ?", req.Des)
	}
	if req.DueTime != "" {
		db1 = db1.Where("proc_task.due_time < ? and proc_task.due_time != ''", req.DueTime)
	}
	if req.Status != "" {
		db1 = db1.Where("`proc_task.status` = ?", req.Status)
	}

	db2 := db.Table("proc_history_task").
		Where("proc_history_task.status != ?", internal.Deleted)
	if req.UserID != "" {
		db2 = db2.Where("proc_history_task.assignee = ?", req.UserID)
	}
	if len(req.InstanceID) > 0 {
		db2 = db2.Where("proc_history_task.proc_instance_id in (?)", req.InstanceID)
	}
	if len(req.ProcessID) > 0 {
		db2 = db2.Where("proc_history_task.proc_id in (?)", req.ProcessID)
	}
	if len(req.TaskID) > 0 {
		db2 = db2.Where("proc_history_task.id in (?)", req.TaskID)
	}
	if req.Name != "" {
		db2 = db2.Where("`name` like ?", "%"+req.Name+"%")
	}
	if req.Des != "" {
		db2 = db2.Where("`desc` REGEXP ?", req.Des)
	}
	if req.Status != "" {
		db2 = db2.Where("`proc_history_task.status` = ?", req.Status)
	}
	orderBy := "create_time desc"
	if len(req.Order) > 0 {
		odb := ""
		for _, od := range req.Order {
			if odb != "" {
				odb = fmt.Sprintf("%s%s", odb, ",")
			}
			odb = fmt.Sprintf("%s%s%s", od.Column, " ", od.OrderType)
		}
		orderBy = odb
	}
	var num, num1, num2 int64
	db1.Model(&models.Task{}).
		Joins("left join proc_instance on proc_task.proc_instance_id = proc_instance.id").
		Where("proc_instance.app_status != ?", internal.Suspend).
		Count(&num1)
	db2.Model(&models.HistoryTask{}).
		Joins("left join proc_instance on proc_history_task.proc_instance_id = proc_instance.id").
		Where("proc_instance.app_status != ?", internal.Suspend).
		Count(&num2)
	num = num1 + num2
	newPage := page2.NewPage(page, limit, num)
	query := fmt.Sprintf("%s%s%s", `? UNION ? order by `, orderBy, ` limit ?,?`)
	select1 := fmt.Sprintf("%s%s%s", "proc_task.id,proc_task.proc_id,proc_task.proc_instance_id,proc_task.execution_id,proc_task.node_id,proc_task.node_def_key,",
		"proc_task.next_node_def_key,proc_task.name,proc_task.desc,proc_task.assignee,proc_task.task_type,proc_task.status,proc_task.end_time,",
		"proc_task.due_time,proc_task.creator_id,proc_task.create_time,proc_task.modifier_id,proc_task.modify_time")
	select2 := fmt.Sprintf("%s%s%s", "proc_history_task.id,proc_history_task.proc_id,proc_history_task.proc_instance_id,proc_history_task.execution_id,proc_history_task.node_id,proc_history_task.node_def_key,",
		"proc_history_task.next_node_def_key,proc_history_task.name,proc_history_task.desc,proc_history_task.assignee,proc_history_task.task_type,proc_history_task.status,proc_history_task.end_time,",
		"proc_history_task.due_time,proc_history_task.creator_id,proc_history_task.create_time,proc_history_task.modifier_id,proc_history_task.modify_time")
	db = db.Raw(query,
		db1.Select(select1).Model(&models.Task{}),
		db2.Select(select2).Model(&models.HistoryTask{}),
		newPage.StartIndex,
		newPage.PageSize,
	)
	var res []*models.Task
	db.Scan(&res)
	return res, num
}
