package modules

import (
	"log"

	"github.com/mastrogiovanni/mychain/internal/stack"
	"github.com/mastrogiovanni/mychain/internal/utxo"
)

type ValueLayer struct {
	*stack.Layer
	*utxo.Engine
}

type ValueLayerConfig struct{}

func NewValueLayer(config *ValueLayerConfig) (*ValueLayer, error) {
	vl := &ValueLayer{}
	vl.Layer = stack.NewLayer(vl.HandleMsg)
	vl.Engine = utxo.NewEngine(vl)
	return vl, nil
}

func (p *ValueLayer) HandleMsg(direction stack.Direction, data []byte, layer *stack.Layer) {

	transaction, err := utxo.NewTransactionFromBytes(data)
	if err != nil {
		log.Println("discarded malformed transaction:", err)
	}

	err = p.Engine.Submit(transaction)
	if err != nil {
		log.Println("discarded uncorrect transaction:", err)
	}

}

// Validate if output to consume and output to insert are sound with respect to the system
func (p *ValueLayer) Validate(transaction *utxo.Transaction, inputs []*utxo.Output, outputs []*utxo.Output) error {
	return nil
}

// Request missing inputs. Transaction is provided to enqueue it again for evaluation once Input is retrieved
func (p *ValueLayer) Request(transaction *utxo.Transaction, input *utxo.Input) (*utxo.Output, error) {
	return nil, nil
}

// Transaction is submitted
func (p *ValueLayer) Submitted(inputs []*utxo.Output, outputs []*utxo.Output) {
}
