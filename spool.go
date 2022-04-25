package spool

import (
	"context"
)

type workFunc func()

type Pool struct {
	cancel   context.CancelFunc
	workPool chan workFunc
}

func NewPool(size int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	workerLine := make(chan *worker, size) // 这里必须使用 size 个缓冲，对应 @workRun 标示处的解释
	workFuncLine := make(chan workFunc, 1)
	dis := newDispatch(ctx, workerLine, workFuncLine)
	go dis.run()
	for i := 0; i < size; i++ {
		w := newWorker(i, ctx, workerLine)
		go w.run()
	}
	return &Pool{
		workPool: workFuncLine,
		cancel:   cancel,
	}
}

func (p *Pool) Submit(work workFunc) {
	p.workPool <- work
}

func (p *Pool) Release() {
	p.cancel()
}

type dispatch struct {
	ctx        context.Context
	workerPool chan *worker
	workPool   chan workFunc
}

func newDispatch(ctx context.Context, w chan *worker, wf chan workFunc) *dispatch {
	return &dispatch{
		ctx:        ctx,
		workerPool: w,
		workPool:   wf,
	}
}

func (d *dispatch) run() {
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

type worker struct {
	workId     int
	ctx        context.Context
	workerPool chan *worker
	work       chan workFunc
}

func newWorker(workId int, ctx context.Context, pool chan *worker) *worker {
	return &worker{
		workId:     workId,
		ctx:        ctx,
		workerPool: pool,
		work:       make(chan workFunc),
	}
}

func (w *worker) submit(f workFunc) {
	w.work <- f
}

// 将自身放入工人池中，获取一份工作池中的工作执行
func (w *worker) run() {
	for {
		w.workerPool <- w // @workRun ： 有缓冲这步非阻塞，不会影响下面的退出
		select {
		case wf := <-w.work:
			wf()
			//这里处理一下过程的记录
		case <-w.ctx.Done():
			return
		}
	}
}
