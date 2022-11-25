package main

import (
	"context"
	"fmt"

	nat "github.com/libp2p/go-libp2p-nat"
)

func main() {
	nat, err := nat.DiscoverNAT(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", nat)
	fmt.Printf("%+v\n", nat.Mappings())
}
