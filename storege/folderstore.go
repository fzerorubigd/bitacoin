package storege

import (
	"bytes"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/helper"
	"log"
	"path/filepath"
)

type folderConfig struct {
	LastHash []byte
}

type folderStore struct {
	root string

	config *folderConfig

	configPath string
}

func (fs *folderStore) Load(hash []byte) (*block.Block, error) {
	path := filepath.Join(fs.root, fmt.Sprintf("%x.json", hash))
	var b block.Block
	if err := helper.ReadJSON(path, &b); err != nil {
		return nil, fmt.Errorf("read JOSN file failed: %w", err)
	}

	return &b, nil
}

func (fs *folderStore) Append(b *block.Block) error {
	if !bytes.Equal(fs.config.LastHash, b.PrevHash) {
		return fmt.Errorf("store is out of sync")
	}

	path := filepath.Join(fs.root, fmt.Sprintf("%x.json", b.Hash))
	if err := helper.WriteJSON(path, b); err != nil {
		return fmt.Errorf("write JSON file failed: %w", err)
	}

	fs.config.LastHash = b.Hash
	if err := helper.WriteJSON(fs.configPath, fs.config); err != nil {
		return fmt.Errorf("write configuration file failed: %w", err)
	}

	return nil
}

func (fs *folderStore) LastHash() ([]byte, error) {
	if len(fs.config.LastHash) == 0 {
		return nil, ErrNotInitialized
	}

	return fs.config.LastHash, nil
}

// NewFolderStore create a file based storage for storing the blocks in the
// files, each block is in one file, and also there is a config file, for
// keep track of the last hash in the block
func NewFolderStore(storePath string) Store {
	fs := &folderStore{
		root:       storePath,
		config:     &folderConfig{},
		configPath: filepath.Join(storePath, "lastHash.json"),
	}

	if err := helper.ReadJSON(fs.configPath, fs.config); err != nil {
		log.Print("Empty store")
		fs.config.LastHash = nil
	}

	return fs
}
