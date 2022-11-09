package chain

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/mastrogiovanni/mychain/internal/account"
)

var (
	ErrBlockNotFound = errors.New("block not found")
)

type Chain struct {
	genesys *Block
	blocks  map[string]*Block
}

func NewChain(acc account.Account) *Chain {

	genesisMessage := []byte("Hello World!")

	chain := &Chain{
		blocks: make(map[string]*Block),
		genesys: &Block{
			PrevBlockIds: make([][]byte, 0),
			Sequence:     0,
			Payload:      genesisMessage,
			Signature:    acc.Sign(genesisMessage),
			Account: account.GenericAccount{
				Type:      acc.Type(),
				PublicKey: acc.PublicKey(),
			},
		},
	}

	chain.blocks[string(chain.genesys.Signature)] = chain.genesys
	return chain
}

func CreateGenesys() {

}

func (c *Chain) Blocks() []*Block {
	v := make([]*Block, 0, len(c.blocks))
	for _, value := range c.blocks {
		v = append(v, value)
	}
	return v
}

func (c *Chain) Genesis() *Block {
	return c.genesys
}

// Max returns the larger of x or y.
func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func (c *Chain) NewBlockFromBlocks(blocks []*Block, payload []byte, acc account.Account) (*Block, error) {
	blockIds := make([][]byte, 0, len(blocks))
	for _, block := range blocks {
		blockIds = append(blockIds, block.Signature)
	}
	return c.NewBlock(blockIds, payload, acc)
}

func (c *Chain) NewBlock(prevBlockIds [][]byte, payload []byte, acc account.Account) (*Block, error) {
	sort.Sort(BytesSlice(prevBlockIds))
	var sequence int64 = 0
	var buffer bytes.Buffer
	for _, prevBlockId := range prevBlockIds {
		block, ok := c.blocks[string(prevBlockId)]
		if !ok {
			return nil, fmt.Errorf("block %s not found", string(prevBlockId))
		}
		buffer.Write(block.Signature)
		sequence = Max(sequence, block.Sequence)
	}
	buffer.Write(payload)
	signature := acc.Sign(buffer.Bytes())
	block := &Block{
		PrevBlockIds: prevBlockIds,
		Sequence:     sequence + 1,
		Payload:      payload,
		Signature:    signature,
		Account: account.GenericAccount{
			Type:      acc.Type(),
			PublicKey: acc.PublicKey(),
		},
	}
	c.blocks[string(signature)] = block
	return block, nil
}

func (c *Chain) Verify(block *Block) (bool, error) {
	var buffer bytes.Buffer
	for _, blockId := range block.PrevBlockIds {
		block, ok := c.blocks[string(blockId)]
		if !ok {
			return false, fmt.Errorf("block %s not found", string(blockId))
		}
		buffer.Write(block.Signature)
	}
	buffer.Write(block.Payload)
	acc, err := account.GetAccountFromPublicKey(string(block.Account.Type), block.Account.PublicKey)
	if err != nil {
		return false, err
	}
	return acc.Verify(buffer.Bytes(), block.Signature), nil
}
