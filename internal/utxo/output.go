package utxo

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/mastrogiovanni/mychain/internal/account"
	"golang.org/x/crypto/blake2b"
)

type Output struct {
	ID        string
	Owner     account.GenericAccount
	Payload   []byte
	Signature []byte
}

func NewOutput(payload []byte, acc account.Account) (*Output, error) {
	output := &Output{
		Payload: payload,
	}
	err := output.buildOutput(acc)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (output *Output) buildID() string {
	var buffer bytes.Buffer
	buffer.Write(output.Payload)
	buffer.Write(output.Signature)
	buffer.Write(output.Owner.Type)
	buffer.Write(output.Owner.PublicKey)
	ID := blake2b.Sum256(buffer.Bytes())
	return hex.EncodeToString(ID[:])
}

func (output *Output) buildOutput(acc account.Account) error {
	if output.Payload == nil {
		return fmt.Errorf("payload cannot be empty")
	}
	output.Owner = account.GenericAccount{
		Type:      acc.Type(),
		PublicKey: acc.PublicKey(),
	}
	output.Signature = acc.Sign(output.Payload)
	output.ID = output.buildID()
	return nil
}

func (output *Output) Verify() error {
	acc, err := account.GetAccountFromPublicKey(string(output.Owner.Type), output.Owner.PublicKey)
	if err != nil {
		return err
	}
	verified := acc.Verify(output.Payload, output.Signature)
	if !verified {
		return fmt.Errorf("signature don't match")
	}
	espectedID := output.buildID()
	if espectedID != output.ID {
		return fmt.Errorf("bad ID: is %s espected %s", output.ID, espectedID)
	}
	return nil
}

func (output *Output) String() string {
	return fmt.Sprintf("Output [ID: %s, Payload: '%s' (%d)]",
		output.ID, string(output.Payload), len(output.Payload),
	)
}
