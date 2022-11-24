package ring

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs/go-log/v2"
	pool "github.com/libp2p/go-buffer-pool"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

var logger = log.Logger("tokenRing")

const ()

const (
	PingSize     = 32
	pingTimeout  = time.Second * 5
	pingInterval = time.Second * 1
	ID           = "/tokenring/v0.0.1"
	PingID       = "/ping/v0.0.1"
	ServiceName  = "libp2p.tokenRing"
)

type TokenRingListener interface {
	PeerAdded(peerID string)
	PeerRemoved(peerID string)
	TokenReceived(token int64)
}

type TokenRingService struct {
	Host host.Host

	peers    map[string]peer.AddrInfo
	tokens   chan string
	maxToken int

	lastTokenSeen  time.Time
	totalRoundtrip time.Duration
	countRoundtrip int64

	listener *TokenRingListener
}

func NewTokenRingServiceWithListener(h host.Host, listener TokenRingListener) *TokenRingService {
	peers := make(map[string]peer.AddrInfo)
	peers[h.ID().Pretty()] = peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	trs := &TokenRingService{
		Host:           h,
		peers:          peers,
		tokens:         make(chan string),
		maxToken:       -1,
		lastTokenSeen:  time.Time{},
		totalRoundtrip: time.Duration(0),
		countRoundtrip: 0,
		listener:       &listener,
	}
	h.SetStreamHandler(ID, trs.TokenRingHandler)
	h.SetStreamHandler(PingID, trs.PingHandler)
	return trs
}

func NewTokenRingService(h host.Host) *TokenRingService {
	return NewTokenRingServiceWithListener(h, nil)
}

func (n *TokenRingService) PingHandler(s network.Stream) {
	if err := s.Scope().SetService(ServiceName); err != nil {
		logger.Debugf("error attaching stream to ping service: %s", err)
		s.Reset()
		return
	}

	if err := s.Scope().ReserveMemory(PingSize, network.ReservationPriorityAlways); err != nil {
		logger.Debugf("error reserving memory for ping stream: %s", err)
		s.Reset()
		return
	}
	defer s.Scope().ReleaseMemory(PingSize)

	buf := pool.Get(PingSize)
	defer pool.Put(buf)

	errCh := make(chan error, 1)
	defer close(errCh)
	timer := time.NewTimer(pingTimeout)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			logger.Debug("ping timeout")
		case err, ok := <-errCh:
			if ok {
				logger.Debug(err)
			} else {
				logger.Error("ping loop failed without error")
			}
		}
		s.Reset()
	}()

	for {
		_, err := io.ReadFull(s, buf)
		if err != nil {
			errCh <- fmt.Errorf("Bad read Ping: %s", err)
			return
		}

		_, err = s.Write(buf)
		if err != nil {
			errCh <- fmt.Errorf("Bad write Pong: %s", err)
			return
		}
		timer.Reset(pingTimeout)
	}
}

func sendPing(stream network.Stream) error {
	logger.Info("Sending Ping")

	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		logger.Errorf("failed to get cryptographic random: %s", err)
		return err
	}

	ra := mrand.New(mrand.NewSource(int64(binary.BigEndian.Uint64(b))))

	if err := stream.Scope().ReserveMemory(2*PingSize, network.ReservationPriorityAlways); err != nil {
		logger.Debugf("error reserving memory for ping stream: %s", err)
		return err
	}
	defer stream.Scope().ReleaseMemory(2 * PingSize)

	buf := pool.Get(PingSize)
	defer pool.Put(buf)

	if _, err := io.ReadFull(ra, buf); err != nil {
		return err
	}

	if _, err := stream.Write(buf); err != nil {
		return err
	}

	rbuf := pool.Get(PingSize)
	defer pool.Put(rbuf)

	if _, err := io.ReadFull(stream, rbuf); err != nil {
		return err
	}

	if !bytes.Equal(buf, rbuf) {
		return errors.New("ping packet was incorrect")
	}

	logger.Info("Pong returned")

	return nil

}

