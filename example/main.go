package main

import (
	"errors"
	"log"
	"spool"
	"time"
)

type handle struct {
	Num int
}

func (h *handle) HandleMessage(m spool.Message) error {
	h.Num++
	log.Println("num:", h.Num, " that your body:", string(m.Body), " and err:", m.Err)
	return nil
}

func main() {
	poolWithFuncLoad()
}

func poolLoad() {
	h := &handle{}
	option := spool.WithHandlePoolMessage(h)
	p := spool.NewPool(5, option)
	for i := 0; i < 10; i++ {
		p.Submit(func() (spool.WorkResult, error) {
			return "spool here", nil
		})
	}
	time.Sleep(time.Second * 2)
	p.Release()
	time.Sleep(time.Second * 3)
}

func poolWithFuncLoad() {
	h := &handle{}
	option := spool.WithHandlePoolMessage(h)
	p := spool.NewPoolWithFunc(3, func(param interface{}) (spool.WorkResult, error) {
		return "spool", errors.New("this is a error;because err not nil,body is nil")
	}, option)
	for i := 0; i < 10; i++ {
		p.Submit(i)
	}
	time.Sleep(time.Second * 2)
	p.Release()
	time.Sleep(time.Second * 3)
}
