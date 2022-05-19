package dispatcher

import (
	"context"
	"git.internal.yunify.com/qxp/process/internal/process"
	"git.internal.yunify.com/qxp/process/internal/server/options"
	"git.internal.yunify.com/qxp/process/pkg/config"
)

// Flow flow
type Flow interface {
	StartFlowInstance(ctx context.Context, req *process.StartProcessReq) (*process.StartProcessResp, error)
}

type workFlow struct {
	task     process.Task
	instance process.Instance
}

// NewWorkFlow NewWorkFlow
func NewWorkFlow(conf *config.Configs, opts ...options.Options) (Flow, error) {
	t, err := process.NewTask(conf, opts...)
	if err != nil {
		return nil, err
	}
	inst, err := process.NewInstance(conf, opts...)
	if err != nil {
		return nil, err
	}
	return &workFlow{
		task:     t,
		instance: inst,
	}, nil
}

func (w *workFlow) StartFlowInstance(ctx context.Context, req *process.StartProcessReq) (*process.StartProcessResp, error) {

	// resp,err := w.instance.Start(ctx,req)
	// if err != nil {
	// 	return nil,err
	// }
	// taskReq := w.task.PackInitTaskReq(ctx,resp.InstanceID)
	// err = w.task.InitTask(ctx,taskReq)
	// if err != nil {
	// 	delReq := &process.DeleteProcessReq{
	// 		InstanceID: resp.InstanceID,
	// 	}
	// 	// 手动回滚
	// 	_ = w.instance.DeleteInstance(ctx,delReq)
	// 	return nil,err
	// }
	// return resp,nil

	return nil, nil
}
