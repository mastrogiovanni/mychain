package main

import (
	"flag"
	"fmt"
	"log"
	mrand "math/rand"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"

	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
)

type tracer struct{}

func (tracer) Trace(evt *holepunch.Event) {
	fmt.Printf("%+v\n", evt)
}

func main() {

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
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(holepunch.WithTracer(t)),
	)
	if err != nil {
		panic(err)
	}
	_, err = relay.New(host)
	if err != nil {
		log.Printf("Failed to instantiate the relay: %v", err)
		panic(err)
	}

	for _, addr := range host.Addrs() {
		if !strings.Contains(addr.String(), "127.0.0.1") {
			fmt.Printf("%s/p2p/%s\n", addr.String(), host.ID())
		}
	}

	select {}

}
