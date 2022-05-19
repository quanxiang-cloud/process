package mysql

import (
	"git.internal.yunify.com/qxp/process/internal/models"
	page2 "git.internal.yunify.com/qxp/process/pkg/page"
	"gorm.io/gorm"
)

type taskIdentityRepo struct{}

func (ti *taskIdentityRepo) TableName() string {
	return "proc_task_identity"
}

// NewTaskIdentityRepo new
func NewTaskIdentityRepo() models.TaskIdentityRepo {
	return &taskIdentityRepo{}
}

func (ti *taskIdentityRepo) Create(db *gorm.DB, entity *models.TaskIdentity) error {
	err := db.Table(ti.TableName()).
		Create(entity).
		Error
	return err
}

func (ti *taskIdentityRepo) Update(db *gorm.DB, entity *models.TaskIdentity) error {
	err := db.Table(ti.TableName()).
		// Where("id = ?", entity.ID).
		Updates(entity).
		Error
	return err
}

func (ti *taskIdentityRepo) DeleteByInstanceID(db *gorm.DB, instanceID string) error {
	err := db.Table(ti.TableName()).
		Where("instance_id = ?", instanceID).
		Delete(&models.Task{}).
		Error
	return err
}

func (ti *taskIdentityRepo) FindUserInstanceTask(db *gorm.DB, instanceID, taskID, userID, groupID string) (*models.TaskIdentity, error) {
	modal := new(models.TaskIdentity)
	err := db.Table(ti.TableName()).
		Where("instance_id = ? and task_id = ? and (user_id = ? or group_id = ?)", instanceID, taskID, userID, groupID).
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

func (ti *taskIdentityRepo) FindByUserID(db *gorm.DB, userID, groupID string) ([]string, error) {
	res := make([]string, 0)
	err := db.Table(ti.TableName()).
		Select("distinct task_id").
		Where("user_id = ?", userID).
		Or("group_id = ?", groupID).
		Find(&res).
		Error
	return res, err
}

func (ti *taskIdentityRepo) FindPageByUserID(db *gorm.DB, page, limit int, instanceID []string, userID, groupID string) ([]*models.TaskIdentity, int64) {
	db = db.Table(ti.TableName()).
		Select("distinct task_id,create_time").
		Where("user_id = ?", userID).
		Or("group_id = ?", groupID).
		Order("create_time desc")
	if len(instanceID) > 0 {
		db = db.Where("instance_id in (?)", instanceID)
	}
	res := make([]*models.TaskIdentity, 0)
	var num int64
	db.Model(&models.TaskIdentity{}).Count(&num)
	newPage := page2.NewPage(page, limit, num)
	db = db.Limit(newPage.PageSize).Offset(newPage.StartIndex)

	affected := db.Find(&res).RowsAffected
	if affected > 0 {
		return res, num
	}
	return nil, 0
}

func (ti *taskIdentityRepo) DeleteByTaskID(db *gorm.DB, taskID string) error {
	err := db.Table(ti.TableName()).
		Where("task_id = ?", taskID).
		Delete(models.TaskIdentity{}).
		Error
	return err
}

func (ti *taskIdentityRepo) DeleteByID(db *gorm.DB, id string) error {
	err := db.Table(ti.TableName()).
		Where("id = ?", id).
		Delete(models.TaskIdentity{}).
		Error
	return err
}
