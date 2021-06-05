package blockchain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/hasher"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/interactor"
	"github.com/fzerorubigd/bitacoin/storege"
	"github.com/fzerorubigd/bitacoin/transaction"
)

var LoadedBlockChain *BlockChain

// BlockChain is the group of a block, with difficulty level
type BlockChain struct {
	Difficulty       int
	Mask             []byte
	TransactionCount int

	storege.Store
}

// MineNewBlock a new data to the end of the block chain by creating a new block
func (bc *BlockChain) MineNewBlock(data ...*transaction.Transaction) (*block.Block, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to add")
	}
	hash, err := bc.Store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("Getting the last block failed: %w", err)
	}

	b := block.Mine(data, bc.Mask, hash)
	err = interactor.Shout(b)
	if err != nil {
		return nil, err
	}

	if err := bc.Store.Append(b); err != nil {
		return nil, fmt.Errorf("Append new block to store failed: %w", err)
	}

	return b, nil
}

// Print the current blockchain in Stdout, it's
func (bc *BlockChain) Print(header bool, count int) error {
	fmt.Printf("Difficulty : %d\nStore Backend: %T\n", bc.Difficulty, bc.Store)
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

// Validate all data in the block chain
func (bc *BlockChain) Validate() error {
	return storege.Iterate(bc.Store, func(b *block.Block) error {
		if err := b.Validate(bc.Mask); err != nil {
			return fmt.Errorf("block chain is not valid: %w", err)
		}

		return nil
	})
}

func (bc *BlockChain) UnspentTxn(address []byte) (map[string]*transaction.Transaction, map[string][]int, int, error) {
	spent := make(map[string][]int)
	txom := make(map[string][]int)
	txns := make(map[string]*transaction.Transaction)
	acc := 0
	err := storege.Iterate(bc.Store, func(b *block.Block) error {
		for _, txn := range b.Transactions {
			txnID := hex.EncodeToString(txn.ID)

			for i := range txn.VOut {
				if txn.VOut[i].TryUnlock(address) && !helper.InArray(i, spent[txnID]) {
					txns[txnID] = txn
					txom[txnID] = append(txom[txnID], i)
					acc += txn.VOut[i].Value
				}
			}

			delete(spent, txnID)

			if txn.IsCoinBase() {
				continue
			}

			for i := range txn.VIn {
				if txn.VIn[i].MatchLock(address) {
					outID := hex.EncodeToString(txn.VIn[i].TXID)
					spent[outID] = append(spent[outID], txn.VIn[i].VOut)
				}
			}

		}

		return nil
	})
	if err != nil {
		return nil, nil, 0, fmt.Errorf("iterate error: %w", err)
	}

	return txns, txom, acc, nil
}

func (bc *BlockChain) NewTransaction(from, to []byte, amount int) (*transaction.Transaction, error) {
	txns, txom, acc, err := bc.UnspentTxn(from)
	if err != nil {
		return nil, fmt.Errorf("get unused txn failed: %w", err)
	}

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be more than 0")
	}

	if acc < amount {
		return nil, fmt.Errorf("not enough money, want %d have %d", amount, acc)
	}

	var (
		vin      []transaction.TXInput
		required = amount
	)

bigLoop:
	for id, txn := range txns {
		for _, v := range txom[id] {
			required -= txn.VOut[v].Value
			vin = append(vin, transaction.TXInput{
				TXID: txn.ID,
				VOut: v,
				Sig:  from, // TODO : real sign
			})

			if required <= 0 {
				break bigLoop
			}
		}
	}

	vout := []transaction.TXOutput{
		{
			Value:  amount,
			PubKey: to,
		},
	}
	if required < 0 {
		vout = append(vout, transaction.TXOutput{
			Value:  -required,
			PubKey: from,
		})
	}

	txn := &transaction.Transaction{
		VIn:  vin,
		VOut: vout,
	}

	txn.ID = transaction.CalculateTxnID(txn)

	return txn, nil
}

// NewBlockChain creates a new block chain with a difficulty, difficulty in this
// block chain is the number of zeros in the beginning of the generated hash
func NewBlockChain(genesis []byte, difficulty, transactionCount int, store storege.Store) (*BlockChain, error) {
	mask := hasher.GenerateMask(difficulty)
	bc := BlockChain{
		Difficulty:       difficulty,
		TransactionCount: transactionCount,
		Mask:             mask,
		Store:            store,
	}

	_, err := store.LastHash()
	if !errors.Is(err, storege.ErrNotInitialized) {
		return nil, fmt.Errorf("store already initialized")
	}
	gbTxn := transaction.NewCoinBaseTxn(genesis, nil)
	gb := block.Mine([]*transaction.Transaction{gbTxn}, bc.Mask, []byte{})
	if err := store.Append(gb); err != nil {
		return nil, fmt.Errorf("MineNewBlock Genesis block to store failed: %w", err)
	}

	return &bc, nil
}

// OpenBlockChain tries to open a blockchain, it returns an error if the blockchain store
// is empty
func OpenBlockChain(difficulty, transactionCount int, store storege.Store) (*BlockChain, error) {
	mask := hasher.GenerateMask(difficulty)
	bc := BlockChain{
		Difficulty:       difficulty,
		TransactionCount: transactionCount,
		Mask:             mask,
		Store:            store,
	}

	_, err := store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("openning the store failed: %w", err)
	}
	LoadedBlockChain = &bc
	return &bc, nil
}
