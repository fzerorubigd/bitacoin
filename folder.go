package bitacoin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func (fs *folderStore) Load(hash []byte) (*Block, error) {
	path := filepath.Join(fs.root, fmt.Sprintf("%x.json", hash))
	var b Block
	if err := readJSON(path, &b); err != nil {
		return nil, fmt.Errorf("read JOSN file failed: %w", err)
	}

	return &b, nil
}

func (fs *folderStore) Append(b *Block) error {
	if !bytes.Equal(fs.config.LastHash, b.PrevHash) {
		return fmt.Errorf("store is out of sync")
	}

	path := filepath.Join(fs.root, fmt.Sprintf("%x.json", b.Hash))
	if err := writeJSON(path, b); err != nil {
		return fmt.Errorf("write JSON file failed: %w", err)
	}

	fs.config.LastHash = b.Hash
	if err := writeJSON(fs.configPath, fs.config); err != nil {
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

func readJSON(path string, v interface{}) error {
	fl, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer fl.Close()

	dec := json.NewDecoder(fl)

	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("decode JSON content failed: %w", err)
	}

	return nil
}

func writeJSON(path string, v interface{}) error {
	// TODO : fail if file exists
	fl, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer fl.Close()

	enc := json.NewEncoder(fl)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding the JSON content failed: %w", err)
	}

	return nil
}

// NewFolderStore create a file based storage for storing the blocks in the
// files, each block is in one file, and also there is a config file, for
// keep track of the last hash in the block
func NewFolderStore(root string) Store {
	fs := &folderStore{
		root:       root,
		config:     &folderConfig{},
		configPath: filepath.Join(root, "config.json"),
	}

	if err := readJSON(fs.configPath, fs.config); err != nil {
		log.Print("Empty store")
		fs.config.LastHash = nil
	}

	return fs
}
