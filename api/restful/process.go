package restful

import (
	"git.internal.yunify.com/qxp/misc/logger"
	"git.internal.yunify.com/qxp/misc/resp"
	"git.internal.yunify.com/qxp/process/internal/dispatcher"
	"git.internal.yunify.com/qxp/process/internal/process"
	listener "git.internal.yunify.com/qxp/process/internal/server/events"
	"git.internal.yunify.com/qxp/process/internal/server/options"
	"git.internal.yunify.com/qxp/process/pkg/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

// Process process
type Process struct {
	process  process.Process
	workFlow dispatcher.Flow
	instance process.Instance
	task     process.Task
	db       *gorm.DB
	l        *listener.Listener
}

// NewProcess new process
func NewProcess(c *config.Configs, opts ...options.Options) (*Process, error) {
	p, err := process.NewProcess(c, opts...)
	if err != nil {
		return nil, err
	}
	i, err := process.NewInstance(c, opts...)
	if err != nil {
		return nil, err
	}
	t, err := process.NewTask(c, opts...)
	if err != nil {
		return nil, err
	}
	f, err := dispatcher.NewWorkFlow(c, opts...)
	if err != nil {
		return nil, err
	}

	return &Process{
		process:  p,
		instance: i,
		workFlow: f,
		task:     t,
	}, nil
}

// SetDB set db
func (p *Process) SetDB(db *gorm.DB) {
	p.db = db
}

// Deploy process model
func (p *Process) Deploy(c *gin.Context) {
	// profile := header2.GetProfile(c)
	rq := &process.AddModelReq{}
	err := c.ShouldBind(rq)
	if err != nil {
		logger.Logger.Error(err)
		resp.Format(nil, err).Context(c)
		return
	}
	// rq.CreatorID = profile.UserID
	r, err := p.process.AddModel(logger.CTXTransfer(c), rq)
	if err != nil {
		logger.Logger.Error(err)
		resp.Format(nil, err).Context(c)
		return
	}
	resp.Format(r, nil).Context(c)
	return
}

// StartInstance start a instance
func (p *Process) StartInstance(c *gin.Context) {
	req := &process.StartProcessReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.Start(logger.CTXTransfer(c), req)).Context(c)
}

// InitInstance init a instance
func (p *Process) InitInstance(c *gin.Context) {
	req := &process.InitInstanceReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.InitInstance(logger.CTXTransfer(c), req)).Context(c)
}

// CompleteTask complete task
func (p *Process) CompleteTask(c *gin.Context) {
	req := &process.CompleteTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.CompleteTask(logger.CTXTransfer(c), req)).Context(c)
}

// BatchCompleteNonModelTask complete non model task
func (p *Process) BatchCompleteNonModelTask(c *gin.Context) {
	req := &process.CompleteNonModelTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.BatchCompleteNonModelTask(logger.CTXTransfer(c), req)).Context(c)
}

// CompleteExecution complete execution
func (p *Process) CompleteExecution(c *gin.Context) {
	req := &process.CompleteExecutionReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.CompleteExecution(logger.CTXTransfer(c), req)).Context(c)
}

// AgencyTask agency task
func (p *Process) AgencyTask(c *gin.Context) {
	req := &process.QueryTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.AgencyTask(logger.CTXTransfer(c), req)).Context(c)
}

// AgencyTaskTotal agency task count
func (p *Process) AgencyTaskTotal(c *gin.Context) {
	req := &process.QueryTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.AgencyTaskTotal(logger.CTXTransfer(c), req)).Context(c)
}

// DoneInstance done instance list
func (p *Process) DoneInstance(c *gin.Context) {
	req := &process.QueryTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.DoneInstance(logger.CTXTransfer(c), req)).Context(c)
}

// AgencyInstance agency instance list
func (p *Process) AgencyInstance(c *gin.Context) {
	req := &process.QueryTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.AgencyInstance(logger.CTXTransfer(c), req)).Context(c)
}

// WholeInstance whole instance list
func (p *Process) WholeInstance(c *gin.Context) {
	req := &process.QueryTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.WholeInstance(logger.CTXTransfer(c), req)).Context(c)
}

// WholeTask all task
func (p *Process) WholeTask(c *gin.Context) {
	req := &process.QueryTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.WholeTask(logger.CTXTransfer(c), req)).Context(c)
}

// InstanceDoneTask instance done task detail
func (p *Process) InstanceDoneTask(c *gin.Context) {
	req := &process.QueryTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.InstanceDoneTasks(logger.CTXTransfer(c), req)).Context(c)
}

// DeleteTask delete task
func (p *Process) DeleteTask(c *gin.Context) {
	req := &process.DeleteTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.DeleteTask(logger.CTXTransfer(c), req)).Context(c)
}

