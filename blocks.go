package bitacoin

import (
	"bytes"
	"fmt"
	"time"
)

// Block is the core data for the block chain. it can contain anything,
// a block is like a record in a table in database
type Block struct {
	Timestamp time.Time
	Data      []byte

	Nonce    int32
	PrevHash []byte
	Hash     []byte
}

func (b *Block) String() string {
	return fmt.Sprintf(
		"Time: %s\nData: %s\nHash: %x\nPrevHash: %x\nNonce: %d\n----\n",
		b.Timestamp, b.Data, b.Hash, b.PrevHash, b.Nonce,
	)
}

// Validate try to validate the current block, it needs a difficulty mask
// for validating the hash difficulty
func (b *Block) Validate(mask []byte) error {
	h := EasyHash(b.Timestamp.UnixNano(), b.Data, b.PrevHash, b.Nonce)
	if !bytes.Equal(h, b.Hash) {
		return fmt.Errorf("the hash is invalid it should be %x is %x", h, b.Hash)
	}

	if !GoodEnough(mask, h) {
		return fmt.Errorf("hash is not good enough with mast %x", mask)
	}

	return nil
}

// NewBlock creates a new block in the system, it needs deficulty mask for
// create a good hash, and also the previous block hash
func NewBlock(data string, mask, prevHash []byte) *Block {
	b := Block{
		Timestamp: time.Now(),
		Data:      []byte(data),
		PrevHash:  prevHash,
	}
	b.Hash, b.Nonce = DifficultHash(mask, b.Timestamp.UnixNano(), b.Data, b.PrevHash)

	return &b
}

// BlockChain is the group of a block, with difficulty level
type BlockChain struct {
	Difficulty int
	Mask       []byte
	Blocks     []*Block
}

// Add a new data to the end of the block chain by creating a new block
func (bc *BlockChain) Add(data string) {
	ln := len(bc.Blocks)

	if ln == 0 {
		panic("why?")
	}

	bc.Blocks = append(
		bc.Blocks,
		NewBlock(data, bc.Mask, bc.Blocks[ln-1].Hash),
	)
}

func (bc *BlockChain) String() string {
	var ret = fmt.Sprintf("Difficulty : %d\n\n", bc.Difficulty)
	for i := range bc.Blocks {
		ret += bc.Blocks[i].String()
	}

	return ret
}

// Validate all data in the block chain
func (bc *BlockChain) Validate() error {
	for i := range bc.Blocks {
		if err := bc.Blocks[i].Validate(bc.Mask); err != nil {
			return fmt.Errorf("block chain is not valid: %w", err)
		}

		if i == 0 {
			continue
		}

		if !bytes.Equal(bc.Blocks[i].PrevHash, bc.Blocks[i-1].Hash) {
			return fmt.Errorf("the order is invalid, it is %x should be %x", bc.Blocks[i].PrevHash, bc.Blocks[i-1].Hash)
		}
	}

	return nil
}

// NewBlockChain creates a new block chain with a difficulty, difficulty in this
// block chain is the number of zeros in the begining of the generated hash
func NewBlockChain(difficulty int) *BlockChain {
	mask := GenerateMask(difficulty)
	bc := BlockChain{
		Difficulty: difficulty,
		Mask:       mask,
	}
	bc.Blocks = []*Block{NewBlock("Genesis Block", bc.Mask, []byte{})}

	return &bc
}
