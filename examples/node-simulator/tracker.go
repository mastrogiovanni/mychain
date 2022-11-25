package main

import (
	"log"
	"strings"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mastrogiovanni/mychain/internal/node"
)

type LogTracer struct {
	node *node.Node
}

func (t LogTracer) id() string {
	longID := t.node.ID().Pretty()
	return longID[(len(longID) - 6):]
}

// AddPeer is invoked when a new peer is added.
func (t LogTracer) AddPeer(p peer.ID, proto protocol.ID) {
	log.Printf("(%s) Add Peer: [%s] (protocol: %s)\n", t.id(), p.Pretty(), proto)
}

// RemovePeer is invoked when a peer is removed.
func (t LogTracer) RemovePeer(p peer.ID) {
	log.Printf("(%s) Add Peer: [%s]\n", t.id(), p.Pretty())
}

// Join is invoked when a new topic is joined
func (t LogTracer) Join(topic string) {
	log.Printf("(%s) Join '%s'\n", t.id(), topic)
}

// Leave is invoked when a topic is abandoned
func (t LogTracer) Leave(topic string) {
	log.Printf("(%s) Leave '%s'\n", t.id(), topic)
}

// Graft is invoked when a new peer is grafted on the mesh (gossipsub)
func (t LogTracer) Graft(p peer.ID, topic string) {
	log.Printf("(%s) Graft: %s (topic: '%s')\n", t.id(), p.Pretty(), topic)
}

// Prune is invoked when a peer is pruned from the message (gossipsub)
func (t LogTracer) Prune(p peer.ID, topic string) {
	log.Printf("(%s) Prune: %s (topic: '%s')\n", t.id(), p.Pretty(), topic)
}

// ValidateMessage is invoked when a message first enters the validation pipeline.
func (t LogTracer) ValidateMessage(msg *pubsub.Message) {
	log.Printf("(%s) ValidateMsg: %s\n", t.id(), msg.String())
}

// DeliverMessage is invoked when a message is delivered
func (t LogTracer) DeliverMessage(msg *pubsub.Message) {
	log.Printf("(%s) DeliverMsg: %s\n", t.id(), msg.String())
}

// RejectMessage is invoked when a message is Rejected or Ignored.
// The reason argument can be one of the named strings Reject*.
func (t LogTracer) RejectMessage(msg *pubsub.Message, reason string) {
	log.Printf("(%s) RejectMsg: %s (reson: %s)\n", t.id(), msg.String(), reason)
}

// DuplicateMessage is invoked when a duplicate message is dropped.
func (t LogTracer) DuplicateMessage(msg *pubsub.Message) {
	log.Printf("(%s) DuplicateMsg: %s\n", t.id(), msg.String())
}

// ThrottlePeer is invoked when a peer is throttled by the peer gater.
func (t LogTracer) ThrottlePeer(p peer.ID) {
	log.Printf("(%s) ThrottlePeer: %s\n", t.id(), p.Pretty())
}

// RecvRPC is invoked when an incoming RPC is received.
func (t LogTracer) RecvRPC(rpc *pubsub.RPC) {
	log.Printf("(%s) RecvRPC: %s\n", t.id(), rpc.String())
	for _, addr := range t.node.Addresses() {
		if !strings.Contains(addr, "127.0.0.1") {
			log.Printf("=> %s", addr)
		}
	}
	for _, msg := range rpc.GetPublish() {
		log.Printf("[msg] %s", msg.String())
	}
}

// SendRPC is invoked when a RPC is sent.
func (t LogTracer) SendRPC(rpc *pubsub.RPC, p peer.ID) {
	log.Printf("(%s) SendRPC to %s: %s\n", t.id(), p.Pretty(), rpc.String())
}

// DropRPC is invoked when an outbound RPC is dropped, typically because of a queue full.
func (t LogTracer) DropRPC(rpc *pubsub.RPC, p peer.ID) {
	log.Printf("(%s) DropRPC to %s: %s\n", t.id(), p.Pretty(), rpc.String())
}

// UndeliverableMessage is invoked when the consumer of Subscribe is not reading messages fast enough and
// the pressure release mechanism trigger, dropping messages.
func (t LogTracer) UndeliverableMessage(msg *pubsub.Message) {
	log.Printf("(%s) UndeliverableMsg: %s\n", t.id(), msg.String())
}
