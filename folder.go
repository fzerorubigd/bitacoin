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
		return nil, err
	}

	return &b, nil
}

func (fs *folderStore) Append(b *Block) error {
	if !bytes.Equal(fs.config.LastHash, b.PrevHash) {
		return fmt.Errorf("store is out of sync")
	}

	path := filepath.Join(fs.root, fmt.Sprintf("%x.json", b.Hash))
	if err := writeJSON(path, b); err != nil {
		return err
	}

	fs.config.LastHash = b.Hash
	if err := writeJSON(fs.configPath, fs.config); err != nil {
		return err
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
		return err
	}
	defer fl.Close()

	dec := json.NewDecoder(fl)

	if err := dec.Decode(v); err != nil {
		return err
	}

	return nil
}

func writeJSON(path string, v interface{}) error {
	// TODO : fail if file exists
	fl, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fl.Close()

	enc := json.NewEncoder(fl)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}

	return nil
}

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
