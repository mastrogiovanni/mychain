package account

import "fmt"

func GetAccount(accountType string) (Account, error) {
	if accountType == Ed25519AccountType {
		return NewEd25519Account(), nil
	}
	return nil, fmt.Errorf("wrong account type passed")
}

func GetAccountFromPublicKey(accountType string, publicKey []byte) (Account, error) {
	if accountType == Ed25519AccountType {
		return NewEd25519AccountFromPublicKey(publicKey), nil
	}
	return nil, fmt.Errorf("wrong account type passed")
}
