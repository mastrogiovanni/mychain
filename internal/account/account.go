package account

type Account interface {
	Sign(payload []byte) []byte
	Verify(payload []byte, signature []byte) bool
}
