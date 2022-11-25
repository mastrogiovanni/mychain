package main

import (
	"context"
	"encoding/hex"
	"flag"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/mastrogiovanni/mychain/internal/account"
	"github.com/mastrogiovanni/mychain/internal/chain"
	"github.com/mastrogiovanni/mychain/internal/node"
)

type Config struct {
	peers   string
	spammer bool
	genesys string
}

type MyChainNode struct {
	account *account.Ed25519Account
	chain   *chain.Chain
	peers   []string
	spammer bool
	node    *node.Node
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewMyChainNode(config *Config) *MyChainNode {
	n := &MyChainNode{}
	n.account = account.NewEd25519Account()
	n.spammer = config.spammer

	// Load genesys block
	var genesys []byte
	if config.genesys != "" {
		data, err := hex.DecodeString(config.genesys)
		if err != nil {
			panic(err)
		}
		genesys = data
	}
	n.chain = chain.NewChain(n.account, genesys)

	ctx, cancel := context.WithCancel(context.Background())
	n.ctx = ctx
	n.cancel = cancel

	// Get peer list
	peers := make([]string, 0)
	addrs := strings.Split(config.peers, ",")
	for _, addr := range addrs {
		s := strings.TrimSpace(addr)
		if s != "" {
			peers = append(peers, s)
		}
	}
	n.peers = peers

	// Create physical layer
	node, err := node.NewNode(
		node.WithNetworkId("test-node"),
	)
	if err != nil {
		panic(err)
	}
	n.node = node

	err = node.Start(ctx)
	if err != nil {
		panic(err)
	}

	// Pool of blocks
	pool := make([]*chain.Block, 0)
	genesis := n.chain.Genesis()
	pool = append(pool, genesis)

	// Connection to peers
	if len(peers) > 0 {
		log.Println("Connecting to peers...")
		for _, addr := range peers {
			err = node.Connect(addr)
			if err != nil {
				log.Printf("Error connecting with '%s': %s", addr, err)
				panic(err)
			} else {
				log.Println("connected to", addr)
			}
		}
	}

	// Rx messages
	go func() {
		log.Println("Start receiving messages")
		for {
			msg, ok := <-node.GetMessages()
			if ok {
				if msg.GetFrom() != node.ID() {
					block := &chain.Block{}
					err := block.Deserialize(msg.GetData())
					if err != nil {
						log.Printf("Message discarded from %s: %s", msg.GetFrom(), err)
					} else {
						err := n.chain.Append(block)
						if err != nil {
							log.Printf("Message discarded from %s: %s", msg.GetFrom(), err)
							continue
						}
						log.Printf("Chain size is now: %d", n.chain.Size())
						pool = append(pool, block)
					}
				}
			} else {
				return
			}
		}
	}()

	if config.spammer {
		go func() {
			log.Println("Starting Spammer")
			NumBlocks := 2
			for {
				time.Sleep(500 * time.Millisecond)
				original := make([]*chain.Block, 0, NumBlocks)
				for numBlocks := 0; numBlocks < NumBlocks; numBlocks++ {
					blockPool := pool[rand.Intn(len(pool))]
					original = append(original, blockPool)
				}
				nextBlock, err := n.chain.NewBlockFromBlocks(original, []byte("payload"), n.account)
				if err != nil {
					panic(err)
				}
				pool = append(pool, nextBlock)
				log.Printf("Block: %s, #%d", nextBlock.String(), len(pool))
				data, err := nextBlock.Serialize()
				if err != nil {
					panic(err)
				}
				node.Broadcast(data)
			}
		}()
	}

	return n
}

func parseFlags() *Config {
	config := &Config{}
	flag.StringVar(&config.peers, "peers", "", "List of peers separated by column")
	flag.BoolVar(&config.spammer, "spammer", false, "True if you want to spam each second")
	flag.StringVar(&config.genesys, "genesys", "", "Genesys block")
	flag.Parse()
	return config
}

func main() {
	config := parseFlags()
	node := NewMyChainNode(config)
	block := node.chain.Genesis()
	data, _ := block.Serialize()

	// log.Println(hex.EncodeToString(data))

	log.Println("Il you want to join:")
	for _, addr := range node.node.Addresses() {
		log.Printf("go run cmd/node/main.go -peers %s -genesys %s -spammer true", addr, hex.EncodeToString(data))
	}

	select {}
}
