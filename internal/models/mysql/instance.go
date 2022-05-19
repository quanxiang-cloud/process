package mysql

import (
	"fmt"
	"git.internal.yunify.com/qxp/misc/time2"
	"git.internal.yunify.com/qxp/process/internal"
	"git.internal.yunify.com/qxp/process/internal/models"
	page2 "git.internal.yunify.com/qxp/process/pkg/page"
	"gorm.io/gorm"
)

type instanceRepo struct{}

func (i *instanceRepo) TableName() string {
	return "proc_instance"
}

// NewInstanceRepo new
func NewInstanceRepo() models.InstanceRepo {
	return &instanceRepo{}
}

func (i *instanceRepo) Create(db *gorm.DB, entity *models.Instance) error {
	entity.AppStatus = internal.Active
	err := db.Table(i.TableName()).
		Create(entity).
		Error
	return err
}

func (i *instanceRepo) FindByID(db *gorm.DB, id string) (*models.Instance, error) {
	modal := new(models.Instance)
	err := db.Table(i.TableName()).
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

func (i *instanceRepo) UpdateAppByProcDefKey(db *gorm.DB, procDefKey []string, appStatus string) error {
	var rest interface{}
	sql := `update proc_instance i inner join proc_model m on i.proc_id = m.id set i.app_status = ? where m.def_key in (?) `
	db = db.Raw(sql, appStatus, procDefKey)
	db.Scan(&rest)
	return nil
}

func (i *instanceRepo) DeleteAppByProcDefKey(db *gorm.DB, procDefKey []string) error {
	var rest interface{}
	sql := `delete proc_instance,proc_model from proc_instance inner join proc_model on proc_instance.proc_id = proc_model.id where proc_model.def_key in (?) `
	db = db.Raw(sql, procDefKey)
	db.Scan(&rest)
	return nil
}

func (i *instanceRepo) DeleteByID(db *gorm.DB, id string) error {
	err := db.Table(i.TableName()).
		Where("id = ?", id).
		Delete(&models.Instance{}).
		Error
	return err
}

func (i *instanceRepo) Update(db *gorm.DB, entity *models.Instance) error {
	entity.ModifyTime = time2.Now()
	err := db.Table(i.TableName()).Updates(entity).Error
	return err
}

func (i *instanceRepo) FindPageInstance(db *gorm.DB, page, limit int, processID, id, parentID []string, name, status string) ([]*models.Instance, int64) {
	db = db.Table(i.TableName()).
		Order("create_time desc")
	if len(id) > 0 {
		db = db.Where("id in (?)", id)
	}
	if len(processID) > 0 {
		db = db.Where("proc_id in (?)", processID)
	}
	if len(parentID) > 0 {
		db = db.Where("p_id in (?)", parentID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if name != "" {
		db = db.Where("name like ?", fmt.Sprintf("%s%s%s", "%", name, "%"))
	}
	res := make([]*models.Instance, 0)
	var num int64
	db.Model(&models.Instance{}).Count(&num)
	newPage := page2.NewPage(page, limit, num)
	db = db.Limit(newPage.PageSize).Offset(newPage.StartIndex)

	affected := db.Find(&res).RowsAffected
	if affected > 0 {
		return res, num
	}
	return nil, 0
}

func (i *instanceRepo) FindByUserID(db *gorm.DB, page, limit int, req *models.QueryTaskCondition) ([]*models.Instance, int64) {
	db1 := db.Table("proc_history_task").
		Select("distinct proc_instance_id").
		Where("proc_history_task.status != ?", internal.Deleted)
	if req.UserID != "" {
		db1 = db1.Where("assignee = ?", req.UserID)
	}
	if req.NodeDefKey != "" {
		db1 = db1.Where("node_def_key = ?", req.NodeDefKey)
	}
	if len(req.InstanceID) > 0 {
		db1 = db1.Where("proc_instance_id in (?)", req.InstanceID)
	}
	if len(req.ProcessID) > 0 {
		db1 = db1.Where("proc_id in (?)", req.ProcessID)
	}
	if len(req.TaskID) > 0 {
		db1 = db1.Where("id in (?)", req.TaskID)
	}
	if req.Name != "" {
		db1 = db1.Where("name like ?", "%"+req.Name+"%")
	}
	if req.Des != "" {
		db1 = db1.Where("`desc` REGEXP ?", req.Des)
	}
	orderBy := "modify_time desc"
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
	var num int64
	var ids []string
	db1.Model(&models.HistoryTask{}).Distinct("proc_instance_id").
		Joins("left join proc_instance on proc_history_task.proc_instance_id = proc_instance.id").
		Where("proc_instance.app_status != ?", internal.Suspend).
		Count(&num)
	newPage := page2.NewPage(page, limit, num)
	query := fmt.Sprintf("%s%s%s", `select * from proc_instance where app_status != ? and id in (?) order by `, orderBy, ` limit ?,?`)
	db = db.Raw(query,
		internal.Suspend,
		db1.Model(&ids),
		newPage.StartIndex,
		newPage.PageSize,
	)
	var res []*models.Instance
	db.Scan(&res)
	return res, num
}

func (i *instanceRepo) FindAgencyByUserID(db *gorm.DB, page, limit int, req *models.QueryTaskCondition) ([]*models.Instance, int64) {
	db1 := db.Table("proc_task").
		Select("distinct proc_instance_id")

	if req.UserID != "" && req.GroupID != "" {
		db1 = db1.Where("proc_task.id in (select distinct task_id from proc_task_identity where user_id = ? or group_id = ?)", req.UserID, req.GroupID)
	} else {
		db1 = db1.Where("proc_task.id in (select distinct task_id from proc_task_identity )")
	}
	if req.NodeDefKey != "" {
		db1 = db1.Where("node_def_key = ?", req.NodeDefKey)
	}
	if len(req.InstanceID) > 0 {
		db1 = db1.Where("proc_instance_id in (?)", req.InstanceID)
	}
	if len(req.ProcessID) > 0 {
		db1 = db1.Where("proc_id in (?)", req.ProcessID)
	}
	if len(req.TaskID) > 0 {
		db1 = db1.Where("proc_task.id in (?)", req.TaskID)
	}
	if req.Name != "" {
		db1 = db1.Where("name like ?", fmt.Sprintf("%s%s%s", "%", req.Name, "%"))
	}
	if req.Des != "" {
		db1 = db1.Where("`desc` REGEXP ?", req.Des)
	}
	orderBy := "modify_time desc"
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
	var num int64
	var ids []string
	db1.Model(&models.HistoryTask{}).Distinct("proc_instance_id").
		Joins("left join proc_instance on proc_task.proc_instance_id = proc_instance.id").
		Where("proc_instance.app_status != ?", internal.Suspend).
		Count(&num)
	newPage := page2.NewPage(page, limit, num)
	query := fmt.Sprintf("%s%s%s", `select * from proc_instance where app_status != ? and id in (?) order by `, orderBy, ` limit ?,?`)
	db = db.Raw(query,
		internal.Suspend,
		db1.Model(&ids),
		newPage.StartIndex,
		newPage.PageSize,
	)
	var res []*models.Instance
	db.Scan(&res)
	return res, num
}

func (i *instanceRepo) FindAllByUserID(db *gorm.DB, page, limit int, req *models.QueryTaskCondition) ([]*models.Instance, int64) {
	db1 := db.Table("proc_history_task").
		Select("distinct proc_instance_id").
		Where("proc_history_task.status != ?", internal.Deleted)
	if req.UserID != "" {
		db1 = db1.Where("assignee = ?", req.UserID)
	}
	if req.NodeDefKey != "" {
		db1 = db1.Where("node_def_key = ?", req.NodeDefKey)
	}
	if len(req.InstanceID) > 0 {
		db1 = db1.Where("proc_instance_id in (?)", req.InstanceID)
	}
	if len(req.ProcessID) > 0 {
		db1 = db1.Where("proc_id in (?)", req.ProcessID)
	}
	if len(req.TaskID) > 0 {
		db1 = db1.Where("proc_history_task.id in (?)", req.TaskID)
	}
	if req.Name != "" {
		db1 = db1.Where("name like ?", fmt.Sprintf("%s%s%s", "%", req.Name, "%"))
	}
	if req.Des != "" {
		db1 = db1.Where("`desc` REGEXP ?", req.Des)
	}

	db2 := db.Table("proc_task").
		Select("distinct proc_instance_id")

	if req.UserID != "" && req.GroupID != "" {
		db2 = db2.Where("proc_task.id in (select distinct task_id from proc_task_identity where user_id = ? or group_id = ?)", req.UserID, req.GroupID)
	} else {
		db2 = db2.Where("proc_task.id in (select distinct task_id from proc_task_identity )")
	}
	if req.NodeDefKey != "" {
		db2 = db2.Where("node_def_key = ?", req.NodeDefKey)
	}
	if len(req.InstanceID) > 0 {
		db2 = db2.Where("proc_instance_id in (?)", req.InstanceID)
	}
	if len(req.ProcessID) > 0 {
		db2 = db2.Where("proc_id in (?)", req.ProcessID)
	}
	if len(req.TaskID) > 0 {
		db2 = db2.Where("proc_task.id in (?)", req.TaskID)
	}
	if req.Name != "" {
		db2 = db2.Where("name like ?", "%"+req.Name+"%")
	}
	if req.Des != "" {
		db2 = db2.Where("`desc` REGEXP ?", req.Des)
	}
	orderBy := "modify_time desc"
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
	var ids1, ids2 []string
	db1.Model(&models.Instance{}).
		Joins("left join proc_instance on proc_history_task.proc_instance_id = proc_instance.id").
		Where("proc_instance.app_status != ?", internal.Suspend).
		Scan(&ids1)
	db2.Model(&models.Instance{}).
		Joins("left join proc_instance on proc_task.proc_instance_id = proc_instance.id").
		Where("proc_instance.app_status != ?", internal.Suspend).
		Scan(&ids2)
	ids := make([]string, 0)
	ids = append(ids, ids1...)
	ids = append(ids, ids2...)
	idsn := removeRepeatedElement(ids)
	num := int64(len(idsn))
	newPage := page2.NewPage(page, limit, num)
	query := fmt.Sprintf("%s%s%s", `select * from proc_instance where app_status != ? and id in (?) order by `, orderBy, ` limit ?,?`)
	db = db.Raw(query,
		internal.Suspend,
		idsn,
		newPage.StartIndex,
		newPage.PageSize,
	)
	var res []*models.Instance
	db.Scan(&res)
	return res, num
}

func removeRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}
