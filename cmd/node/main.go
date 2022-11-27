package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mastrogiovanni/mychain/internal/modules"
	"github.com/mastrogiovanni/mychain/internal/stack"
)

type Config struct {
	modules.PhysicalLayerConfig
	modules.UtxoLayerConfig
	modules.ValueLayerConfig

	spammer bool
}

func parseFlags() *Config {
	config := &Config{}
	flag.StringVar(&config.PhysicalLayerConfig.Peers, "peers", "", "List of peers separated by column")
	flag.StringVar(&config.PhysicalLayerConfig.Genesys, "genesys", "", "Genesys block")
	flag.BoolVar(&config.spammer, "spammer", false, "True if you want to spam each second")
	flag.Parse()
	return config
}

func main() {
	config := parseFlags()

	pl, err := modules.NewPhysicalLayer(&config.PhysicalLayerConfig)
	if err != nil {
		panic(err)
	}

	ul, err := modules.NewUtxoLayer(&config.UtxoLayerConfig)
	if err != nil {
		panic(err)
	}
	ul.Layer.AttachLower(pl.Layer)

	vl, err := modules.NewValueLayer(&config.ValueLayerConfig)
	if err != nil {
		panic(err)
	}
	vl.Layer.AttachLower(ul.Layer)

	genesys, err := pl.GetGenesysRaw()
	if err != nil {
		panic(err)
	}

	log.Println("If want to join:")
	for _, addr := range pl.Node.Addresses() {
		log.Printf("go run cmd/node/main.go -peers %s -genesys %s -spammer true", addr, hex.EncodeToString(genesys))
	}

	for _, addr := range pl.Node.Addresses() {
		log.Printf("I'm %s", addr)
	}

	if config.spammer {
		go func() {
			log.Println("Starting Spammer")
			count := 0
			for {
				time.Sleep(1 * time.Second)
				pl.Recv(stack.Down, []byte(fmt.Sprintf("msg (%d) from: %s", count, pl.Node.ID())))
				count++
			}
		}()
	}

	select {}
}
