package code

import "github.com/quanxiang-cloud/process/pkg/misc/error2"

func init() {
	error2.CodeTable = CodeTable
}

const (
	// InvalidURI 无效的URI
	InvalidURI = 90014000000
	// InvalidParams 无效的参数
	InvalidParams = 90014000001
	// InvalidTimestamp 无效的时间格式
	InvalidTimestamp = 90014000002
	// NameExist 名字已经存在
	NameExist = 90014000003
	// InvalidDel 无效的删除
	InvalidDel = 90014000004
	// InvalidProcessID process id no exist
	InvalidProcessID = 90014000005
	// InvalidTaskID task id no exist
	InvalidTaskID = 90014000006
	// NoResult NoResult
	NoResult = 90014000007
	// AllConditionMissMatch no condition match
	AllConditionMissMatch = 90014000008
	// InvalidExecutionID execution id no exist
	InvalidExecutionID = 90014000009
	// NoBranchExecution no branch execution
	NoBranchExecution = 90014000010
	// InvalidNodeDefKey node defKey no exist
	InvalidNodeDefKey = 90014000011
	// ConditionParamError condition judge error
	ConditionParamError = 90014000012
	// InstanceInitError instance already init
	InstanceInitError = 90014000013
)

// CodeTable 码表
var CodeTable = map[int64]string{
	InvalidURI:            "无效的URI.",
	InvalidParams:         "无效的参数.",
	InvalidTimestamp:      "无效的时间格式.",
	NameExist:             "名称已被使用！请检查后重试！",
	InvalidDel:            "删除无效！对象不存在或请检查参数！",
	InvalidProcessID:      "流程id不存在",
	InvalidTaskID:         "任务id不存在",
	InvalidExecutionID:    "执行id不存在",
	NoResult:              "查询不到结果",
	AllConditionMissMatch: "所有分支的条件都不满足",
	NoBranchExecution:     "非分支执行节点",
	InvalidNodeDefKey:     "节点defKey不存在",
	ConditionParamError:   "分支条件判断错误",
	InstanceInitError:     "实例已经初始化",
}