// DeleteInstance delete instance
func (p *Process) DeleteInstance(c *gin.Context) {
	req := &process.DeleteProcessReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.DeleteInstance(logger.CTXTransfer(c), req)).Context(c)
}

// TerminatedInstance delete instance
func (p *Process) TerminatedInstance(c *gin.Context) {
	req := &process.DeleteProcessReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.TerminatedInstance(logger.CTXTransfer(c), req)).Context(c)
}

// InstanceList list instance
func (p *Process) InstanceList(c *gin.Context) {
	req := &process.ListProcessReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.InstanceList(logger.CTXTransfer(c), req)).Context(c)
}

// AddTask add task
func (p *Process) AddTask(c *gin.Context) {
	req := &process.AddTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.AddNonNodeTask(logger.CTXTransfer(c), req)).Context(c)
}

// AddFrondModelTask add task
func (p *Process) AddFrondModelTask(c *gin.Context) {
	req := &process.AddTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.AddNodeTask(logger.CTXTransfer(c), req)).Context(c)
}

// AddBackModelTask add task back
func (p *Process) AddBackModelTask(c *gin.Context) {
	req := &process.AddTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.AddBackNodeTask(logger.CTXTransfer(c), req)).Context(c)
}

// BackReFillTask refill task
func (p *Process) BackReFillTask(c *gin.Context) {
	req := &process.AddTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.BackReFill(logger.CTXTransfer(c), req)).Context(c)
}

// FallbackTask fallback task
func (p *Process) FallbackTask(c *gin.Context) {
	req := &process.AddTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.FallbackTask(logger.CTXTransfer(c), req)).Context(c)
}

// UpdateTask update task
func (p *Process) UpdateTask(c *gin.Context) {
	req := &process.AddTaskConditionReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// profile := header2.GetProfile(c)
	// req.UserID = profile.UserID
	resp.Format(p.task.AddTaskCondition(logger.CTXTransfer(c), req)).Context(c)
}

// TransferTask update task
func (p *Process) TransferTask(c *gin.Context) {
	req := &process.AddTaskConditionReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// profile := header2.GetProfile(c)
	// req.UserID = profile.UserID
	resp.Format(p.task.TransferTask(logger.CTXTransfer(c), req)).Context(c)
}

// ListProcessNode list node
func (p *Process) ListProcessNode(c *gin.Context) {
	req := &process.QueryNodeReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.process.FindProcessNode(logger.CTXTransfer(c), req)).Context(c)
}

// AddHistoryTask AddHistoryTask
func (p *Process) AddHistoryTask(c *gin.Context) {
	req := &process.AddHistoryTaskReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.AddHistoryTask(logger.CTXTransfer(c), req)).Context(c)
}

// AddVariables AddVariables
func (p *Process) AddVariables(c *gin.Context) {
	req := &process.SaveVariablesReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.SaveVariables(logger.CTXTransfer(c), req)).Context(c)
}

// GetVariables GetVariables
func (p *Process) GetVariables(c *gin.Context) {
	req := &process.GetVariablesReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.FindVariables(logger.CTXTransfer(c), req)).Context(c)
}

// GetGateWayExecution GetGateWayExecution
func (p *Process) GetGateWayExecution(c *gin.Context) {
	req := &process.GateWayExecutionReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.FindGateWayExecutions(logger.CTXTransfer(c), req)).Context(c)
}

// InclusiveGateWayExecution InclusiveGateWayExecution
func (p *Process) InclusiveGateWayExecution(c *gin.Context) {
	req := &process.ParentExecutionReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.FindParentExecutions(logger.CTXTransfer(c), req)).Context(c)
}

// TaskPreNode TaskPreNode
func (p *Process) TaskPreNode(c *gin.Context) {
	req := &process.TaskPreNodeReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.task.FindTaskPreNode(logger.CTXTransfer(c), req)).Context(c)
}

// AppDelete application delete
func (p *Process) AppDelete(c *gin.Context) {
	req := &process.AppDelReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(p.instance.AppDeleteHandler(logger.CTXTransfer(c), req)).Context(c)
}

// CompleteNode init next node
func (p *Process) CompleteNode(c *gin.Context) {
	req := &process.InitNextNodeReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// profile := header2.GetProfile(c)
	// req.UserID = profile.UserID

	resp.Format(nil, p.instance.CompleteNode(logger.CTXTransfer(c), req)).Context(c)
}

// NodeInstanceList get node instance list
func (p *Process) NodeInstanceList(c *gin.Context) {
	req := &process.NodeInstanceListReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	resp.Format(p.instance.NodeInstanceList(logger.CTXTransfer(c), req)).Context(c)
}
