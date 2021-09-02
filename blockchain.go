package bitacoin

import (
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	Difficulty = 20
)

// BlockChain is the group of a block, with difficulty level
type BlockChain struct {
	Difficulty int
	Mask       []byte

	store Store
}

// Add a new data to the end of the block chain by creating a new block
func (bc *BlockChain) Add(data ...*Transaction) (*Block, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to add")
	}
	hash, err := bc.store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("Getting the last block failed: %w", err)
	}
	b := NewBlock(data, bc.Difficulty, hash)
	if err := bc.store.Append(b); err != nil {
		return nil, fmt.Errorf("Append new block to store failed: %w", err)
	}

	return b, nil
}

// Print the current blockchain in Stdout, it's
func (bc *BlockChain) Print(header bool, count int) error {
	fmt.Printf("Difficulty : %d\nStore Backend: %T\n", bc.Difficulty, bc.store)
	if header {
		return nil
	}
	var errEnough = fmt.Errorf("enough")
	err := Iterate(bc.store, func(b *Block) error {
		if count > 0 {
			count--
		}
		fmt.Print(b)

		if count == 0 {
			return errEnough
		}
		return nil
	})

	if errors.Is(err, errEnough) {
		return nil
	}

	return err
}

// AllBlocks get list of all block
func (bc *BlockChain) AllBlocks() []*Block {
	var blocks []*Block
	err := Iterate(bc.store, func(b *Block) error {
		blocks = append(blocks, b)
		return nil
	})

	if err != nil {

	}
	return blocks
}

// Validate all data in the block chain
func (bc *BlockChain) Validate() error {
	return Iterate(bc.store, func(b *Block) error {
		if err := b.Validate(bc.Difficulty); err != nil {
			return fmt.Errorf("block chain is not valid: %w", err)
		}

		return nil
	})
}

func (bc *BlockChain) UnspentTxn(address []byte) (map[string]*Transaction, map[string][]int, int, error) {
	spent := make(map[string][]int)
	txom := make(map[string][]int)
	txns := make(map[string]*Transaction)
	acc := 0
	err := Iterate(bc.store, func(b *Block) error {
		for _, txn := range b.Transactions {
			txnID := hex.EncodeToString(txn.ID)

			for i := range txn.VOut {
				if txn.VOut[i].TryUnlock(address) && !inArray(i, spent[txnID]) {
					txns[txnID] = txn
					txom[txnID] = append(txom[txnID], i)
					acc += txn.VOut[i].Value
				}
			}

			delete(spent, txnID)

			if txn.IsCoinBase() {
				continue
			}

			for i := range txn.VIn {
				if txn.VIn[i].MatchLock(address) {
					outID := hex.EncodeToString(txn.VIn[i].TXID)
					spent[outID] = append(spent[outID], txn.VIn[i].VOut)
				}
			}

		}

		return nil
	})
	if err != nil {
		return nil, nil, 0, fmt.Errorf("iterate error: %w", err)
	}

	return txns, txom, acc, nil
}

// NewBlockChain creates a new block chain with a difficulty, difficulty in this
// block chain is the number of zeros in the begining of the generated hash
func NewBlockChain(genesis []byte, difficulty int, store Store) (*BlockChain, error) {
	mask := GenerateMask(difficulty)
	bc := BlockChain{
		Difficulty: difficulty,
		Mask:       mask,
		store:      store,
	}

	_, err := store.LastHash()
	if !errors.Is(err, ErrNotInitialized) {
		return nil, fmt.Errorf("store already initialized")
	}
	gbTxn := NewCoinBaseTxn(genesis, nil)
	gb := NewBlock([]*Transaction{gbTxn}, bc.Difficulty, []byte{})
	if err := store.Append(gb); err != nil {
		return nil, fmt.Errorf("Add Genesis block to store failed: %w", err)
	}

	return &bc, nil
}

// OpenBlockChain tries to open a blockchain, it returns an error if the blockchain store
// is empty
func OpenBlockChain(difficulty int, store Store) (*BlockChain, error) {
	mask := GenerateMask(difficulty)
	bc := BlockChain{
		Difficulty: difficulty,
		Mask:       mask,
		store:      store,
	}

	_, err := store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("openning the store failed: %w", err)
	}

	return &bc, nil
}
