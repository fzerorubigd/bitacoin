package bitacoin

import (
	"bytes"
	"fmt"
	"time"
)

// Block is the core data for the block chain. it can contain anything,
// a block is like a record in a table in database
type Block struct {
	Timestamp    time.Time
	Transactions []*Transaction

	Nonce    int32
	PrevHash []byte
	Hash     []byte
}

func (b *Block) String() string {
	return fmt.Sprintf(
		"Time: %s\nTxn Count: %d\nHash: %x\nPrevHash: %x\nNonce: %d\n----\n",
		b.Timestamp, len(b.Transactions), b.Hash, b.PrevHash, b.Nonce,
	)
}

// Validate try to validate the current block, it needs a difficulty mask
// for validating the hash difficulty
func (b *Block) Validate(mask []byte) error {
	h := EasyHash(b.Timestamp.UnixNano(), calculateTxnsHash(b.Transactions...), b.PrevHash, b.Nonce)
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
func NewBlock(txns []*Transaction, mask, prevHash []byte) *Block {
	b := Block{
		Timestamp:    time.Now(),
		Transactions: txns,
		PrevHash:     prevHash,
	}
	b.Hash, b.Nonce = DifficultHash(mask, b.Timestamp.UnixNano(), calculateTxnsHash(b.Transactions...), b.PrevHash)

	return &b
}
