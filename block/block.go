package block

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fzerorubigd/bitacoin/hasher"
	"github.com/fzerorubigd/bitacoin/transaction"
	"time"
)

// Block is the core data for the blockchain. it can contain anything,
// a block is like a record in a table in database
type Block struct {
	Time         int64
	Transactions []*transaction.Transaction
	Nonce        uint64
	PrevHash     []byte
	Hash         []byte
}

func (b *Block) String() string {
	return fmt.Sprintf(
		"Time: %d\nTxn Count: %d\nHash: %x\nPrevHash: %x\nNonce: %d\n----\n",
		b.Time, len(b.Transactions), b.Hash, b.PrevHash, b.Nonce,
	)
}

// Validate try to validate the current block, it needs a difficulty mask
// for validating the hash difficulty
func (b *Block) Validate(mask []byte) error {
	h := hasher.EasyHash(b.Time, transaction.CalculateTxnsHash(b.Transactions...), b.PrevHash, b.Nonce)
	if !bytes.Equal(h, b.Hash) {
		return fmt.Errorf("the hash is invalid it should be %x is %x", h, b.Hash)
	}

	if !hasher.GoodEnough(mask, h) {
		return fmt.Errorf("hash is not good enough with mask %x", mask)
	}

	return nil
}

func (b *Block) Contains(txnID []byte) bool {
	for i := range b.Transactions {
		if bytes.Equal(txnID, b.Transactions[i].ID) {
			return true
		}
	}

	return false
}

// Mine creates a new block in the system, it needs difficulty mask for
// create a good hash, and also the previous block hash
func Mine(ctx context.Context, txns []*transaction.Transaction, mask, prevHash []byte) *Block {
	b := Block{
		Time:         time.Now().UnixNano(),
		Transactions: txns,
		PrevHash:     prevHash,
	}
	b.Hash, b.Nonce = hasher.DifficultHash(
		ctx,
		mask,
		b.Time,
		transaction.CalculateTxnsHash(b.Transactions...),
		b.PrevHash,
	)

	if len(b.Hash) == 0 {
		return nil
	}

	return &b
}
