package main

import (
	"log"
	"spool"
	"time"
)

func main() {
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
