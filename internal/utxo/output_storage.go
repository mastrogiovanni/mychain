package utxo

import (
	"bytes"
	"fmt"
)

type OutputStorage struct {
	Outputs map[string]*Output
}

func NewOutputStorage() *OutputStorage {
	return &OutputStorage{
		Outputs: make(map[string]*Output),
	}
}

func (storage *OutputStorage) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("OutputStorage:\n")
	for _, output := range storage.Outputs {
		buffer.WriteString(fmt.Sprintf("* %s\n", output))
	}
	return buffer.String()
}

func (storage *OutputStorage) Add(output *Output) error {
	err := output.Verify()
	if err != nil {
		return err
	}
	storage.Outputs[output.ID] = output
	return nil
}

func (storage *OutputStorage) findOutputFromTransaction(transaction *Transaction) ([]*Output, error) {
	outputs := make([]*Output, 0)
	for _, input := range transaction.Inputs {
		output, ok := storage.Outputs[input.ID]
		if !ok {
			return nil, fmt.Errorf("output not found %s", input.ID)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (storage *OutputStorage) Submit(transaction *Transaction) error {

	// Retrieve output from database
	outputs, err := storage.findOutputFromTransaction(transaction)
	if err != nil {
		return err
	}

	// Verify each output
	for _, output := range transaction.Outputs {
		err := output.Verify()
		if err != nil {
			return err
		}
	}

	// Consume outputs
	for _, output := range outputs {
		delete(storage.Outputs, output.ID)
	}

	// Insert other Outputs
	for _, output := range transaction.Outputs {
		storage.Outputs[output.ID] = output
	}

	return nil
}
