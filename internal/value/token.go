package value

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"

	"github.com/mastrogiovanni/mychain/internal/account"
	"github.com/mastrogiovanni/mychain/internal/utxo"
)

type ValueOutput struct {
	owner account.GenericAccount
	value int64
}

func NewValueOutput(owner account.GenericAccount, value int64) *ValueOutput {
	return &ValueOutput{
		owner: owner,
		value: value,
	}
}

func NewValueOutputFromBytes(data []byte) (*ValueOutput, error) {
	vo := &ValueOutput{}
	err := vo.Deserialize(data)
	if err != nil {
		return nil, err
	}
	return vo, nil
}

func (vo *ValueOutput) Serialize() ([]byte, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := gob.NewEncoder(writer)
	err := encoder.Encode(vo)
	if err != nil {
		return nil, err
	}
	writer.Flush()
	return b.Bytes(), nil
}

func (vo *ValueOutput) Deserialize(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(vo)
}

type ValueEngineHandler struct{}

func (engine ValueEngineHandler) Validate(transaction *utxo.Transaction, inputs []*utxo.Output, outputs []*utxo.Output) error {

	// transaction signer must be equal to output owner
	for _, input := range inputs {
		if !reflect.DeepEqual(input.Owner, transaction.Account) {
			return fmt.Errorf("transaction signer don't match with block owner")
		}
	}

	// Compute sum of inputs
	sum := int64(0)
	for _, input := range inputs {
		vo, err := NewValueOutputFromBytes(input.Payload)
		if err != nil {
			return fmt.Errorf("impossible to parse value output: %s", err)
		}
		if !reflect.DeepEqual(vo.owner, transaction.Account) {
			return fmt.Errorf("transaction signer don't match with value owner")
		}
		sum += vo.value
	}

	// Comput sum of outputs
	sumOutput := int64(0)
	for _, output := range outputs {
		vo, err := NewValueOutputFromBytes(output.Payload)
		if err != nil {
			return fmt.Errorf("impossible to parse value output: %s", err)
		}
		sumOutput += vo.value
	}

	if sumOutput != sum {
		return fmt.Errorf("sum of inputs differs from sum of outputs")
	}

	return nil
}

func (engine ValueEngineHandler) Request(transaction *utxo.Transaction, input *utxo.Input) (*utxo.Output, error) {
	return nil, fmt.Errorf("block not found")
}
