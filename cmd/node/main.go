package main

import (
	"context"
	"fmt"
	"log"
	mrand "math/rand"
	"time"

	"github.com/mastrogiovanni/mychain/internal/node"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func createNodes(ctx context.Context, count int, networkId string) []*node.Node {
	nodes := make([]*node.Node, 0)
	for i := 0; i < count; i++ {
		node, err := node.NewNode(
			node.WithNetworkId(networkId),
		)
		if err != nil {
			panic(err)
		}
		if i == count-1 {
			tracker := &GossipTracker{node: node}
			err = node.Start(ctx, pubsub.WithRawTracer(tracker))
			if err != nil {
				panic(err)
			}
			nodes = append(nodes, node)
		} else {
			err = node.Start(ctx)
			if err != nil {
				panic(err)
			}
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func stopNodes(nodes []*node.Node) {
	log.Println("Stopping all nodes")
	for _, node := range nodes {
		node.Stop()
	}
}

func lineConnected(ctx context.Context, nodes []*node.Node) {
	log.Printf("Connecting %d nodes", len(nodes))
	for i := 0; i < len(nodes)-1; i++ {
		nodes[i].Connect(nodes[i+1].Addresses()[0])
	}
}

func fullConnected(ctx context.Context, nodes []*node.Node) {
	log.Printf("Connecting %d nodes", len(nodes))
	for i := 0; i < len(nodes); i++ {
		for j := 0; j < i; j++ {
			if !(i == len(nodes)-1 && j == 0) {
				nodes[j].Connect(nodes[i].Addresses()[0])
			}
		}
	}
}

func randomConnection(ctx context.Context, nodes []*node.Node, edges int) {
	log.Printf("Creating %d connections", edges)
	for i := 0; i <= edges; {
		j := mrand.Intn(len(nodes))
		x := mrand.Intn(len(nodes))
		if i != x {
			nodes[j].Connect(nodes[x].Addresses()[0])
			i++
		}
	}
}

func startBroadcast(ctx context.Context, nodes []*node.Node, index int) {
	go func() {
		for i := 0; ; i++ {
			// log.Printf("(%s) Bcast: PRIMO %d", nodes[0].host.ID().Pretty(), i)
			nodes[index].Broadcast([]byte(fmt.Sprintf("<%d>", i)))
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func startAllBroadcast(ctx context.Context, nodes []*node.Node) {
	go func() {
		for counter := 0; ; counter++ {
			j := mrand.Intn(len(nodes))
			if nodes[j] == nil {
				continue
			}
			log.Printf("(%d) Bcast: %d", j, counter)
			nodes[j].Broadcast([]byte(fmt.Sprintf("<%d>", counter)))
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func startReceiving(ctx context.Context, nodes []*node.Node, index int) {
	go func() {
		for {
			select {
			case msg, ok := <-nodes[index].GetMessages():
				if ok {
					if msg.GetFrom() != nodes[index].ID() {
						log.Printf("%d) Recv: %s (from: %s)", index, string(msg.GetData()), msg.GetFrom())
						// log.Printf("%+v", nodes[index].pubsub.ListPeers(nodes[index].networkId))
					}
				} else {
					return
				}
			}
		}
	}()
}

func startAllReceiving(ctx context.Context, nodes []*node.Node) {
	for i := 0; i < len(nodes); i++ {
		startReceiving(ctx, nodes, i)
	}
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func deleteNodes(ctx context.Context, nodes []*node.Node, duration time.Duration, excludes ...int) {
	go func() {
		a := makeRange(0, len(nodes)-1)
		for i := 0; i < len(excludes); i++ {
			a[excludes[i]] = -1
		}
		mrand.Seed(time.Now().UnixNano())
		mrand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
		for i := 0; i < len(nodes); i++ {
			time.Sleep(duration)
			index := a[i]
			if index > -1 && nodes[index] != nil {
				log.Printf("Disconnecting %d", index)
				nodes[index].Stop()
				nodes[index] = nil
			}
		}
	}()
}

func main() {
	max := 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("Creating %d nodes", max)

	nodes := createNodes(ctx, max, "test")
	defer stopNodes(nodes)

	lineConnected(ctx, nodes)
	// fullConnected(ctx, nodes)
	// randomConnection(ctx, nodes, max*(max-1))

	log.Println("Starting protocol")

	startAllBroadcast(ctx, nodes)
	// startBroadcast(ctx, nodes, 0)

	startAllReceiving(ctx, nodes)
	// startReceiving(ctx, nodes, max-1)

	deleteNodes(ctx, nodes, 200*time.Millisecond)

	select {}

}
