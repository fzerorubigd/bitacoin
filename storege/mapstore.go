package storege

import (
	"bytes"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
)

type mapStore struct {
	data map[string]*block.Block
	last []byte
}

func (ms *mapStore) Load(hash []byte) (*block.Block, error) {
	x := fmt.Sprintf("%x", hash)
	if b, ok := ms.data[x]; ok {
		return b, nil
	}

	return nil, fmt.Errorf("block is not in this store")
}

func (ms *mapStore) Append(b *block.Block) error {
	if !bytes.Equal(ms.last, b.PrevHash) {
		return fmt.Errorf("store is out of sync")
	}

	x := fmt.Sprintf("%x", b.Hash)
	if _, ok := ms.data[x]; ok {
		return fmt.Errorf("duplicate block")
	}

	ms.data[x] = b
	ms.last = b.Hash
	return nil
}

func (ms *mapStore) LastHash() ([]byte, error) {
	if len(ms.last) == 0 {
		return nil, ErrNotInitialized
	}

	return ms.last, nil
}

// NewMapStore is used to create an in memory and not persistent storage, useful
// for tests
func NewMapStore() Store {
	return &mapStore{
		data: make(map[string]*block.Block),
	}
}
