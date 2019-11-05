package bitacoin

import "errors"

var (
	ErrNotInitialized = errors.New("bita coin store is empty")
)

type Store interface {
	Load(hash []byte) (*Block, error)

	Append(b *Block) error

	LastHash() ([]byte, error)
}

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
