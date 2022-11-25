package node

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	mrand "math/rand"
	"strings"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// NodeMessagesQueueSize is the number of incoming messages to buffer for each topic.
const NodeMessagesQueueSize = 128

type Node struct {
	networkId    string
	privateKey   crypto.PrivKey
	publicKey    crypto.PubKey
	host         host.Host
	topic        *pubsub.Topic
	pubSub       *pubsub.PubSub
	subscription *pubsub.Subscription
	messages     chan *pubsub.Message
	self         peer.ID
	context      context.Context
	cancel       context.CancelFunc
}

type NodeOpt func(*Node) error

func WithSeed(seed int64) NodeOpt {
	return func(n *Node) error {
		r := mrand.New(mrand.NewSource(seed))
		privKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
		if err != nil {
			return err
		}
		n.privateKey = privKey
		n.publicKey = pubKey
		return nil
	}
}

func WithNetworkId(networkId string) NodeOpt {
	return func(n *Node) error {
		n.networkId = networkId
		return nil
	}
}

func NewNode(opts ...NodeOpt) (*Node, error) {
	node := &Node{}

	// Load options
	for _, opt := range opts {
		err := opt(node)
		if err != nil {
			return nil, err
		}
	}

	// Default values
	if node.networkId == "" {
		node.networkId = "default-netowork-id"
	}
	if node.privateKey == nil || node.publicKey == nil {
		// Generate a key pair for this host. We will use it at least
		// to obtain a valid host ID.
		privKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
		if err != nil {
			return nil, err
		}
		node.privateKey = privKey
		node.publicKey = pubKey
	}

	// Create Host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Identity(node.privateKey),
	)
	if err != nil {
		return nil, err
	}
	node.host = host
	node.self = node.host.ID()

	return node, nil
}

func (n *Node) ID() peer.ID {
	return n.self
}

func (n *Node) Start(parentCtx context.Context, opts ...pubsub.Option) error {

	ctx, cancel := context.WithCancel(parentCtx)
	n.cancel = cancel
	n.context = ctx

	switch {
	case len(opts) == 1:
		pubSub, err := pubsub.NewGossipSub(n.context, n.host, opts[0])
		if err != nil {
			return err
		}
		n.pubSub = pubSub
	case len(opts) == 0:
		pubSub, err := pubsub.NewGossipSub(n.context, n.host)
		if err != nil {
			return err
		}
		n.pubSub = pubSub
	default:
		return fmt.Errorf("you can specify at most one option")
	}

	// // setup local mDNS discovery
	// if err := n.setupDiscovery(); err != nil {
	// 	panic(err)
	// }

	n.messages = make(chan *pubsub.Message, NodeMessagesQueueSize)

	topic, err := n.pubSub.Join(n.networkId)
	if err != nil {
		return err
	}
	n.topic = topic

	n.subscription, err = n.topic.Subscribe()
	if err != nil {
		return err
	}

	go n.readLoop()

	return nil

}

func (n *Node) Stop() {
	n.subscription.Cancel()
	n.topic.Close()
	n.host.Close()
	n.cancel()
}

func (n *Node) GetMessages() chan *pubsub.Message {
	return n.messages
}

func (n *Node) GetPeers() []peer.ID {
	return n.pubSub.ListPeers(n.networkId)
}

func (n *Node) NewStream(p peer.ID, pids protocol.ID) (network.Stream, error) {
	return n.host.NewStream(n.context, p, pids)
}

func (n *Node) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	n.host.SetStreamHandler(pid, handler)
}

func (n *Node) Connect(address string) error {
	addr, err := peer.AddrInfoFromString(address)
	if err != nil {
		return err
	}
	return n.host.Connect(n.context, *addr)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (n *Node) readLoop() {
	log.Println("Start reading messages")
	for {
		msg, err := n.subscription.Next(n.context)
		if err != nil {
			log.Println(err)
			close(n.messages)
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == n.self {
			continue
		}
		// send valid messages onto the Messages channel
		// log.Printf("Received msg: %+v", cm)
		n.messages <- msg
	}
}

// Publish sends a message to the pubsub topic.
func (n *Node) Broadcast(data []byte) error {
	n.topic.Relay()
	return n.topic.Publish(n.context, data)
}

func (n *Node) ListPeers() []peer.ID {
	return n.pubSub.ListPeers(n.networkId)
}

func (n *Node) DumpAddresses() {
	for _, addr := range n.Addresses() {
		if !strings.Contains(addr, "127.0.0.1") {
			log.Printf("Listen: %s", addr)
		}
	}
}

func (n *Node) Addresses() []string {
	list := make([]string, 0)
	for _, addr := range n.host.Addrs() {
		if !strings.Contains(addr.String(), "127.0.0.1") {
			// log.Printf("Addr: %s/p2p/%s", addr.String(), n.host.ID().String())
			list = append(list, fmt.Sprintf("%s/p2p/%s", addr.String(), n.host.ID().String()))
		}
	}
	return list
}

// func (n *Node) joinChatRoom(ctx context.Context) {

// 	// join the pubsub topic
// 	topic, err := n.pubSub.Join(n.networkId)
// 	if err != nil {
// 		panic(err)
// 	}
// 	n.topic = topic

// 	log.Println("Joining network:", n.networkId)

// 	// and subscribe to it
// 	sub, err := n.topic.Subscribe()
// 	if err != nil {
// 		panic(err)
// 	}

// 	n.subscription = sub
// 	n.self = n.host.ID()

// 	log.Printf("Host ID: %v", n.self)
// 	go n.readLoop()

// }

// func (n *Node) HandlePeerFound(pi peer.AddrInfo) {
// 	log.Printf("discovered new peer %s\n", pi.ID.Pretty())
// 	err := n.host.Connect(context.Background(), pi)
// 	if err != nil {
// 		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
// 	}
// }

// // setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// // This lets us automatically discover peers on the same LAN and connect to them.
// func (n *Node) setupDiscovery() error {

// 	ctx := context.Background()

// 	kademliaDHT, err := dht.New(ctx, n.host)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Bootstrap the DHT. In the default configuration, this spawns a Background
// 	// thread that will refresh the peer table every five minutes.
// 	log.Println("Bootstrapping the DHT")
// 	return kademliaDHT.Bootstrap(ctx)

// 	// // setup mDNS discovery to find local peers
// 	// s := mdns.NewMdnsService(n.host, n.networkId, n)
// 	// return s.Start()
// }