func (n *TokenRingService) HandlePeerFound(pi peer.AddrInfo) {
	logger.Info("Peer found:", pi)
	// time.Sleep(1 * time.Second)
	ctx := context.Background()
	if err := n.Host.Connect(ctx, pi); err != nil {
		logger.Error(err)
		n.HandlePeerDisconnected(pi.ID.Pretty())
		return
	}
	n.peers[pi.ID.Pretty()] = pi
	if n.listener != nil {
		(*(n.listener)).PeerAdded(pi.ID.Pretty())
	}
	n.printNodes()

	stream, err := n.Host.NewStream(ctx, pi.ID, PingID)
	if err != nil {
		logger.Error(err)
		n.HandlePeerDisconnected(pi.ID.Pretty())
		return
	}
	defer stream.Close()

	// Continue to ping the peer until it disappear
	for {
		err = sendPing(stream)
		if err != nil {
			logger.Error(err)
			n.HandlePeerDisconnected(pi.ID.Pretty())
			return
		}
		time.Sleep(pingInterval)
	}

}

func (n *TokenRingService) startTokenRing(ctx context.Context) error {
	for {
		// If a token was seen less than 5 seconds ago
		if !n.lastTokenSeen.IsZero() && time.Now().Before(n.lastTokenSeen.Add(5*time.Millisecond)) {
			time.Sleep(time.Second * 5)
			continue
		}
		// logger.Info("Last time:", n.lastTokenSeen)
		localId := n.Host.ID().Pretty()
		smaller := localId
		nextBig := ""
		for id := range n.peers {
			if strings.Compare(id, smaller) < 0 {
				smaller = id
			}
			if strings.Compare(localId, id) < 0 {
				if nextBig == "" || strings.Compare(nextBig, id) > 0 {
					nextBig = id
				}
			}
		}
		// logger.Infof("[START] me: %s, smaller: %s, nextBig: %s", n.h.ID().Pretty(), smaller, nextBig)
		if localId == smaller && nextBig != "" {
			logger.Infof("%s STARTING sending token to %s", localId, nextBig)
			addr := n.peers[nextBig]
			err := deliverToken(n.Host, addr, n.maxToken+1)
			if err != nil {
				logger.Error(err)
			} else {
				n.lastTokenSeen = time.Now()
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func (n *TokenRingService) Start(ctx context.Context, discoveryTag string) error {

	go n.startTokenRing(ctx)
	time.Sleep(1 * time.Second)

	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(n.Host, discoveryTag, n)
	return s.Start()
}

func (n *TokenRingService) TokenRingHandler(stream network.Stream) {

	if err := stream.Scope().SetService(ServiceName); err != nil {
		logger.Debugf("error attaching stream to token ring service: %s", err)
		stream.Reset()
		return
	}

	// if err := stream.Scope().ReserveMemory(PingSize, network.ReservationPriorityAlways); err != nil {
	// 	log.Debugf("error reserving memory for ping stream: %s", err)
	// 	s.Reset()
	// 	return
	// }
	// defer s.Scope().ReleaseMemory(PingSize)

	now := time.Now()
	defer stream.Close()

	for {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		token, err := rw.ReadString('\n')
		if err != nil {
			logger.Error(err)
			return
		}
		// reader.Reset(s)
		token = strings.TrimSuffix(token, "\n")
		tokenInt, err := strconv.Atoi(token)
		if err != nil {
			logger.Error(err)
			return
		}
		_, err = rw.WriteString("ack\n")
		if err != nil {
			logger.Error(err)
			return
		}
		err = rw.Flush()
		if err != nil {
			logger.Error(err)
			return
		}
		if tokenInt <= n.maxToken {
			logger.Error(fmt.Errorf("token %d is lower or equal than current maxToken %d: discard", tokenInt, n.maxToken))
			return
		}

		duration := now.Sub(n.lastTokenSeen)
		n.totalRoundtrip = n.totalRoundtrip + duration
		n.countRoundtrip = n.countRoundtrip + 1
		avgDuration := time.Duration(float64(n.totalRoundtrip) / float64(n.countRoundtrip))

		n.maxToken = tokenInt
		n.lastTokenSeen = now
		sender := stream.Conn().RemotePeer().Pretty()
		localId := n.Host.ID().Pretty()

		// logger.Infof("\t{%s} Rx %d from %s", localId, tokenInt, sender)

		bigger := localId
		smaller := localId
		nextBig := ""
		nextSmaller := ""

		for id := range n.peers {
			if strings.Compare(bigger, id) < 0 {
				bigger = id
			}
			if strings.Compare(smaller, id) > 0 {
				smaller = id
			}
			cmp := strings.Compare(localId, id)
			if cmp < 0 {
				if nextBig == "" || strings.Compare(id, nextBig) < 0 {
					nextBig = id
				}
			} else if cmp > 0 {
				if nextSmaller == "" || strings.Compare(id, nextSmaller) > 0 {
					nextSmaller = id
				}
			}
		}

		// logger.Infof("me: %s", localId)
		// logger.Infof("smaller: %s", smaller)
		// logger.Infof("bigger: %s", bigger)
		// logger.Infof("nextSmaller: %s", nextSmaller)
		// logger.Infof("nextBigger: %s", nextBig)
		// logger.Infof("sender: %s", sender)

		// if n.h.ID().Pretty() != smaller && s.Conn().RemotePeer().Pretty() == nextSmaller {
		// 	// Discard this message
		// 	continue
		// }

		next := nextBig
		if next == "" && n.Host.ID().Pretty() == bigger {
			next = smaller
		}
		if next == localId {
			return
		}
		// logger.Infof("next: %s", next)

		// time.Sleep(50 * time.Millisecond)
		addr := n.peers[next]

		// logger.Info("Continue sending token to", addr)
		// time.Sleep(1 * time.Millisecond)
		// logger.Infof("Wrote %d", res)
		logger.Infof("{%s, avg: %s} %s => %d => %s", duration, avgDuration, sender, tokenInt+1, next)

		err = deliverToken(n.Host, addr, tokenInt)
		if err != nil {
			logger.Error(fmt.Errorf("%s, %v", err, addr))

			// Remove peer
			n.HandlePeerDisconnected(next)

		} else {
			if n.listener != nil {
				(*(n.listener)).TokenReceived(int64(tokenInt + 1))
			}
			return
		}
	}

}

func deliverToken(host host.Host, addr peer.AddrInfo, tokenInt int) error {
	ctx := context.Background()

	stream, err := host.NewStream(ctx, addr.ID, ID)
	if err != nil {
		return err
	}
	defer stream.Reset()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	_, err = rw.WriteString(fmt.Sprintf("%d\n", tokenInt+1))
	if err != nil {
		return err
	}

	err = rw.Flush()
	if err != nil {
		return err
	}

	dataStream := make(chan string, 1)
	go func() {
		str, err := rw.ReadString('\n')
		if err != nil {
			dataStream <- fmt.Sprintf("error: %s", err)
			return
		}
		if str != "ack\n" {
			dataStream <- "error: ack is different"
			return
		}
		dataStream <- "ack"
	}()

	select {
	case msg := <-dataStream:
		if strings.HasPrefix(msg, "error") {
			return fmt.Errorf(msg)
		} else {
			logger.Infof("Rx ACK")
		}
	case <-time.After(1 * time.Second):
		return fmt.Errorf("Timeout")
	}
	return nil
}

func (n *TokenRingService) printNodes() {
	logger.Infof("[ Update Neighbors ]")
	for id := range n.peers {
		comp := strings.Compare(n.Host.ID().Pretty(), id)
		ps, _ := n.Host.Peerstore().GetProtocols(peer.ID(id))
		logger.Infof("- (%d) %s, %s", comp, id, ps)
	}
}

func (n *TokenRingService) HandlePeerDisconnected(id string) {
	logger.Infof("Disconnected %s", id)
	_, ok := n.peers[id]
	if ok {
		delete(n.peers, id)
		if n.listener != nil {
			(*(n.listener)).PeerRemoved(id)
		}
		n.printNodes()
	}
}
