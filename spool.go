package spool

import (
	"context"
)

type workFunc func() ([]byte, error)

type Pool struct {
	cancel   context.CancelFunc // context 的根节点取消方法，用于杀死所有工作逻辑
	workPool chan workFunc      // 工作池，需要执行的逻辑通道
}

// 反馈一个普通池
// 开启调度器
// 开启对应数量的协程
func NewPool(size int, opt ...Option) *Pool {
	op := &option{
		HandlePoolMessage: &defaultPoolMessage{},
	}

	for _, o := range opt {
		o(op)
	}

	ctx, cancel := context.WithCancel(context.Background())

	workerLine := make(chan *worker, size) // 这里必须使用 size 个缓冲，对应 @workRun 标示处的解释
	workFuncLine := make(chan workFunc, 1)
	mesLine := make(chan Message, 1) // 消息传输通道
	d := newDispatch(ctx, op.HandlePoolMessage, workerLine, workFuncLine, mesLine)
	go d.run()
	go d.mesRun()
	for i := 0; i < size; i++ {
		w := newWorker(i, ctx, workerLine, mesLine)
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

// 调度器
// 用于调度待执行的任务和可执行的协程之间的关系
// 并将执行结果反馈给 HandleMessage
type dispatch struct {
	ctx           context.Context
	workerPool    chan *worker // 工作者池，空闲的工作者通道
	workPool      chan workFunc
	handleMessage chan Message // 工作逻辑反馈通道
	poolMessage   PoolMessage
}

func newDispatch(ctx context.Context, pm PoolMessage, w chan *worker, wf chan workFunc, m chan Message) *dispatch {
	return &dispatch{
		ctx:           ctx,
		workerPool:    w,
		workPool:      wf,
		poolMessage:   pm,
		handleMessage: m,
	}
}

func (d *dispatch) run() {
	for {
		select {
		case work := <-d.workPool: // 获取一个需要执行的工作，并获取一个空闲的工作者
			worker, ok := <-d.workerPool
			if !ok {
				return
			}
			worker.submit(work) // 将获取的工作放入对应工作者的工作池进行处理
		case <-d.ctx.Done():
			return
		}
	}
}

// 收集反馈的信息
func (d *dispatch) mesRun() {
	for {
		select {
		case mes := <-d.handleMessage: // 获取执行完毕的反馈结果
			d.poolMessage.HandleMessage(mes)
		case <-d.ctx.Done():
			return
		}
	}
}

type worker struct {
	workId        int
	ctx           context.Context
	workerPool    chan *worker
	work          chan workFunc
	handleMessage chan Message
}

func newWorker(workId int, ctx context.Context, pool chan *worker, hMessage chan Message) *worker {
	return &worker{
		workId:        workId,
		ctx:           ctx,
		workerPool:    pool,
		work:          make(chan workFunc),
		handleMessage: hMessage,
	}
}

// 将需要执行的工作逻辑，放入工作池中，等待工作者处理
func (w *worker) submit(f workFunc) {
	w.work <- f
}

// 将自身放入工人池中，获取一份工作池中的工作执行
func (w *worker) run() {
	for {
		w.workerPool <- w // @workRun ： 有缓冲这步非阻塞，不会影响下面的退出
		select {
		case wf := <-w.work: // 在自己的工作池中，获取一个调度器给的工作，并执行
			b, err := wf()
			w.handleMessage <- Message{b, err} // 将执行反馈结果放入反馈通道中，在调度器中取出
			//这里处理一下过程的记录
		case <-w.ctx.Done():
			return
		}
	}
}
