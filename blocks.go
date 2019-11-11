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
		return fmt.Errorf("hash is not good enough with mask %x", mask)
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
