package blockchain

import (
	"context"
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
	CancelMining     context.CancelFunc
	storege.Store
}

// StartMining a new data to the end of the block chain by creating a new block
func (bc *BlockChain) StartMining(ctx context.Context, data ...*transaction.Transaction) (*block.Block, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to add")
	}
	hash, err := bc.Store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("Getting the last block failed: %w", err)
	}

	b := block.StartMining(ctx, data, bc.Mask, hash)
	err = interactor.Shout(b)
	if err != nil {
		return nil, err
	}

	if err := bc.Store.AppendBlock(b); err != nil {
		return nil, fmt.Errorf("AppendBlock new block to store failed: %w", err)
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

func (bc *BlockChain) UnspentTxn(pubKey []byte) (map[string]*transaction.Transaction, map[string][]int, int, error) {
	spent := make(map[string][]int)
	txom := make(map[string][]int)
	txns := make(map[string]*transaction.Transaction)
	acc := 0
	err := storege.Iterate(bc.Store, func(b *block.Block) error {
		for _, txn := range b.Transactions {
			txnID := hex.EncodeToString(txn.ID)

			for i := range txn.OutputCoins {
				if txn.OutputCoins[i].OwnedBy(pubKey) && !helper.InArray(i, spent[txnID]) {
					txns[txnID] = txn
					txom[txnID] = append(txom[txnID], i)
					acc += txn.OutputCoins[i].Amount
				}
			}

			delete(spent, txnID)

			if txn.IsCoinBase() {
				continue
			}

			for i := range txn.InputCoins {
				if txn.InputCoins[i].OwnedBy(pubKey) {
					outID := hex.EncodeToString(txn.InputCoins[i].TXID)
					spent[outID] = append(spent[outID], txn.InputCoins[i].OutputTransactionIndex)
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

func (bc *BlockChain) NewTransaction(tnxRequest *transaction.TransactionRequest) (*transaction.Transaction, error) {
	err := transaction.VerifySig(tnxRequest)
	if err != nil {
		return nil, fmt.Errorf("vrify signiture err: %s", err.Error())
	}

	tnxID := transaction.ExtractTxnID(tnxRequest)

	txns, txnsOut, acc, err := bc.UnspentTxn(tnxRequest.FromPubKey)
	if err != nil {
		return nil, fmt.Errorf("get unused txn failed: %w", err)
	}

	if tnxRequest.Amount <= 0 {
		return nil, fmt.Errorf("amount must be more than 0")
	}

	if acc < tnxRequest.Amount {
		return nil, fmt.Errorf("not enough money, want %d have %d", tnxRequest.Amount, acc)
	}

	var (
		vin      []transaction.InputCoin
		required = tnxRequest.Amount
	)

bigLoop:
	for id, txn := range txns {
		for _, v := range txnsOut[id] {
			required -= txn.OutputCoins[v].Amount
			vin = append(vin, transaction.InputCoin{
				TXID:                   txn.ID,
				OutputTransactionIndex: v,
				FromPubKey:             tnxRequest.FromPubKey,
			})

			if required <= 0 {
				break bigLoop
			}
		}
	}

	vout := []transaction.OutputCoin{
		{
			Amount:   tnxRequest.Amount,
			ToPubKey: tnxRequest.ToPubKey,
		},
	}
	if required < 0 {
		vout = append(vout, transaction.OutputCoin{
			Amount:   -required,
			ToPubKey: tnxRequest.FromPubKey,
		})
	}

	txn := &transaction.Transaction{
		ID:          tnxID,
		Time:        tnxRequest.Time,
		InputCoins:  vin,
		OutputCoins: vout,
		Sig:         tnxRequest.Signature,
	}

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
	gbTxn := transaction.NewCoinBaseTxn(genesis)
	gb := block.StartMining(context.Background(), []*transaction.Transaction{gbTxn}, bc.Mask, []byte{})
	if err := store.AppendBlock(gb); err != nil {
		return nil, fmt.Errorf("StartMining Genesis block to store failed: %w", err)
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
