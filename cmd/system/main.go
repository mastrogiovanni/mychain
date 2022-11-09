package main

import (
	"context"
	"log"
	"time"

	"github.com/mastrogiovanni/mychain/internal/node"
)

func test() {

	defer log.Println("STOCAZZOOOOOO")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 3; i++ {
		rcv := node.NewNode(false, 0, "stocazzo")
		defer rcv.Close()
		rcv.Run(ctx)
	}

	sender := node.NewNode(false, 0, "stocazzo")
	defer sender.Close()
	sender.Run(ctx)

	for i := 0; i < 10; i++ {
		peers := sender.ListPeers()
		log.Println("Peers:", peers)
		if len(peers) > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	log.Println("Sending...")
	err := sender.Publish("Stocazzo sono io!!!!")
	if err != nil {
		panic(err)
	}

	select {}

}

func main() {
	test()
}
