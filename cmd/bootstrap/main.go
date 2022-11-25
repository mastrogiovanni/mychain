package main

import (
	"flag"
	"fmt"
	mrand "math/rand"
	"os"
	"strings"

	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/core/crypto"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
)

type tracer struct{}

func (tracer) Trace(evt *holepunch.Event) {
	fmt.Printf("%+v\n", evt)
}

// AllowReserve returns true if a reservation from a peer with the given peer ID and multiaddr
// is allowed.
func (tracer) AllowReserve(p peer.ID, a ma.Multiaddr) bool {
	fmt.Printf("AllowReserve %s, %+v\n", p, a)
	return true
}

// AllowConnect returns true if a source peer, with a given multiaddr is allowed to connect
// to a destination peer.
func (tracer) AllowConnect(src peer.ID, srcAddr ma.Multiaddr, dest peer.ID) bool {
	fmt.Printf("AllowReserve src: %s, srcAddr: %+v, dest: %s\n", src, srcAddr, dest)
	return true
}

func main() {

	log.SetAllLoggers(log.LevelDebug)

	help := flag.Bool("help", false, "Display Help")

	flag.Parse()

	if *help {
		fmt.Printf("This is a simple bootstrap node for kad-dht application using libp2p\n\n")
		fmt.Printf("Usage: \n   Run './bootnode'\nor Run './bootnode -host [host] -port [port]'\n")
		os.Exit(0)
	}

	r := mrand.New(mrand.NewSource(42))

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	t := tracer{}
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Identity(prvKey),
		// libp2p.EnableRelay(),
		libp2p.EnableHolePunching(holepunch.WithTracer(t)),
	)
	if err != nil {
		panic(err)
	}

	sub, err := host.EventBus().Subscribe(event.WildcardSubscription)
	if err != nil {
		panic(err)
	}
	go func() {
		defer sub.Close()
		for e := range sub.Out() {
			fmt.Printf("%+v\n", e)
		}
	}()

	_, err = relay.New(host, relay.WithACL(t))
	if err != nil {
		fmt.Printf("Failed to instantiate the relay: %v", err)
		panic(err)
	}

	for _, addr := range host.Addrs() {
		if !strings.Contains(addr.String(), "127.0.0.1") {
			fmt.Printf("%s/p2p/%s\n", addr.String(), host.ID())
		}
	}

	select {}

}
