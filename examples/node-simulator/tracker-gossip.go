package main

import (
	"log"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mastrogiovanni/mychain/internal/node"
)

type GossipTracker struct {
	node *node.Node
}

func (t GossipTracker) id() string {
	longID := t.node.ID().Pretty()
	return longID[(len(longID) - 6):]
}

// AddPeer is invoked when a new peer is added.
func (t GossipTracker) AddPeer(p peer.ID, proto protocol.ID) {
	log.Printf("(%s) Add Peer: [%s] (protocol: %s)\n", t.id(), p.Pretty(), proto)
}

// RemovePeer is invoked when a peer is removed.
func (t GossipTracker) RemovePeer(p peer.ID) {
	log.Printf("(%s) Add Peer: [%s]\n", t.id(), p.Pretty())
}

// Join is invoked when a new topic is joined
func (t GossipTracker) Join(topic string) {
}

// Leave is invoked when a topic is abandoned
func (t GossipTracker) Leave(topic string) {
}

// Graft is invoked when a new peer is grafted on the mesh (gossipsub)
func (t GossipTracker) Graft(p peer.ID, topic string) {
	log.Printf("(%s) Graft: %s (topic: '%s')\n", t.id(), p.Pretty(), topic)
}

// Prune is invoked when a peer is pruned from the message (gossipsub)
func (t GossipTracker) Prune(p peer.ID, topic string) {
	log.Printf("(%s) Prune: %s (topic: '%s')\n", t.id(), p.Pretty(), topic)
}

// ValidateMessage is invoked when a message first enters the validation pipeline.
func (t GossipTracker) ValidateMessage(msg *pubsub.Message) {
}

// DeliverMessage is invoked when a message is delivered
func (t GossipTracker) DeliverMessage(msg *pubsub.Message) {
}

// RejectMessage is invoked when a message is Rejected or Ignored.
// The reason argument can be one of the named strings Reject*.
func (t GossipTracker) RejectMessage(msg *pubsub.Message, reason string) {
}

// DuplicateMessage is invoked when a duplicate message is dropped.
func (t GossipTracker) DuplicateMessage(msg *pubsub.Message) {
}

// ThrottlePeer is invoked when a peer is throttled by the peer gater.
func (t GossipTracker) ThrottlePeer(p peer.ID) {
	log.Printf("(%s) ThrottlePeer: %s\n", t.id(), p.Pretty())
}

// RecvRPC is invoked when an incoming RPC is received.
func (t GossipTracker) RecvRPC(rpc *pubsub.RPC) {
}

// SendRPC is invoked when a RPC is sent.
func (t GossipTracker) SendRPC(rpc *pubsub.RPC, p peer.ID) {
}

// DropRPC is invoked when an outbound RPC is dropped, typically because of a queue full.
func (t GossipTracker) DropRPC(rpc *pubsub.RPC, p peer.ID) {
}

// UndeliverableMessage is invoked when the consumer of Subscribe is not reading messages fast enough and
// the pressure release mechanism trigger, dropping messages.
func (t GossipTracker) UndeliverableMessage(msg *pubsub.Message) {
}
