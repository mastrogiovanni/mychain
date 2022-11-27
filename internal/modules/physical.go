package modules

import (
	"bytes"
	"context"
	"encoding/hex"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio"
	"github.com/mastrogiovanni/mychain/internal/account"
	"github.com/mastrogiovanni/mychain/internal/chain"
	"github.com/mastrogiovanni/mychain/internal/node"
	"github.com/mastrogiovanni/mychain/internal/stack"
)

const (
	NumTipsBlocks      int    = 2
	NumOfBlocksForTips int    = 10
	ProtocolSnapshot   string = "/snapshot/0.0.1"
	ProtocolUtxo       string = "/utxo/protocol/0.0.1"
)

type PhysicalLayer struct {
	*stack.Layer
	*node.Node

	account *account.Ed25519Account
	chain   *chain.Chain
	peers   []string
	ctx     context.Context
	cancel  context.CancelFunc
}

type PhysicalLayerConfig struct {
	Peers   string
	Genesys string
}

func NewPhysicalLayer(config *PhysicalLayerConfig) (*PhysicalLayer, error) {
	n := &PhysicalLayer{}
	n.Layer = stack.NewLayer(n.HandleMsg)
	n.account = account.NewEd25519Account()

	n.loadOrCreateGenesysBlock(config)
	n.createContext()
	n.loadPeerList(config)

	// Create physical layer
	node, err := node.NewNode(
		node.WithNetworkId(ProtocolUtxo),
	)
	if err != nil {
		return nil, err
	}
	n.Node = node

	n.Node.SetStreamHandler(protocol.ID(ProtocolSnapshot), func(stream network.Stream) {
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

	// , pubsub.WithRawTracer(n)
	err = node.Start(n.ctx)
	if err != nil {
		panic(err)
	}

	// Connect to peers
	if len(n.peers) > 0 {
		log.Println("Connecting to peers...")
		for _, addr := range n.peers {
			err = node.Connect(addr)
			if err != nil {
				log.Printf("Error connecting with '%s': %s", addr, err)
				return nil, err
			} else {
				log.Println("connected to", addr)
			}
		}

		// Start download snapshot
		n.waitForSnapshotFromFirstPeer()
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
						n.SendUp(block.Payload)
					}
				}
			} else {
				return
			}
		}
	}()

	return n, nil
}

func (p *PhysicalLayer) Close() {
	p.Node.Stop()
	p.cancel()
}

func (p *PhysicalLayer) HandleMsg(direction stack.Direction, data []byte, layer *stack.Layer) {
	if direction == stack.Down {
		original := make([]*chain.Block, 0, NumTipsBlocks)
		for len(original) < NumTipsBlocks {
			randomTip := p.chain.GetOneOfLastBlocksAtRandom(NumOfBlocksForTips)
			alreadyExists := false
			for _, item := range original {
				if !bytes.Equal(item.Signature, randomTip.Signature) {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				original = append(original, randomTip)
			}
		}
		nextBlock, err := p.chain.NewBlockFromBlocks(original, data, p.account)
		if err != nil {
			panic(err)
		}
		log.Printf("Block: %s, #%d", nextBlock.String(), p.chain.Size())
		data, err := nextBlock.Serialize()
		if err != nil {
			panic(err)
		}
		p.Node.Broadcast(data)
	}
}

func (p *PhysicalLayer) GetGenesysRaw() ([]byte, error) {
	block := p.chain.Genesis()
	data, err := block.Serialize()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (n *PhysicalLayer) waitForSnapshotFromFirstPeer() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if len(n.Node.GetPeers()) > 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	wg.Wait()

	for _, peerID := range n.Node.GetPeers() {
		log.Println("Start downloading snapshot from:", peerID)
		stream, err := n.Node.NewStream(peerID, protocol.ID(ProtocolSnapshot))
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
}

func (n *PhysicalLayer) createContext() {
	ctx, cancel := context.WithCancel(context.Background())
	n.ctx = ctx
	n.cancel = cancel
}

func (n *PhysicalLayer) loadPeerList(config *PhysicalLayerConfig) {
	peers := make([]string, 0)
	addrs := strings.Split(config.Peers, ",")
	for _, addr := range addrs {
		s := strings.TrimSpace(addr)
		if s != "" {
			peers = append(peers, s)
		}
	}
	n.peers = peers
}

func (n *PhysicalLayer) loadOrCreateGenesysBlock(config *PhysicalLayerConfig) {
	var genesys []byte
	if config.Genesys != "" {
		data, err := hex.DecodeString(config.Genesys)
		if err != nil {
			panic(err)
		}
		genesys = data
	}
	n.chain = chain.NewChain(n.account, genesys)
}
