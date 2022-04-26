package spool

import (
	"context"
	"fmt"
)

type workParamFunc func(interface{})
type workParam interface{}

type PoolWithFunc struct {
	cancel   context.CancelFunc
	workPool chan workParam
}

// NewPoolWithFunc 反馈一个固定方法的池
func NewPoolWithFunc(size int, fn workParamFunc) *PoolWithFunc {
	ctx, cancel := context.WithCancel(context.Background())

	workerLine := make(chan *workerWithFunc, size)
	workFuncLine := make(chan workParam, 1)
	dis := newDispatchWithFunc(ctx, workerLine, workFuncLine)
	go dis.run()
	for i := 0; i < size; i++ {
		w := newWorkerWithFunc(i, ctx, workerLine, fn)
		go w.run()
	}
	return &PoolWithFunc{
		workPool: workFuncLine,
		cancel:   cancel,
	}
}

func (p *PoolWithFunc) Submit(work workParam) {
	p.workPool <- work
}

func (p *PoolWithFunc) Release() {
	p.cancel()
}

type dispatchWithFunc struct {
	ctx        context.Context
	workerPool chan *workerWithFunc
	workPool   chan workParam
}

func newDispatchWithFunc(ctx context.Context, w chan *workerWithFunc, wf chan workParam) *dispatchWithFunc {
	return &dispatchWithFunc{
		ctx:        ctx,
		workerPool: w,
		workPool:   wf,
	}
}

func (d *dispatchWithFunc) run() {
	for {
		select {
		case work := <-d.workPool:
			worker, ok := <-d.workerPool
			if !ok {
				return
			}
			worker.submit(work)
		case <-d.ctx.Done():
			return
		}
	}
}

type workerWithFunc struct {
	workId     int
	ctx        context.Context
	workerPool chan *workerWithFunc
	work       chan workParam
	workFunc   workParamFunc
}

func newWorkerWithFunc(workId int, ctx context.Context, pool chan *workerWithFunc, fn workParamFunc) *workerWithFunc {
	return &workerWithFunc{
		workId:     workId,
		ctx:        ctx,
		workerPool: pool,
		work:       make(chan workParam),
		workFunc:   fn,
	}
}

func (w *workerWithFunc) submit(f interface{}) {
	w.work <- f
}

func (w *workerWithFunc) run() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("worker run:", err)
		}
	}()
	for {
		w.workerPool <- w
		select {
		case param := <-w.work:
			w.workFunc(param)
		case <-w.ctx.Done():
			return
		}
	}
}
