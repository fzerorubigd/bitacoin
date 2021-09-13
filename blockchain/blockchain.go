package blockchain

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/config"
	"github.com/fzerorubigd/bitacoin/hasher"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/interactor"
	"github.com/fzerorubigd/bitacoin/storege"
	"github.com/fzerorubigd/bitacoin/transaction"
	"log"
	"sync"
)

var LoadedBlockChain *BlockChain

// BlockChain is the group of a block, with difficulty level
type BlockChain struct {
	Mask             []byte
	CancelMining     context.CancelFunc
	difficulty       int
	transactionCount int
	minerPubKey      []byte
	spent            map[string]struct{}
	stackPool        map[string]*transaction.Transaction
	miningPool       map[string]*transaction.Transaction
	sync.Mutex
	storege.Store
}

func (bc *BlockChain) AppendBlock(b *block.Block) error {
	for _, txn := range b.Transactions {
		delete(bc.miningPool, hex.EncodeToString(txn.ID))
	}

	err := bc.Store.AppendBlock(b)
	if err != nil {
		return err
	}

	return nil
}

// MineNewBlock a new data to the end of the blockchain by creating a new block
func (bc *BlockChain) MineNewBlock() (*block.Block, error) {
	log.Println("mining new block has been started")
	ctx, cancel := context.WithCancel(context.Background())
	bc.CancelMining = cancel

	for txnID, txn := range bc.stackPool {
		bc.miningPool[txnID] = txn
	}

	bc.stackPool = make(map[string]*transaction.Transaction)
	bc.spent = make(map[string]struct{})

	transactions := make([]*transaction.Transaction, len(bc.miningPool)+1)
	transactions[0] = transaction.NewRewardTxn(bc.minerPubKey)
	i := 1
	for _, txn := range bc.miningPool {
		transactions[i] = txn
		i++
	}

	hash, err := bc.Store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("Getting the last block failed: %w", err)
	}

	b := block.Mine(ctx, transactions, bc.Mask, hash)

	if b == nil {
		return nil, fmt.Errorf("mining canceled")
	}

	log.Println("new block mined successfully")
	err = interactor.Shout(b)
	if err != nil {
		return nil, err
	}

	if err = bc.Store.AppendBlock(b); err != nil {
		return nil, fmt.Errorf("AppendBlock new block to store failed: %w", err)
	}

	return b, nil
}

// Print the current blockchain in Stdout, it's
func (bc *BlockChain) Print(header bool, count int) error {
	fmt.Printf("difficulty : %d\nStore Backend: %T\n", bc.difficulty, bc.Store)
	if header {
		return nil
	}
	var errEnough = fmt.Errorf("enough")
	err := storege.Iterate(bc.Store, func(b *block.Block) error {
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

// NewBlockChain creates a new blockchain with a difficulty, difficulty in this
// blockchain is the number of zeros in the beginning of the generated hash
func NewBlockChain(pubKey []byte, difficulty, transactionCount int, store storege.Store) (*BlockChain, error) {
	mask := hasher.GenerateMask(difficulty)
	bc := BlockChain{
		difficulty:       difficulty,
		transactionCount: transactionCount,
		Mask:             mask,
		Store:            store,
	}

	_, err := store.LastHash()
	if !errors.Is(err, storege.ErrNotInitialized) {
		return nil, fmt.Errorf("store already initialized")
	}

	txns := make([]*transaction.Transaction, transactionCount)
	for i := range txns {
		txns[i] = transaction.NewRewardTxn(pubKey)
	}
	gb := block.Mine(context.Background(), txns, bc.Mask, []byte{})
	if err = store.AppendBlock(gb); err != nil {
		return nil, fmt.Errorf("Mine Genesis block to store failed: %w", err)
	}

	return &bc, nil
}

// OpenBlockChain tries to open a blockchain, it returns an error if the blockchain store
// is empty
func OpenBlockChain(difficulty, transactionCount int, store storege.Store) error {
	mask := hasher.GenerateMask(difficulty)
	minerPubKey, err := helper.ReadKeyFromPemFile(config.Config.PubKeyPath)
	if err != nil {
		log.Fatalf("read minerPubKey failed err: %s\n", err.Error())
	}

	bc := BlockChain{
		difficulty:       difficulty,
		Mask:             mask,
		transactionCount: transactionCount,
		minerPubKey:      minerPubKey,
		spent:            make(map[string]struct{}),
		stackPool:        make(map[string]*transaction.Transaction),
		miningPool:       make(map[string]*transaction.Transaction),
		Store:            store,
	}

	_, err = store.LastHash()
	if err != nil {
		return fmt.Errorf("openning the store failed: %w", err)
	}
	LoadedBlockChain = &bc
	return nil
}
