package main

import (
	"log"
	"time"
)

func main() {
	ch := make(chan int)
	go func() {
		time.Sleep(time.Second)
		ch <- 23
		time.Sleep(time.Second)
		close(ch)
	}()
	for {
		select {
		case value := <-ch:
			log.Println(value)
		case <-ch:
			log.Println("Closed")
			return
		}
	}
}
