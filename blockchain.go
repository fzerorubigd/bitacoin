package bitacoin

import (
	"errors"
	"fmt"
)

// BlockChain is the group of a block, with difficulty level
type BlockChain struct {
	Difficulty int
	Mask       []byte

	store Store
}

// Add a new data to the end of the block chain by creating a new block
func (bc *BlockChain) Add(data string) (*Block, error) {
	hash, err := bc.store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("Getting the last block failed: %w", err)
	}
	b := NewBlock(data, bc.Mask, hash)
	if err := bc.store.Append(b); err != nil {
		return nil, fmt.Errorf("Append new block to store failed: %w", err)
	}

	return b, nil
}

// Print the current blockchain in Stdout, it's
func (bc *BlockChain) Print() error {
	fmt.Printf("Difficulty : %d\nStore Backend: %T\n", bc.Difficulty, bc.store)

	return Iterate(bc.store, func(b *Block) error {
		fmt.Print(b)
		return nil
	})
}

// Validate all data in the block chain
func (bc *BlockChain) Validate() error {
	return Iterate(bc.store, func(b *Block) error {
		if err := b.Validate(bc.Mask); err != nil {
			return fmt.Errorf("block chain is not valid: %w", err)
		}

		return nil
	})
}

// NewBlockChain creates a new block chain with a difficulty, difficulty in this
// block chain is the number of zeros in the begining of the generated hash
func NewBlockChain(difficulty int, store Store) (*BlockChain, error) {
	mask := GenerateMask(difficulty)
	bc := BlockChain{
		Difficulty: difficulty,
		Mask:       mask,
		store:      store,
	}

	_, err := store.LastHash()
	if err == nil {
		return &bc, nil
	}

	if !errors.Is(err, ErrNotInitialized) {
		return nil, fmt.Errorf("getting the last hash failed: %w", err)
	}

	gb := NewBlock("Genesis Block", bc.Mask, []byte{})
	if err := store.Append(gb); err != nil {
		return nil, fmt.Errorf("Add Genesis block to store failed: %w", err)
	}

	return &bc, nil
}
