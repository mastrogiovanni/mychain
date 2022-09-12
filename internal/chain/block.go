package chain

import (
	"encoding/hex"
	"fmt"
)

type Block struct {
	Sequence     int64    // It is the max of the prev + 1
	PrevBlockIds [][]byte // address == signature of previous blocks
	Payload      []byte   // Payload of current block
	Signature    []byte   // Signature of prev blocks signature + current payload. Blocks are ordered by address = signature
}

func (b *Block) String() string {
	return fmt.Sprintf("Block[Seq: %d, Payload: %s, Signature: %s]",
		b.Sequence,
		hex.EncodeToString(b.Payload),
		hex.EncodeToString(b.Signature),
	)
}
