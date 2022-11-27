package account

import (
	"encoding/hex"
	"fmt"
)

type Account interface {
	Type() []byte
	PublicKey() []byte
	Sign(payload []byte) []byte
	Verify(payload []byte, signature []byte) bool
}

type GenericAccount struct {
	Type      []byte
	PublicKey []byte
}

func (g GenericAccount) String() string {
	return fmt.Sprintf("%s:%s", g.Type, hex.EncodeToString(g.PublicKey))
}

func (g GenericAccount) Verify(payload []byte, signature []byte) (bool, error) {
	acc, err := GetAccountFromPublicKey(string(g.Type), g.PublicKey)
	if err != nil {
		return false, nil
	}
	return acc.Verify(payload, signature), nil
}
