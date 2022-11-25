package main

import (
	"context"
	"fmt"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/gdamore/tcell/v2"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	noise "github.com/libp2p/go-libp2p-noise"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/mastrogiovanni/mychain/internal/ring"
	"github.com/rivo/tview"
)

var logger = log.Logger("rendezvous")

const (
	DiscoveryServiceTag = "tokenring"
)

func createAndRunNode(seed int64) (host.Host, error) {
	ctx := context.Background()

	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)
	if err != nil {
		panic(err)
	}

	var idht *dht.IpfsDHT

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		panic(err)
	}
	host, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),
		// Multiple listen addresses
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",      // regular tcp connections
			"/ip4/0.0.0.0/udp/0/quic", // a UDP endpoint for the QUIC transport
		),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(connmgr),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h)
			return idht, err
		}),
		// Let this host use relays and advertise itself on relays if
		// it finds it is behind NAT. Use libp2p.Relay(options...) to
		// enable active relays and more.
		libp2p.EnableAutoRelay(),
		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		//
		// This service is highly rate-limited and should not cause any
		// performance issues.
		libp2p.EnableNATService(),
	)
	if err != nil {
		return nil, err
	}
	return host, nil
}

type MyListener struct {
	table *tview.Table
	app   *tview.Application

	token int64
	peers map[string]struct{}
}

func (l *MyListener) draw() {
	l.table.Clear()
	l.table.SetCell(0, 0, tview.NewTableCell("Token").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	l.table.SetCell(0, 1, tview.NewTableCell(fmt.Sprintf("%d", l.token)).SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	row := 1
	for peer := range l.peers {
		l.table.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", row)).SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
		l.table.GetCell(row, 1).SetText(peer)
		row++
	}
	l.app.Draw()
}

func (l *MyListener) PeerAdded(peerID string) {
	l.peers[peerID] = struct{}{}
	l.draw()
}

func (l *MyListener) PeerRemoved(peerID string) {
	delete(l.peers, peerID)
	l.draw()
}

func (l *MyListener) TokenReceived(token int64) {
	l.token = token
	l.draw()
}

func main() {

	log.SetAllLoggers(log.LevelFatal)
	log.SetLogLevel("rendezvous", "info")
	// log.SetLogLevel("tokenRing", "info")
	// help := flag.Bool("h", false, "Display Help")
	config, err := ParseFlags()
	if err != nil {
		panic(err)
	}

	table := tview.NewTable().SetBorders(true)
	table.SetCell(0, 0, tview.NewTableCell("Token").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("-").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))

	app := tview.NewApplication()
	app.SetRoot(table, true)

	// for seed := int64(0); seed < 10; seed++ {
	go func(seed int64) {

		host, err := createAndRunNode(seed)
		if err != nil {
			panic(err)
		}

		listener := &MyListener{
			table: table,
			app:   app,
		}

		trs := ring.NewTokenRingServiceWithListener(host, listener)

		err = trs.Start(context.Background(), DiscoveryServiceTag)
		if err != nil {
			panic(err)
		}

		if seed == 0 {
			go func() {
				fmt.Println(http.ListenAndServe("127.0.0.1:6060", nil))
			}()
		}

	}(config.Seed)

	if err := app.Run(); err != nil {
		panic(err)
	}
	// }

	// mux := http.NewServeMux()
	// mux.HandleFunc("/custom_debug_path/profile", pprof.Profile)
	// logger.Fatal(http.ListenAndServe(":7777", mux))

	// select {}

}
