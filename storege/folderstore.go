package storege

import (
	"bytes"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/repository"
	"log"
	"path/filepath"
)

type folderStore struct {
	dataPath         string
	lastHash         []byte
	lastFourthBlocks []*block.Block
}

func (fs *folderStore) Load(hash []byte) (*block.Block, error) {
	path := filepath.Join(fs.dataPath, fmt.Sprintf("%x.json", hash))
	var b block.Block
	if err := helper.ReadJSON(path, &b); err != nil {
		return nil, fmt.Errorf("read JOSN file failed: %w", err)
	}

	return &b, nil
}

func (fs *folderStore) AppendBlock(b *block.Block) error {
	if !bytes.Equal(fs.lastHash, b.PrevHash) {
		return fmt.Errorf("store is out of sync")
	}

	path := filepath.Join(fs.dataPath, fmt.Sprintf("%x.json", b.Hash))
	if err := helper.WriteJSON(path, b); err != nil {
		return fmt.Errorf("write JSON file failed: %w", err)
	}

	fs.lastHash = b.Hash
	if err := helper.WriteJSON(filepath.Join(fs.dataPath, repository.LastHashFileName),
		map[string][]byte{"lastHash": fs.lastHash}); err != nil {
		return fmt.Errorf("write configuration file failed: %w", err)
	}

	lastFourthBlocks := make([]*block.Block, len(fs.lastFourthBlocks))
	lastFourthBlocks[0] = b
	if len(fs.lastFourthBlocks) < 4 {
		lastFourthBlocks = append(lastFourthBlocks, fs.lastFourthBlocks...)
		fs.lastFourthBlocks = lastFourthBlocks
	} else {
		lastFourthBlocks = append(lastFourthBlocks, fs.lastFourthBlocks[:3]...)
		fs.lastFourthBlocks = lastFourthBlocks
	}

	return nil
}

func (fs *folderStore) LastHash() ([]byte, error) {
	if len(fs.lastHash) == 0 {
		return nil, ErrNotInitialized
	}

	return fs.lastHash, nil
}

func (fs *folderStore) DataPath() string {
	return fs.dataPath
}

func (fs *folderStore) WriteJSON(fileName string, data interface{}) error {
	path := filepath.Join(fs.dataPath, fileName)
	err := helper.WriteJSON(path, data)
	if err != nil {
		return err
	}

	return nil
}

func (fs *folderStore) LastFourthBlocks() []*block.Block {
	return fs.lastFourthBlocks
}

// NewFolderStore create a file based storage for storing the blocks in the
// files, each block is in one file, and also there is a config file, for
// keep track of the last hash in the block
func NewFolderStore(storePath string) Store {
	fs := &folderStore{
		dataPath: storePath,
	}
	lastHashFile := make(map[string][]byte)
	if err := helper.ReadJSON(filepath.Join(fs.dataPath, repository.LastHashFileName), &lastHashFile); err != nil {
		log.Print("Empty store")
		fs.lastHash = nil
	} else if lastHashFile["lastHash"] != nil {
		fs.lastHash = lastHashFile["lastHash"]
	}

	return fs
}
