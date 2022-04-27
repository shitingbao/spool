package spool

import "log"

type option struct {
	HandlePoolMessage PoolMessage
}

type Option func(*option)

// 传入自己实现的 HandleMessage 对象
func WithHandlePoolMessage(pm PoolMessage) Option {
	return func(o *option) {
		o.HandlePoolMessage = pm
	}
}

type Message struct {
	Body []byte
	Err  error
}

type PoolMessage interface {
	HandleMessage(Message) error
}

type defaultPoolMessage struct{}

// 默认的参数接受，不做任何处理
func (d *defaultPoolMessage) HandleMessage(e Message) error {
	log.Println("body==:", string(e.Body), "||", e.Err)
	return nil
}
