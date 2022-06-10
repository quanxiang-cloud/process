package restful

import (
	"github.com/quanxiang-cloud/process/internal/server/events"
	"github.com/quanxiang-cloud/process/internal/server/options"
	"github.com/quanxiang-cloud/process/pkg/config"
	"github.com/quanxiang-cloud/process/pkg/misc/logger"
	"github.com/quanxiang-cloud/process/pkg/misc/mysql2"

	"github.com/gin-gonic/gin"
)

const (
	// DebugMode indicates mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates mode is release.
	ReleaseMode = "release"
)

// Router 路由
type Router struct {
	c *config.Configs

	engine *gin.Engine
}

// NewRouter 开启路由
func NewRouter(c *config.Configs) (*Router, error) {
	engine, err := newRouter(c)
	if err != nil {
		return nil, err
	}
	db, err := mysql2.New(c.Mysql, logger.Logger)
	if err != nil {
		return nil, err
	}
	optDB := options.WithDB(db)
	listen, err := events.NewListener(c)
	if err != nil {
		return nil, err
	}
	optListen := options.WithListener(listen)

	process, err := NewProcess(c, optDB, optListen)
	if err != nil {
		return nil, err
	}
	v1 := engine.Group("/api/v1/process")
	{
		v1.POST("/deploy", process.Deploy)
		v1.POST("/startInstance", process.StartInstance)
		v1.POST("/initInstance", process.InitInstance)
		// 完成任务 当指定下个节点时，会按指定的节点走
		v1.POST("/completeTask", process.CompleteTask)
		v1.POST("/completeNonModelTasks", process.BatchCompleteNonModelTask)
		// 完成execution，需保证是在分支上或会签节点任务上调用
		v1.POST("/completeExecution", process.CompleteExecution)
		v1.POST("/agencyTask", process.AgencyTask)
		v1.POST("/agencyTaskTotal", process.AgencyTaskTotal)
		// 查询指定人参与的task的流程实例
		v1.POST("/doneInstance", process.DoneInstance)
		v1.POST("/agencyInstance", process.AgencyInstance)
		v1.POST("/wholeInstance", process.WholeInstance)
		// 查询人参与的流程实例的已完成的task任务
		v1.POST("/doneTask", process.InstanceDoneTask)
		v1.POST("/wholeTask", process.WholeTask)
		v1.POST("/deleteTask", process.DeleteTask)
		v1.POST("/deleteInstance", process.DeleteInstance)
		v1.POST("/terminatedInstance", process.TerminatedInstance)
		// 根据instanceID查询instance
		v1.POST("/listInstance", process.InstanceList)
		// copy for, read etc
		v1.POST("/addTask", process.AddTask)
		// 前加签
		v1.POST("/addFrondTask", process.AddFrondModelTask)
		// 后加签
		v1.POST("/addBackTask", process.AddBackModelTask)
		// 回退
		v1.POST("/fallbackTask", process.FallbackTask)
		v1.POST("/refillTask", process.BackReFillTask)
		// 指定处理人(转交)
		v1.POST("/assigneeTask", process.TransferTask)
		v1.POST("/dueTimeTask", process.UpdateTask)
		v1.POST("/ListProcessNode", process.ListProcessNode)
		// 增加一个已办任务
		v1.POST("/addDoneTask", process.AddHistoryTask)
		// 保存变量
		v1.POST("/saveVariables", process.AddVariables)
		v1.POST("/getVariables", process.GetVariables)
		// 查询分支的其他execution
		v1.POST("/gatewayExecution", process.GetGateWayExecution)
		// 查询指定合流网关的下级网关的executionID
		v1.POST("/inclusiveExecution", process.InclusiveGateWayExecution)
		// 查询任务节点的前节点
		v1.POST("/preNode", process.TaskPreNode)
		// 应用删除时，对应的流程实例挂起 或恢复
		v1.POST("/appOperation", process.AppDelete)
	}

	return &Router{
		c:      c,
		engine: engine,
	}, nil
}

func newRouter(c *config.Configs) (*gin.Engine, error) {
	if c.Model == "" || (c.Model != ReleaseMode && c.Model != DebugMode) {
		c.Model = ReleaseMode
	}
	gin.SetMode(c.Model)
	engine := gin.New()
	engine.Use(logger.GinLogger(), logger.GinRecovery())
	return engine, nil
}

// Run 启动服务
func (r *Router) Run() {
	r.engine.Run(r.c.Port)
}

// Close 关闭服务
func (r *Router) Close() {
}
