package bitacoin

import "errors"

var (
	// ErrNotInitialized should be returned when the store
	//needs the genesis block (no data is stored in the store)
	ErrNotInitialized = errors.New("the store is not initialied, there is no block")
)

// Store is an interface to handle the blockchain storage
type Store interface {
	// Load should return the block from the store based on the requested hash
	Load(hash []byte) (*Block, error)

	// Append should append the block to the store, it should check if the
	// last block hash match with the hash in the new block and also updates
	// the last hash
	Append(b *Block) error

	// LastHash returns the last hash in the store, if there is no block
	// (not even the genesis block) it should return the ErrNotInitialized
	LastHash() ([]byte, error)
}

// Iterate over the blocks in the store, if the callback returns an error it
// stops the loop and return the error to the caller
func Iterate(store Store, fn func(b *Block) error) error {
	last, err := store.LastHash()
	if err != nil {
		return err
	}

	for {
		b, err := store.Load(last)
		if err != nil {
			return err
		}

		if err := fn(b); err != nil {
			return err
		}

		if len(b.PrevHash) == 0 {
			return nil
		}

		last = b.PrevHash
	}
}
