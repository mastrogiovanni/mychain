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

func (storage *OutputStorage) AddOutput(output *Output) error {
	err := output.Verify()
	if err != nil {
		return err
	}
	storage.Outputs[output.ID] = output
	return nil
}

func (storage *OutputStorage) FindOutputFromInput(input *Input) (*Output, error) {
	output, ok := storage.Outputs[input.ID]
	if !ok {
		return nil, fmt.Errorf("output not found %s", input.ID)
	}
	return output, nil
}
