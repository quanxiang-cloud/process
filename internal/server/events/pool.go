package events

import (
	"context"
	"sync"
	"sync/atomic"
)

// TaskFunc TaskFunc
type TaskFunc func(ctx context.Context, param Param) error

// Task Task
type Task struct {
	f   TaskFunc
	ctx context.Context
	p   Param
}

// WorkPool WorkPool
type WorkPool struct {
	pool           chan *Task
	workerCount    int
	started        int32
	wg             sync.WaitGroup
	stopCtx        context.Context
	stopCancelFunc context.CancelFunc
}

// Execute Execute
func (t *Task) Execute() {
	t.f(t.ctx, t.p)
}

// NewPool New
func NewPool(workerCount, poolLen int, started int32) *WorkPool {
	return &WorkPool{
		pool:        make(chan *Task, poolLen),
		workerCount: workerCount,
		started:     started,
	}
}

// PushTask PushTask
func (w *WorkPool) PushTask(t *Task) {
	w.pool <- t
}

// PushTaskFunc PushTask
func (w *WorkPool) PushTaskFunc(ctx context.Context, param Param, f TaskFunc) {
	w.pool <- &Task{
		f:   f,
		ctx: ctx,
		p:   param,
	}
}

func (w *WorkPool) work(ctx context.Context) {
	for {
		select {
		case <-w.stopCtx.Done():
			w.wg.Done()
			return
		case t := <-w.pool:
			t.Execute()
		case <-ctx.Done():
			return
		}
	}
}

// Start Start
func (w *WorkPool) Start(ctx context.Context) *WorkPool {
	sd := atomic.CompareAndSwapInt32(&w.started, 0, 1)
	if sd {
		w.wg.Add(w.workerCount)
		w.stopCtx, w.stopCancelFunc = context.WithCancel(context.Background())
		for i := 0; i < w.workerCount; i++ {
			go w.work(ctx)
		}
	}
	return w
}

// Stop Stop
func (w *WorkPool) Stop() {
	w.stopCancelFunc()
	w.wg.Wait()
}
