package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-msgio"
	"github.com/mastrogiovanni/mychain/internal/account"
	"github.com/mastrogiovanni/mychain/internal/chain"
	"github.com/mastrogiovanni/mychain/internal/node"
	"github.com/mastrogiovanni/mychain/internal/stack"
)

const (
	NumTipsBlocks int = 2
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
	layer   *stack.Layer
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
		node.WithNetworkId("/utxo/protocol/0.0.1"),
	)
	if err != nil {
		panic(err)
	}
	n.node = node

	currentLayer := stack.NewLayer(func(direction stack.Direction, data []byte, layer *stack.Layer) {
		if direction == stack.Down {
			layer.SendDown(data)
		}
	})

	lowerLayer := stack.NewLayer(func(direction stack.Direction, data []byte, layer *stack.Layer) {
		// I'm sending data
		if direction == stack.Down {
			original := make([]*chain.Block, 0, NumTipsBlocks)
			for numBlocks := 0; numBlocks < NumTipsBlocks; numBlocks++ {
				randomTip := n.chain.GetOneOfLastBlocksAtRandom(10)
				original = append(original, randomTip)
			}
			nextBlock, err := n.chain.NewBlockFromBlocks(original, data, n.account)
			if err != nil {
				panic(err)
			}
			log.Printf("Block: %s, #%d", nextBlock.String(), n.chain.Size())
			data, err := nextBlock.Serialize()
			if err != nil {
				panic(err)
			}
			node.Broadcast(data)
		}
	})

	currentLayer.AttachLower(lowerLayer)

	n.layer = currentLayer

	n.node.SetStreamHandler("/snapshot/0.0.1", func(stream network.Stream) {
		log.Println("Preparing to send snapshot")
		writer := msgio.NewWriter(stream)
		for _, block := range n.chain.Snapshot() {
			log.Printf("Sending: %s", hex.EncodeToString(block.Signature))
			data, err := block.Serialize()
			if err != nil {
				log.Printf("Impossible to upload snapshot's block (%s): %s", hex.EncodeToString(block.Signature), err)
				continue
			}
			writer.WriteMsg(data)
		}
		stream.Close()
	})

	err = node.Start(ctx)
	if err != nil {
		panic(err)
	}

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

	time.Sleep(5 * time.Second)

	// Start download snapshot
	for _, peerID := range n.node.GetPeers() {
		log.Println("Start downloading snapshot from:", peerID)
		stream, err := n.node.NewStream(peerID, "/snapshot/0.0.1")
		if err != nil {
			panic(err)
		}
		reader := msgio.NewReader(stream)
		for {
			msg, err := reader.ReadMsg()
			if err != nil {
				log.Println(err)
				break
			}
			block, err := chain.NewBlock(msg)
			if err != nil {
				log.Println(err)
				break
			}
			err = n.chain.Append(block)
			if err != nil {
				log.Printf("Block discarded")
				continue
			}
			log.Println("Found block:", hex.EncodeToString(block.Signature))
		}
		stream.Close()
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
						n.layer.SendUp(block.Payload)
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
			for {
				time.Sleep(500 * time.Millisecond)
				n.layer.SendDown([]byte(fmt.Sprintf("msg from: %s", n.node.ID())))
			}
		}()
	}

	return n
}

func (n *MyChainNode) Layer() *stack.Layer {
	return n.layer
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

	log.Println("If want to join:")
	for _, addr := range node.node.Addresses() {
		log.Printf("go run cmd/node/main.go -peers %s -genesys %s -spammer true", addr, hex.EncodeToString(data))
	}

	for _, addr := range node.node.Addresses() {
		log.Printf("I'm %s", addr)
	}

	node.Layer().AttachUpper(stack.NewLayer(func(direction stack.Direction, data []byte, layer *stack.Layer) {
		log.Println(direction, string(data))
	}))

	select {}
}
