package spool

import (
	"context"
	"encoding/json"
)

type workParamFunc func(interface{}) (WorkResult, error)

type workParam interface{}

type WorkResult interface{}

type PoolWithFunc struct {
	ctx      context.Context
	cancel   context.CancelFunc
	workPool chan workParam
}

// NewPoolWithFunc 反馈一个固定方法的池
func NewPoolWithFunc(size int, fn workParamFunc, opt ...Option) *PoolWithFunc {
	op := &option{
		HandlePoolMessage: &defaultPoolMessage{},
	}

	for _, o := range opt {
		o(op)
	}

	ctx, cancel := op.getCollectiveContext()

	workerLine := make(chan *workerWithFunc, size)
	workFuncLine := make(chan workParam, 1)
	mesLine := make(chan Message, 1) // 消息传输通道
	d := newDispatchWithFunc(ctx, op.HandlePoolMessage, workerLine, workFuncLine, mesLine)
	go d.run()
	go d.mesRun()
	for i := 0; i < size; i++ {
		w := newWorkerWithFunc(i, ctx, mesLine, workerLine, fn)
		go w.run()
	}
	return &PoolWithFunc{
		workPool: workFuncLine,
		cancel:   cancel,
		ctx:      ctx,
	}
}

// 提交一个任意类型的参数
func (p *PoolWithFunc) Submit(work workParam) {
	select {
	case <-p.ctx.Done():
		return
	default:
		p.workPool <- work
	}
}

// Release 关闭整个池
func (p *PoolWithFunc) Release() {
	p.cancel()
}

type dispatchWithFunc struct {
	ctx           context.Context
	workerPool    chan *workerWithFunc
	workPool      chan workParam
	handleMessage chan Message // 工作逻辑反馈通道
	poolMessage   PoolMessage
}

func newDispatchWithFunc(ctx context.Context, pm PoolMessage, w chan *workerWithFunc, wf chan workParam, m chan Message) *dispatchWithFunc {
	return &dispatchWithFunc{
		ctx:           ctx,
		workerPool:    w,
		workPool:      wf,
		poolMessage:   pm,
		handleMessage: m,
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

// 收集反馈的信息
func (d *dispatchWithFunc) mesRun() {
	for {
		select {
		case mes := <-d.handleMessage: // 获取执行完毕的反馈结果
			d.poolMessage.HandleMessage(mes)
		case <-d.ctx.Done():
			return
		}
	}
}

type workerWithFunc struct {
	workId        int
	ctx           context.Context
	workerPool    chan *workerWithFunc
	work          chan workParam
	workFunc      workParamFunc
	handleMessage chan Message
}

func newWorkerWithFunc(workId int, ctx context.Context, hMessage chan Message, pool chan *workerWithFunc, fn workParamFunc) *workerWithFunc {
	return &workerWithFunc{
		workId:        workId,
		ctx:           ctx,
		workerPool:    pool,
		work:          make(chan workParam),
		workFunc:      fn,
		handleMessage: hMessage,
	}
}

func (w *workerWithFunc) submit(f interface{}) {
	select {
	case <-w.ctx.Done():
		return
	default:
		w.work <- f
	}
}

func (w *workerWithFunc) run() {
	for {
		w.workerPool <- w
		select {
		case param := <-w.work:
			body := []byte{}
			b, err := w.workFunc(param)
			if err == nil {
				body, err = json.Marshal(b)
			}
			w.handleMessage <- Message{body, err}
		case <-w.ctx.Done():
			return
		}
	}
}
