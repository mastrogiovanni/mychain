package node

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const NodeMessagesQueueSize = 128

// NodeMessage gets converted to/from JSON and sent in the body of pubsub messages.
type NodeMessage struct {
	Message  string
	SenderID string
}

type Node struct {
	networkId    string
	privateKey   crypto.PrivKey
	publicKey    crypto.PubKey
	host         host.Host
	topic        *pubsub.Topic
	pubSub       *pubsub.PubSub
	subscription *pubsub.Subscription
	context      context.Context
	messages     chan *NodeMessage
	self         peer.ID
}

func NewNode(insecure bool, randseed int64, networkId string) *Node {

	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it at least
	// to obtain a valid host ID.
	privKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Identity(privKey),
		libp2p.DisableRelay(),
	}

	if insecure {
		opts = append(opts, libp2p.NoSecurity)
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	return &Node{
		privateKey: privKey,
		publicKey:  pubKey,
		host:       host,
		networkId:  networkId,
	}

}

func (n *Node) Close() {
	n.subscription.Cancel()
	n.host.Close()
}

func (n *Node) HandlePeerFound(pi peer.AddrInfo) {
	log.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.host.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func (n *Node) setupDiscovery() error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(n.host, n.networkId, n)
	return s.Start()
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (n *Node) readLoop() {
	log.Println("Start reading messages")
	for {
		msg, err := n.subscription.Next(n.context)
		if err != nil {
			close(n.messages)
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == n.self {
			continue
		}
		cm := new(NodeMessage)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel
		log.Printf("Received msg: %+v", cm)
		n.messages <- cm
	}
}

// Publish sends a message to the pubsub topic.
func (n *Node) Publish(message string) error {
	m := NodeMessage{
		Message:  message,
		SenderID: n.self.Pretty(),
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	log.Printf("Sending: %s", string(msgBytes))
	return n.topic.Publish(n.context, msgBytes)
}

func (n *Node) ListPeers() []peer.ID {
	return n.pubSub.ListPeers(n.networkId)
}

func (n *Node) joinChatRoom(ctx context.Context) {

	// join the pubsub topic
	topic, err := n.pubSub.Join(n.networkId)
	if err != nil {
		panic(err)
	}
	n.topic = topic

	log.Println("Joining network:", n.networkId)

	// and subscribe to it
	sub, err := n.topic.Subscribe()
	if err != nil {
		panic(err)
	}

	n.subscription = sub
	n.self = n.host.ID()

	log.Printf("Host ID: %v", n.self)
	go n.readLoop()

}

func (n *Node) Run(ctx context.Context) {

	// create a new PubSub service using the GossipSub router
	pubSub, err := pubsub.NewGossipSub(ctx, n.host)
	if err != nil {
		panic(err)
	}
	n.pubSub = pubSub

	// setup local mDNS discovery
	if err := n.setupDiscovery(); err != nil {
		panic(err)
	}

	n.context = ctx
	n.messages = make(chan *NodeMessage, NodeMessagesQueueSize)

	// join the chat room
	n.joinChatRoom(ctx)

}
