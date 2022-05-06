package spool

import (
	"context"
	"log"
	"time"
)

type option struct {
	HandlePoolMessage PoolMessage
	DeadTimeDuration  time.Duration // 设置整个池执行的死亡时间
}

type Option func(*option)

// WithHandlePoolMessage 传入自己实现的 HandleMessage 对象
func WithHandlePoolMessage(pm PoolMessage) Option {
	return func(o *option) {
		o.HandlePoolMessage = pm
	}
}

// WithDeadTimeDuration 写入池的执行死亡时间
// 注意这个时间并不会阻止已经开始的逻辑
func WithDeadTimeDuration(td time.Duration) Option {
	return func(o *option) {
		o.DeadTimeDuration = td
	}
}

func (o *option) getCollectiveContext() (context.Context, context.CancelFunc) {
	if o.DeadTimeDuration == 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), o.DeadTimeDuration)
}

type Message struct {
	Body []byte
	Err  error
}

// 用于接收反馈内容的接口
// 可自行定义 HandleMessage 的内容来处理反馈内容
type PoolMessage interface {
	HandleMessage(Message) error
}

type defaultPoolMessage struct{}

// 默认的参数接受，不做任何处理
func (d *defaultPoolMessage) HandleMessage(e Message) error {
	log.Println("body==:", string(e.Body), "||", e.Err)
	return nil
}
