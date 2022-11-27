package chain

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/mastrogiovanni/mychain/internal/account"
)

var (
	ErrBlockNotFound = errors.New("block not found")
)

type Chain struct {
	genesys  *Block
	blocks   map[string]*Block
	snapshot []*Block
}

func NewChain(acc account.Account, genesys []byte) *Chain {
	genesysBlock := newGenesysBlock(acc, genesys)
	chain := &Chain{
		blocks:  make(map[string]*Block),
		genesys: genesysBlock,
	}
	chain.appendBlockNoCheck(genesysBlock)
	return chain
}

func (chain *Chain) appendBlockNoCheck(block *Block) {
	_, ok := chain.blocks[string(block.Signature)]
	if !ok {
		chain.blocks[string(block.Signature)] = block
		chain.snapshot = append(chain.snapshot, block)
	} else {
		panic(fmt.Errorf("found two blocks with same Signature"))
	}
}

func newGenesysBlock(acc account.Account, genesys []byte) *Block {
	if genesys != nil {
		genesysBlock := &Block{}
		genesysBlock.Deserialize(genesys)
		return genesysBlock
	} else {
		genesisMessage := []byte("Hello World!")
		genesysBlock := &Block{
			PrevBlockIds: make([][]byte, 0),
			Sequence:     0,
			Payload:      genesisMessage,
			Signature:    acc.Sign(genesisMessage),
			Account: account.GenericAccount{
				Type:      acc.Type(),
				PublicKey: acc.PublicKey(),
			},
		}
		return genesysBlock
	}
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
	c.appendBlockNoCheck(block)
	return block, nil
}

func (c *Chain) GetOneOfLastBlocksAtRandom(length int) *Block {
	if length > c.Size() {
		return c.snapshot[len(c.snapshot)-rand.Intn(c.Size())-1]
	} else {
		return c.snapshot[len(c.snapshot)-rand.Intn(length)-1]
	}
}

func (c *Chain) Size() int {
	return len(c.blocks)
}

func (c *Chain) Snapshot() []*Block {
	return c.snapshot
}

func (c *Chain) Append(block *Block) error {
	ok, err := c.Verify(block)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("block not signed correctly")
	}
	// Discard genesys block
	if bytes.Equal(block.Signature, c.genesys.Signature) {
		return nil
	}
	c.appendBlockNoCheck(block)
	return nil
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
