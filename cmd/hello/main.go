package main

import (
	"encoding/hex"
	"log"
	"math/rand"

	"github.com/mastrogiovanni/mychain/internal/account"
	"github.com/mastrogiovanni/mychain/internal/chain"
)

func main() {

	pool := make([]*chain.Block, 0)

	log.Println("Hello World!")

	originator := account.NewEd25519Account()

	ch := chain.NewChain(originator)

	genesis := ch.Genesis()
	pool = append(pool, genesis)

	log.Println("Genesys:", genesis)

	acc := account.NewEd25519Account()

	block, err := ch.NewBlock([][]byte{genesis.Signature}, []byte("payload"), acc)
	if err != nil {
		panic(err)
	}

	pool = append(pool, block)

	const NumBlocks = 2

	for len(pool) < 10000 {

		original := make([]*chain.Block, 0, NumBlocks)
		for numBlocks := 0; numBlocks < NumBlocks; numBlocks++ {
			blockPool := pool[rand.Intn(len(pool))]
			original = append(original, blockPool)
		}

		nextBlock, err := ch.NewBlockFromBlocks(original, []byte("payload"), acc)
		if err != nil {
			panic(err)
		}

		pool = append(pool, nextBlock)
		log.Printf("Block: %s, #%d", nextBlock.String(), len(pool))

	}

	for _, block := range pool {
		ok, err := ch.Verify(block)
		if err != nil {
			log.Printf("Block: %s", block.String())
			log.Println(err)
			panic(err)
		}
		log.Printf("%s => %v", hex.EncodeToString(block.Signature), ok)
	}

}
