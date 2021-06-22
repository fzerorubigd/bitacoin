package blockchain

import (
	"context"
	"errors"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/hasher"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/interactor"
	"github.com/fzerorubigd/bitacoin/repository"
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
	MinerPubKey      []byte
	storege.Store
}

// StartMining a new data to the end of the block chain by creating a new block
func (bc *BlockChain) StartMining(ctx context.Context, transactions ...*transaction.Transaction) (*block.Block, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions to add")
	}
	hash, err := bc.Store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("Getting the last block failed: %w", err)
	}

	coinbaseTnx := transaction.NewCoinBaseTxn(bc.MinerPubKey)
	transactionsPlusCoinbase := make([]*transaction.Transaction, len(transactions)+1)
	transactionsPlusCoinbase[0] = coinbaseTnx
	transactionsPlusCoinbase = append(transactionsPlusCoinbase, transactions...)

	b := block.StartMining(ctx, transactionsPlusCoinbase, bc.Mask, hash)
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

func (bc *BlockChain) UnspentTxn(pubKey []byte) ([]*transaction.UnspentTransaction, int, error) {
	spent := []int{}
	unspentTxns := []*transaction.UnspentTransaction{}
	IndexAmount := make(map[int]int)
	Balance := 0
	err := storege.Iterate(bc.Store, func(b *block.Block) error {
		for _, txn := range b.Transactions {
			for outputCoinIndex, OutputCoin := range txn.OutputCoins {
				if OutputCoin.OwnedBy(pubKey) && !helper.InArray(outputCoinIndex, spent) {
					IndexAmount[outputCoinIndex] = OutputCoin.Amount
					Balance += OutputCoin.Amount
				}
			}

			spent = spent[:]

			if txn.IsCoinBase() {
				continue
			}

			for _, inputCoin := range txn.InputCoins {
				if inputCoin.OwnedBy(pubKey) {
					spent = append(spent, inputCoin.OutputCoinIndex)
				}
			}

			if len(IndexAmount) > 0 {
				unspentTxns = append(unspentTxns, &transaction.UnspentTransaction{
					ID:                     txn.ID,
					OutputCoinsIndexAmount: IndexAmount,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("iterate error: %w", err)
	}

	return unspentTxns, Balance, nil
}

func (bc *BlockChain) NewTransaction(tnxRequest *transaction.TransactionRequest) (*transaction.Transaction, error) {
	lastFourthBlocks := bc.LastFourthBlocks()
	if lastFourthBlocks[len(lastFourthBlocks)-1].Timestamp.After(tnxRequest.Time) {
		return nil, fmt.Errorf("transaction is expired")
	}

	tnxID := transaction.ExtractTxnID(tnxRequest)

	for _, oldBlock := range lastFourthBlocks {
		if oldBlock.Contains(tnxID) {
			return nil, fmt.Errorf("transaction already exist in the blockchain")
		}
	}

	err := transaction.VerifySig(tnxRequest)
	if err != nil {
		return nil, fmt.Errorf("vrify signiture err: %s", err.Error())
	}

	unspentTxns, balance, err := bc.UnspentTxn(tnxRequest.FromPubKey)
	if err != nil {
		return nil, fmt.Errorf("get unused txn failed: %w", err)
	}

	if tnxRequest.Amount <= 0 {
		return nil, fmt.Errorf("amount must be more than 0")
	}

	if balance < tnxRequest.Amount {
		return nil, fmt.Errorf("not enough money, want %d have %d", tnxRequest.Amount, balance)
	}

	var (
		inputCoins []*transaction.InputCoin
		required   = tnxRequest.Amount + repository.TransactionFree
	)

bigLoop:
	for _, unspentTxn := range unspentTxns {
		for outputCoinIndex, outputCoinAmount := range unspentTxn.OutputCoinsIndexAmount {
			required -= outputCoinAmount
			inputCoins = append(inputCoins, &transaction.InputCoin{
				TxnID:           unspentTxn.ID,
				OutputCoinIndex: outputCoinIndex,
				FromPubKey:      tnxRequest.FromPubKey,
			})

			if required <= 0 {
				break bigLoop
			}
		}
	}

	OutputCoins := []*transaction.OutputCoin{
		{
			Amount:   tnxRequest.Amount,
			ToPubKey: tnxRequest.ToPubKey,
		},
		{
			Amount:   repository.TransactionFree,
			ToPubKey: bc.MinerPubKey,
		},
	}

	if required < 0 {
		OutputCoins = append(OutputCoins, &transaction.OutputCoin{
			Amount:   -required,
			ToPubKey: tnxRequest.FromPubKey,
		})
	}

	txn := &transaction.Transaction{
		ID:          tnxID,
		Time:        tnxRequest.Time,
		Sig:         tnxRequest.Signature,
		InputCoins:  inputCoins,
		OutputCoins: OutputCoins,
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
