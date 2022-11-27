package utxo

import (
	"fmt"
	"testing"

	"github.com/mastrogiovanni/mychain/internal/account"
	"github.com/stretchr/testify/require"
)

type EngineHandlerStub struct{}

func (engine EngineHandlerStub) Validate(transaction *Transaction, inputs []*Output, outputs []*Output) error {
	return nil
}

func (engine EngineHandlerStub) Request(transaction *Transaction, input *Input) (*Output, error) {
	return nil, fmt.Errorf("block not found")
}

func TestConsume(t *testing.T) {
	handler := &EngineHandlerStub{}
	engine := NewEngine(handler)
	acc := account.NewEd25519Account()
	o1, err := NewOutput([]byte("Hello World 1"), acc)
	require.NoError(t, err)
	err = engine.AddOutput(o1)
	require.NoError(t, err)
	o2, err := NewOutput([]byte("Hello World 2"), acc)
	require.NoError(t, err)
	err = engine.Submit(&Transaction{
		Inputs:  NewInputsFromOutputs([]*Output{o1}),
		Outputs: []*Output{o2},
	})
	require.NoError(t, err)
}

func TestMissingOutput(t *testing.T) {
	handler := &EngineHandlerStub{}
	engine := NewEngine(handler)
	acc := account.NewEd25519Account()
	o1, err := NewOutput([]byte("Hello World 1"), acc)
	require.NoError(t, err)
	err = engine.AddOutput(o1)
	require.NoError(t, err)
	o2, err := NewOutput([]byte("Hello World 2"), acc)
	require.NoError(t, err)
	err = engine.AddOutput(o2)
	require.NoError(t, err)
	o3, err := NewOutput([]byte("Hello World 3"), acc)
	require.NoError(t, err)
	err = engine.Submit(&Transaction{
		Inputs:  NewInputsFromOutputs([]*Output{o1}),
		Outputs: []*Output{o2},
	})
	require.NoError(t, err)
	err = engine.Submit(&Transaction{
		Inputs:  NewInputsFromOutputs([]*Output{o1}),
		Outputs: []*Output{o3},
	})
	require.Error(t, err)
}
