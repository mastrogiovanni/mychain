package main

import (
	"context"
	"fmt"
	mrand "math/rand"

	"github.com/ipfs/go-log/v2"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	ma "github.com/multiformats/go-multiaddr"
)

type tracer struct{}

func (tracer) Trace(evt *holepunch.Event) {
	fmt.Printf("Event: %+v\n", evt)
}

// func createBootstrapNode(seed int64) (host.Host, error) {
// 	// Creates a new RSA key pair for this host.
// 	r := mrand.New(mrand.NewSource(seed))
// 	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
// 	if err != nil {
// 		panic(err)
// 	}
// 	t := tracer{}
// 	node, err := libp2p.New(
// 		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
// 		libp2p.Identity(prvKey),
// 		libp2p.EnableHolePunching(holepunch.WithTracer(t)),
// 	)
// 	if err != nil {
// 		log.Printf("Failed to create unreachable1: %v", err)
// 		return nil, err
// 	}
// 	_, err = relay.New(node)
// 	if err != nil {
// 		log.Printf("Failed to instantiate the relay: %v", err)
// 		return nil, err
// 	}
// 	return node, nil
// }

func createAndRunNode(seed int64) (host.Host, error) {
	// Creates a new RSA key pair for this host.
	r := mrand.New(mrand.NewSource(seed))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}
	t := tracer{}
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Identity(prvKey),
		// libp2p.EnableRelay(),
		libp2p.EnableHolePunching(holepunch.WithTracer(t)),
	)
	if err != nil {
		fmt.Printf("Failed to create unreachable1: %v\n", err)
		return nil, err
	}

	sub, err := node.EventBus().Subscribe(event.WildcardSubscription)
	if err != nil {
		panic(err)
	}
	go func() {
		defer sub.Close()
		for e := range sub.Out() {
			fmt.Printf("%+v\n", e)
		}
	}()

	return node, nil
}

func main() {

	log.SetAllLoggers(log.LevelDebug)

	n1, err := createAndRunNode(0)
	if err != nil {
		panic(err)
	}
	n2, err := createAndRunNode(1)
	if err != nil {
		panic(err)
	}

	// boostrap, err := createBootstrapNode(2)
	// if err != nil {
	// 	panic(err)
	// }

	ctx := context.Background()

	bootstrapAddr, err := peer.AddrInfoFromString("/ip4/95.217.213.204/tcp/35535/p2p/QmQi91gEajE29jixXi95sw6BtvgsArbCFX4jHGDekPrKYq")
	if err != nil {
		panic(err)
	}

	// bootstrapAddr := peer.AddrInfo{
	// 	ID:    boostrap.ID(),
	// 	Addrs: boostrap.Addrs(),
	// }

	err = n1.Connect(ctx, *bootstrapAddr)
	if err != nil {
		panic(err)
	}

	err = n2.Connect(ctx, *bootstrapAddr)
	if err != nil {
		panic(err)
	}

	// Now, to test the communication, let's set up a protocol handler on unreachable2
	n2.SetStreamHandler("/customprotocol", func(s network.Stream) {
		log.Println("Awesome! We're now communicating via the relay!")
		s.Close()
	})

	// Hosts that want to have messages relayed on their behalf need to reserve a slot
	// with the circuit relay service host
	// As we will open a stream to unreachable2, unreachable2 needs to make the
	// reservation
	_, err = client.Reserve(context.Background(), n2, *bootstrapAddr)
	if err != nil {
		log.Printf("n2 failed to receive a relay reservation from bootstrap. %v", err)
		return
	}

	// Now create a new address for unreachable2 that specifies to communicate via
	// relay1 using a circuit relay
	relayaddr, err := ma.NewMultiaddr("/p2p/" + bootstrapAddr.ID.String() + "/p2p-circuit/p2p/" + n2.ID().String())
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Now let's attempt to connect the hosts via the relay node")

	// Open a connection to the previously unreachable host via the relay address
	n2ViaBootstrap := peer.AddrInfo{
		ID:    n2.ID(),
		Addrs: []ma.Multiaddr{relayaddr},
	}
	if err := n1.Connect(context.Background(), n2ViaBootstrap); err != nil {
		log.Printf("Unexpected error here. Failed to connect unreachable1 and unreachable2: %v", err)
		return
	}

	log.Println("Yep, that worked!")

}
