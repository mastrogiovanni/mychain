package chain

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"

	"github.com/mastrogiovanni/mychain/internal/account"
)

type Block struct {
	Sequence     int64                  // It is the max of the prev + 1
	PrevBlockIds [][]byte               // address == signature of previous blocks
	Payload      []byte                 // Payload of current block
	Account      account.GenericAccount // Creator Public Key
	Signature    []byte                 // Signature of prev blocks signature + current payload. Blocks are ordered by address = signature
}

func (block *Block) Serialize() ([]byte, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := gob.NewEncoder(writer)
	err := encoder.Encode(block)
	if err != nil {
		return nil, err
	}
	writer.Flush()
	return b.Bytes(), nil
}

func (block *Block) Deserialize(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(block)
}

func (b *Block) String() string {
	return fmt.Sprintf("Block[Seq: %d, PublicKey: %s:%s, Payload: %s, Signature: %s]",
		b.Sequence,
		string(b.Account.Type),
		hex.EncodeToString(b.Account.PublicKey),
		hex.EncodeToString(b.Payload),
		hex.EncodeToString(b.Signature),
	)
}
