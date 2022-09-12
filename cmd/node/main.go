package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/mastrogiovanni/mychain/internal/node"
)

func main() {

	senderFlag := flag.Bool("sender", false, "if node is also a sender")
	flag.Parse()

	log.Println("Hello Node!")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := node.NewNode(false, 0, "stocazzo")
	defer n.Close()

	n.Run(ctx)

	if *senderFlag {

		for {
			peers := n.ListPeers()
			log.Println("Peers:", peers)
			if len(peers) > 0 {
				break
			}
			time.Sleep(1 * time.Second)
		}

		log.Println("Sending...")
		err := n.Publish("Stocazzo sono io!!!!")
		if err != nil {
			panic(err)
		}
	}

	select {}
}
