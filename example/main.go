package example

import (
	"log"
	"spool"
	"time"
)

func poolLoad() {
	p := spool.NewPool(5)
	for i := 0; i < 10; i++ {
		p.Submit(func() {
			log.Println("123")
		})
	}
	time.Sleep(time.Second * 2)
	p.Release()
	time.Sleep(time.Second * 3)
}

func poolWithFuncLoad() {

	p := spool.NewPoolWithFunc(3, func(param interface{}) {
		log.Println("123:", param)
	})
	for i := 0; i < 10; i++ {
		p.Submit(i)
	}
	time.Sleep(time.Second * 2)
	p.Release()
	time.Sleep(time.Second * 3)
}
