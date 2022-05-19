package mysql

import (
	"encoding/json"
	"git.internal.yunify.com/qxp/misc/time2"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/pkg"
	"gorm.io/gorm"
)

type variablesRepo struct{}

func (v *variablesRepo) TableName() string {
	return "proc_variables"
}

// NewVariablesRepo new
func NewVariablesRepo() models.VariablesRepo {
	return &variablesRepo{}
}

// Create create variable
func (v *variablesRepo) Create(db *gorm.DB, entity *models.Variables) error {
	err := db.Table(v.TableName()).
		Create(entity).
		Error
	return err
}

func (v *variablesRepo) Update(db *gorm.DB, entity *models.Variables) error {
	entity.ModifyTime = time2.Now()
	err := db.Table(v.TableName()).
		Where("id = ?", entity.ID).
		Updates(entity).
		Error
	return err
}

// GetStringValue get value is string type
func (v *variablesRepo) GetStringValue(db *gorm.DB, instanceID string, name string) (string, error) {
	variables := new(models.Variables)
	err := db.Table(v.TableName()).
		Where(map[string]interface{}{
			"proc_instance_id": instanceID,
			"name":             name,
			"var_type":         "string",
		}).
		Find(variables).
		Error
	if err != nil {
		return "", err
	}
	if variables != nil {
		return variables.Value, nil
	}

	return "", nil
}

// GetStringArrayValue get value is []string type
func (v *variablesRepo) GetStringArrayValue(db *gorm.DB, instanceID, nodeID, name string) ([]string, error) {
	variables := make([]*models.Variables, 0)
	err := db.Table(v.TableName()).
		Where("proc_instance_id = ? and name = ? and node_id = ?", instanceID, name, nodeID).
		Find(&variables).
		Error
	if err != nil {
		return nil, err
	}
	if len(variables) > 0 {
		var vv []string
		err := json.Unmarshal([]byte(variables[0].ComplexValue.String()), &vv)
		if err != nil {
			return nil, err
		}
		return vv, nil
	}
	return nil, nil
}

func (v *variablesRepo) GetInstanceValue(db *gorm.DB, instanceID string) (map[string]interface{}, error) {
	variables := make([]*models.Variables, 0)
	res := make(map[string]interface{}, 0)
	err := db.Table(v.TableName()).
		Where(map[string]interface{}{
			"proc_instance_id": instanceID,
		}).
		Find(&variables).
		Error
	if err != nil {
		return nil, err
	}
	if len(variables) > 0 {
		for _, v := range variables {
			switch v.VarType {
			case "string":
				res[v.Name] = v.Value
			case "[]string":
				res[v.Name] = v.ComplexValue
			}
		}
	}
	return res, nil
}

func (v *variablesRepo) GetInstanceValueAndParams(db *gorm.DB, instanceID string, params map[string]interface{}) (map[string]interface{}, error) {
	variables, err := v.GetInstanceValue(db, instanceID)
	if err != nil {
		return nil, err
	}

	return pkg.MergeMap(variables, params), nil
}

func (v *variablesRepo) GetInstanceValueByName(db *gorm.DB, instanceID string, names []string) (map[string]interface{}, error) {
	variables := make([]*models.Variables, 0)
	res := make(map[string]interface{}, 0)
	err := db.Table(v.TableName()).
		Where("proc_instance_id = ? and name in (?)", instanceID, names).
		Find(&variables).
		Error
	if err != nil {
		return nil, err
	}
	if len(variables) > 0 {
		for _, v := range variables {
			switch v.VarType {
			case "string", "int", "float":
				res[v.Name] = v.Value
			case "[]string":
				res[v.Name] = v.ComplexValue
			}
		}
	}
	return res, nil
}

func (v *variablesRepo) FindVariablesByName(db *gorm.DB, instanceID, nodeID, name string) (*models.Variables, error) {
	variables := new(models.Variables)
	err := db.Table(v.TableName()).
		Where("proc_instance_id = ? and node_id = ? and name = ?", instanceID, nodeID, name).
		Find(variables).
		Error
	if err != nil {
		return nil, err
	}
	if variables.ID == "" {
		return nil, nil
	}
	return variables, nil
}
