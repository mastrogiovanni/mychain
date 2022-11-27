package modules

import (
	"log"

	"github.com/mastrogiovanni/mychain/internal/stack"
	"github.com/mastrogiovanni/mychain/internal/utxo"
)

type UtxoLayer struct {
	*stack.Layer
}

type UtxoLayerConfig struct {
}

func NewUtxoLayer(config *UtxoLayerConfig) (*UtxoLayer, error) {
	ul := &UtxoLayer{}
	ul.Layer = stack.NewLayer(ul.HandleMsg)
	return ul, nil
}

func (p *UtxoLayer) HandleMsg(direction stack.Direction, data []byte, layer *stack.Layer) {

	transaction, err := utxo.NewTransactionFromBytes(data)
	if err != nil {
		log.Println("discarded malformed transaction:", err)
	}
	checked, err := transaction.Verify()
	if err != nil {
		log.Println("discarded unverified transaction:", err)
	}
	if !checked {
		log.Println("discarded unverified transaction:", err)
	}

	if direction == stack.Up {
		go p.Layer.SendUp(data)
	} else if direction == stack.Down {
		go p.Layer.SendDown(data)
	}

}
