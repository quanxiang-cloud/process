package mysql

import (
	"git.internal.yunify.com/qxp/misc/id2"
	"git.internal.yunify.com/qxp/misc/time2"
	"git.internal.yunify.com/qxp/process/internal/models"
	"gorm.io/gorm"
)

type executionRepo struct{}

func (e *executionRepo) TableName() string {
	return "proc_execution"
}

// NewExecutionRepo new
func NewExecutionRepo() models.ExecutionRepo {
	return &executionRepo{}
}

// Create create execution
func (e *executionRepo) Create(db *gorm.DB, entity *models.Execution) error {
	entity.ID = id2.GenID()
	entity.CreateTime = time2.Now()
	err := db.Table(e.TableName()).
		Create(entity).
		Error
	return err
}

// Update update execution
func (e *executionRepo) Update(db *gorm.DB, ID string, updateMap map[string]interface{}) (*models.Execution, error) {
	updateMap["modify_time"] = time2.NowUnix()
	err := db.Table(e.TableName()).Where("id=?", ID).Updates(updateMap).Error

	if err != nil {
		return nil, err
	}
	return e.FindByID(db, ID)
}

// SetActive set active status
func (e *executionRepo) SetActive(db *gorm.DB, ID string, isActive int8) error {
	updateEntity := map[string]interface{}{
		"modify_time": time2.NowUnix(),
		"is_active":   isActive,
	}
	err := db.Table(e.TableName()).Where("id=?", ID).Updates(updateEntity).Error
	return err
}

// DeleteByID delete execution by ID
func (e *executionRepo) DeleteByID(db *gorm.DB, ID string) error {
	entity := &models.Execution{ID: ID}
	err := db.Table(e.TableName()).Delete(entity).Error
	return err
}

func (e *executionRepo) DeleteByInstanceID(db *gorm.DB, instanceID string) error {
	entity := &models.Execution{}
	err := db.Table(e.TableName()).
		Where("proc_instance_id=?", instanceID).
		Delete(entity).
		Error
	return err
}

func (e *executionRepo) FindByInstanceID(db *gorm.DB, instanceID string) ([]*models.Execution, error) {
	es := make([]*models.Execution, 0)
	err := db.Table(e.TableName()).
		Where("proc_instance_id=?", instanceID).
		Find(&es).
		Error
	return es, err
}

// FindByID find by ID
func (e *executionRepo) FindByID(db *gorm.DB, ID string) (*models.Execution, error) {
	entity := new(models.Execution)
	err := db.Table(e.TableName()).
		Where("id = ?", ID).
		Find(entity).
		Error
	if err != nil {
		return nil, err
	}
	if entity.ID == "" {
		return nil, nil
	}
	return entity, nil
}

// FindByPID find by PID
func (e *executionRepo) FindByPID(db *gorm.DB, instanceID string, PID string, active int8) ([]*models.Execution, error) {
	es := make([]*models.Execution, 0)
	err := db.Table(e.TableName()).
		Where(map[string]interface{}{
			"proc_instance_id": instanceID,
			"p_id":             PID,
			"is_active":        active,
		}).
		Find(&es).
		Error
	if err != nil {
		return nil, err
	}
	return es, nil
}

// FindByPID find by PID
func (e *executionRepo) FindAllByPID(db *gorm.DB, instanceID string, PID string) ([]*models.Execution, error) {
	es := make([]*models.Execution, 0)
	err := db.Table(e.TableName()).
		Where(map[string]interface{}{
			"proc_instance_id": instanceID,
			"p_id":             PID,
		}).
		Find(&es).
		Error
	if err != nil {
		return nil, err
	}
	return es, nil
}

func (e *executionRepo) FindInstanceParentExecution(db *gorm.DB, instanceID string) (*models.Execution, error) {
	entity := new(models.Execution)
	err := db.Table(e.TableName()).
		Where("proc_instance_id = ? and p_id = ''", instanceID).
		Find(entity).
		Error
	if err != nil {
		return nil, err
	}
	if entity.ID == "" {
		return nil, nil
	}
	return entity, nil
}
