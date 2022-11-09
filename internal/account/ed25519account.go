package account

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

const Ed25519AccountType = "ed25519"

type Ed25519Account struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewEd25519AccountFromPublicKey(publicKey []byte) *Ed25519Account {
	return &Ed25519Account{
		privateKey: nil,
		publicKey:  ed25519.PublicKey(publicKey),
	}
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

func (account *Ed25519Account) Type() []byte {
	return []byte(Ed25519AccountType)
}

func (account *Ed25519Account) PublicKey() []byte {
	return account.publicKey
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
