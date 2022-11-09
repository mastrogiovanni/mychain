package account

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
