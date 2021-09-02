package bitacoin

import (
	"errors"
	"github.com/boltdb/bolt"
	"log"
)

const dbName = "my.db"

var blockBucket = []byte("blocks")

const genesisData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type BoltStore struct {
}

func (bol *BoltStore) Load(hash []byte) (*Block, error) {
	var block Block
	db := initdb()
	defer db.Close()
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(blockBucket)
		blockData := b.Get(hash)
		if blockData == nil {
			return errors.New("Block not found")
		}
		block = *DeserializeBlock(blockData)
		return nil
	})

	if err != nil {
		return &block, err
	}

	return &block, nil
}

func (bol *BoltStore) Append(block *Block) error {
	db := initdb()
	defer db.Close()
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(blockBucket)
		blockCheck := b.Get(block.Hash)
		if blockCheck != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("lasthash"), block.Hash)
		if err != nil {
			log.Panic(err)
		}
		return err
	})
	if err != nil {
		log.Panic(err)
	}

	return nil
}

func (bol *BoltStore) LastHash() ([]byte, error) {
	db := initdb()
	defer db.Close()
	var hash []byte
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(blockBucket)
		x := b.Get([]byte("lasthash"))

		if x == nil {
			return ErrNotInitialized
		}

		//to fix a wired bug of boltDB
		ng := make([]byte, len(x))
		copy(ng, x)
		hash = ng

		return nil
	})
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func initdb() *bolt.DB {
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}

	tx, err := db.Begin(true)
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists(blockBucket)

	if err != nil {
		log.Fatalln(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalln(err)
	}

	return db

}

func NewDBStore() Store {
	d := &BoltStore{}
	return d
}
