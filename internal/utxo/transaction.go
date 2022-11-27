package utxo

import (
	"bufio"
	"bytes"
	"encoding/gob"

	"github.com/mastrogiovanni/mychain/internal/account"
)

type Transaction struct {
	Account   account.GenericAccount
	Inputs    []*Input
	Outputs   []*Output
	Signature []byte
}

func NewTransactionFromBytes(data []byte) (*Transaction, error) {
	transaction := &Transaction{}
	err := transaction.Deserialize(data)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (transaction *Transaction) Serialize() ([]byte, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := gob.NewEncoder(writer)
	err := encoder.Encode(transaction)
	if err != nil {
		return nil, err
	}
	writer.Flush()
	return b.Bytes(), nil
}

func (transaction *Transaction) Deserialize(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(transaction)
}

func (t *Transaction) Sign(acc account.Account) error {
	t.Account = account.GenericAccount{
		Type:      acc.Type(),
		PublicKey: acc.PublicKey(),
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := gob.NewEncoder(writer)
	err := encoder.Encode(t.Inputs)
	if err != nil {
		return err
	}
	err = encoder.Encode(t.Outputs)
	if err != nil {
		return err
	}
	writer.Flush()
	t.Signature = acc.Sign(b.Bytes())
	return nil
}

func (t *Transaction) Verify() (bool, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := gob.NewEncoder(writer)
	err := encoder.Encode(t.Inputs)
	if err != nil {
		return false, err
	}
	err = encoder.Encode(t.Outputs)
	if err != nil {
		return false, err
	}
	writer.Flush()
	return t.Account.Verify(b.Bytes(), t.Signature)
}
