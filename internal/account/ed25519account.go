package account

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

type Ed25519Account struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewEd25519Account() *Ed25519Account {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return &Ed25519Account{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (account *Ed25519Account) Sign(payload []byte) []byte {
	return ed25519.Sign(account.privateKey, payload)
}

func (account *Ed25519Account) Verify(payload []byte, signature []byte) bool {
	return ed25519.Verify(account.publicKey, payload, signature)
}

func (account *Ed25519Account) String() string {
	return fmt.Sprintf("Account (ed25519) [PublicKey: %s]",
		hex.EncodeToString(account.publicKey),
	)
}
