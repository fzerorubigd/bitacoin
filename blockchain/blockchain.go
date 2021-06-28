package blockchain

import (
	"bytes"
	"context"
	"encoding/hex"
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

	rewardTnx := transaction.NewRewardTxn(bc.MinerPubKey)
	transactionsPlusCoinbase := make([]*transaction.Transaction, len(transactions)+1)
	transactionsPlusCoinbase[0] = rewardTnx
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

func (bc *BlockChain) ValidateIncomingTransactions(transactions []*transaction.Transaction) error {
	rewardCoinCount := 0
	for index, txn := range transactions[1:] {
		if len(txn.InputCoins) < 1 || len(txn.OutputCoins) < 1 {
			return fmt.Errorf("transaction needs at least one inputCoin and one outputCoin, txn index: %d", index)
		}

		if len(txn.OutputCoins) > 1 {
			if txn.OutputCoins[len(txn.OutputCoins)-1].Amount != repository.TransactionFree {
				return fmt.Errorf("transaction fee is %d, it must be %d, txn index: %d",
					txn.OutputCoins[1].Amount, repository.TransactionFree, index)
			}
		}

		txnID := transaction.ExtractTxnID(txn.Sig, txn.Time)
		if !bytes.Equal(txnID, txn.ID) {
			return fmt.Errorf("transaction id is wrong, txn index: %d", index)
		}

		if txn.IsCoinBase() {
			rewardCoinCount++
			if rewardCoinCount > 1 {
				return fmt.Errorf("there has to be only one reward transaction in a block, txn index: %d", index)
			}
		} else if len(txn.OutputCoins) < 2 {
			return fmt.Errorf("there is no transaction fee, txn index: %d", index)
		} else {
			fromPubKey := txn.InputCoins[0].PubKey
			err := transaction.VerifySig(txn.Time, fromPubKey,
				txn.OutputCoins[0].PubKey, txn.OutputCoins[0].Amount, txn.Sig)
			if err != nil {
				return fmt.Errorf("verify siniture err: %s, txn index: %d", err.Error(), index)
			}

			unspentTxns, balance, err := bc.UnspentTxn(txn.InputCoins[0].PubKey)
			if err != nil {
				return fmt.Errorf("balance err: %s, txn index: %d", err.Error(), index)
			}

			if balance < txn.OutputCoins[0].Amount {
				return fmt.Errorf("balance is lower than amount, txn index: %d", index)
			}

			if balance > txn.OutputCoins[0].Amount && (len(txn.OutputCoins) != 3 ||
				txn.OutputCoins[2].Amount != balance-txn.OutputCoins[0].Amount || !txn.OutputCoins[2].OwnedBy(fromPubKey)) {
				return fmt.Errorf("rest balance in outputCoin is wrong, txn index: %d", index)
			}

			strTxnID := hex.EncodeToString(txnID)
			for _, inputCoin := range txn.InputCoins {
				if !inputCoin.OwnedBy(fromPubKey) {
					return fmt.Errorf("spent someone else coin, wrong inputCoin, txn index: %d", index)
				}
				if !unspentTxns[strTxnID][inputCoin.OutputCoinIndex].OwnedBy(fromPubKey) {
					return fmt.Errorf("spent someone else coin, wrong outputCoin, txn index: %d", index)
				}
			}
		}
	}

	return nil
}

func (bc *BlockChain) UnspentTxn(pubKey []byte) (map[string]map[int]*transaction.OutputCoin, int, error) {
	spent := []int{}
	unspent := make(map[string]map[int]*transaction.OutputCoin)
	balance := 0
	err := storege.Iterate(bc.Store, func(b *block.Block) error {
		for _, txn := range b.Transactions {
			for outputCoinIndex, OutputCoin := range txn.OutputCoins {
				if OutputCoin.OwnedBy(pubKey) && !helper.InArray(outputCoinIndex, spent) {
					unspent[hex.EncodeToString(txn.ID)][outputCoinIndex] = OutputCoin
					balance += OutputCoin.Amount
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
		}

		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("iterate error: %w", err)
	}

	return unspent, balance, nil
}

func (bc *BlockChain) NewTransaction(tnxRequest *transaction.TransactionRequest) (*transaction.Transaction, error) {
	lastFourthBlocks := bc.LastFourthBlocks()
	if lastFourthBlocks[len(lastFourthBlocks)-1].Timestamp.After(tnxRequest.Time) {
		return nil, fmt.Errorf("transaction is expired")
	}

	tnxID := transaction.ExtractTxnID(tnxRequest.Signature, tnxRequest.Time)

	for _, oldBlock := range lastFourthBlocks {
		if oldBlock.Contains(tnxID) {
			return nil, fmt.Errorf("transaction already exist in the blockchain")
		}
	}

	err := transaction.VerifySig(tnxRequest.Time, tnxRequest.FromPubKey,
		tnxRequest.ToPubKey, tnxRequest.Amount, tnxRequest.Signature)
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

	required := tnxRequest.Amount + repository.TransactionFree
	if balance < required {
		return nil, fmt.Errorf("not enough money, want %d have %d", required, balance)
	}

	inputCoins := []*transaction.InputCoin{}

bigLoop:
	for strTxnID, outputCoins := range unspentTxns {
		txnId, _ := hex.DecodeString(strTxnID)
		for outputCoinIndex, outputCoin := range outputCoins {
			required -= outputCoin.Amount
			inputCoins = append(inputCoins, &transaction.InputCoin{
				TxnID:           txnId,
				OutputCoinIndex: outputCoinIndex,
				PubKey:          tnxRequest.FromPubKey,
			})

			if required <= 0 {
				break bigLoop
			}
		}
	}

	OutputCoins := []*transaction.OutputCoin{
		{
			Amount: tnxRequest.Amount,
			PubKey: tnxRequest.ToPubKey,
		},
		{
			Amount: repository.TransactionFree,
			PubKey: bc.MinerPubKey,
		},
	}

	if required < 0 {
		OutputCoins = append(OutputCoins, &transaction.OutputCoin{
			Amount: -required,
			PubKey: tnxRequest.FromPubKey,
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
	gbTxn := transaction.NewRewardTxn(genesis)
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
