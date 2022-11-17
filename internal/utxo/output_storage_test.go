package utxo

import (
	"testing"

	"github.com/mastrogiovanni/mychain/internal/account"
	"github.com/stretchr/testify/require"
)

func TestConsume(t *testing.T) {
	storage := NewOutputStorage()
	acc := account.NewEd25519Account()
	o1, err := NewOutput([]byte("Hello World 1"), acc)
	require.NoError(t, err)
	err = storage.Add(o1)
	require.NoError(t, err)
	o2, err := NewOutput([]byte("Hello World 2"), acc)
	require.NoError(t, err)
	err = storage.Submit(&Transaction{
		Inputs:  NewInputsFromOutputs([]*Output{o1}),
		Outputs: []*Output{o2},
	})
	require.NoError(t, err)
}

func TestMissingOutput(t *testing.T) {
	storage := NewOutputStorage()
	acc := account.NewEd25519Account()
	o1, err := NewOutput([]byte("Hello World 1"), acc)
	require.NoError(t, err)
	err = storage.Add(o1)
	require.NoError(t, err)
	o2, err := NewOutput([]byte("Hello World 2"), acc)
	require.NoError(t, err)
	err = storage.Add(o2)
	require.NoError(t, err)
	o3, err := NewOutput([]byte("Hello World 3"), acc)
	require.NoError(t, err)
	err = storage.Submit(&Transaction{
		Inputs:  NewInputsFromOutputs([]*Output{o1}),
		Outputs: []*Output{o2},
	})
	require.NoError(t, err)
	err = storage.Submit(&Transaction{
		Inputs:  NewInputsFromOutputs([]*Output{o1}),
		Outputs: []*Output{o3},
	})
	require.Error(t, err)
}
