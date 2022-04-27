package main

import (
	"encoding/json"
	"log"
	"spool"
	"time"
)

func main() {
	poolWithFuncLoad()
}

func poolLoad() {
	p := spool.NewPool(5)
	for i := 0; i < 10; i++ {
		p.Submit(func() ([]byte, error) {
			log.Println("123:")
			return json.Marshal("shitingbao")
		})
	}
	time.Sleep(time.Second * 2)
	p.Release()
	time.Sleep(time.Second * 3)
}

func poolWithFuncLoad() {

	p := spool.NewPoolWithFunc(3, func(param interface{}) ([]byte, error) {
		log.Println("123:", param)
		return json.Marshal("shitingbao")
	})
	for i := 0; i < 10; i++ {
		p.Submit(i)
	}
	time.Sleep(time.Second * 2)
	p.Release()
	time.Sleep(time.Second * 3)
}
