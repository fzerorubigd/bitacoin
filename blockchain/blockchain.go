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
	"log"
	"strconv"
)

var LoadedBlockChain *BlockChain

// BlockChain is the group of a block, with difficulty level
type BlockChain struct {
	Difficulty       int
	Mask             []byte
	TransactionCount int
	CancelMining     context.CancelFunc
	MinerPubKey      []byte
	Spent            map[string]struct{}
	memPool          map[string]*transaction.Transaction
	storege.Store
}

func (bc *BlockChain) AddToMemPool(txn *transaction.Transaction) {
	txnID := hex.EncodeToString(txn.ID)
	bc.memPool[txnID] = txn
	if len(bc.memPool) >= bc.TransactionCount && bc.CancelMining == nil {
		ctx, cancel := context.WithCancel(context.Background())
		bc.CancelMining = cancel
		bc.memPool = make(map[string]*transaction.Transaction)
		go func() {
			newBlock, err := bc.StartMining(ctx)
			if err != nil {
				log.Printf("StartMining err: %s\n", err.Error())
			} else {
				log.Printf("new block added to the blockchain successfully.\n%s\n", newBlock.String())
			}
			bc.CancelMining = nil
		}()
	}
}

// StartMining a new data to the end of the block chain by creating a new block
func (bc *BlockChain) StartMining(ctx context.Context) (*block.Block, error) {
	log.Println("mining new block has been started")
	bc.Spent = make(map[string]struct{})

	transactions := make([]*transaction.Transaction, len(bc.memPool)+1)
	transactions[0] = transaction.NewRewardTxn(bc.MinerPubKey)
	i := 1
	for _, txn := range bc.memPool {
		transactions[i] = txn
	}
	bc.memPool = make(map[string]*transaction.Transaction)

	hash, err := bc.Store.LastHash()
	if err != nil {
		return nil, fmt.Errorf("Getting the last block failed: %w", err)
	}

	b := block.StartMining(ctx, transactions, bc.Mask, hash)
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
			return fmt.Errorf("transaction needs at least one inputCoin and one outputCoin, transaction index: %d", index)
		}

		if len(txn.OutputCoins) > 1 {
			if txn.OutputCoins[1].Amount != repository.TransactionFree {
				return fmt.Errorf("transaction fee is %d, it must be %d, transactions index: %d",
					txn.OutputCoins[1].Amount, repository.TransactionFree, index)
			}
		}

		txnID := transaction.CalculateTxnID(txn.Sig, txn.Time)
		if !bytes.Equal(txnID, txn.ID) {
			return fmt.Errorf("transaction id is wrong, transaction index: %d", index)
		}

		if txn.IsCoinBase() {
			rewardCoinCount++
			if rewardCoinCount > 1 {
				return fmt.Errorf("there has to be only one reward transaction in a block, transaction index: %d", index)
			}
		} else if len(txn.OutputCoins) < 2 {
			return fmt.Errorf("there is no transaction fee, transaction index: %d", index)
		} else {
			fromPubKey := txn.InputCoins[0].PubKey
			err := transaction.VerifySig(txn.Time, fromPubKey,
				txn.OutputCoins[0].PubKey, txn.OutputCoins[0].Amount, txn.Sig)
			if err != nil {
				return fmt.Errorf("verify siniture err: %s, transaction index: %d", err.Error(), index)
			}

			unspentTxns, balance, err := bc.UnspentTxn(txn.InputCoins[0].PubKey)
			if err != nil {
				return fmt.Errorf("balance err: %s, transaction index: %d", err.Error(), index)
			}

			if balance < txn.OutputCoins[0].Amount {
				return fmt.Errorf("balance is lower than amount, transaction index: %d", index)
			}

			if len(txn.OutputCoins) != 3 && balance > txn.OutputCoins[0].Amount {
				return fmt.Errorf("rest balance coin doesn't exist in outputCoin, transaction index: %d", index)
			}

			inputTotalAmount := 0
			for _, inputCoin := range txn.InputCoins {
				if !inputCoin.OwnedBy(fromPubKey) {
					return fmt.Errorf("spent someone else coin, wrong inputCoin, transaction index: %d", index)
				}
				outputCoin := unspentTxns[hex.EncodeToString(inputCoin.TxnID)][inputCoin.OutputCoinIndex]
				if !outputCoin.OwnedBy(fromPubKey) {
					return fmt.Errorf("spent someone else coin, wrong outputCoin, transaction index: %d", index)
				}
				inputTotalAmount += unspentTxns[hex.EncodeToString(inputCoin.TxnID)][inputCoin.OutputCoinIndex].Amount
			}

			restBalance := inputTotalAmount - (txn.OutputCoins[0].Amount + txn.OutputCoins[1].Amount)
			if txn.OutputCoins[2].Amount != restBalance {
				return fmt.Errorf("rest balance coin has wrong amout, expected %d but recieved: %d, transaction index: %d",
					restBalance, txn.OutputCoins[2].Amount, index)
			}

			if txn.OutputCoins[2] == nil || !txn.OutputCoins[2].OwnedBy(fromPubKey) {
				return fmt.Errorf("rest balance coin has wrong pubKey, transaction index: %d", index)
			}
		}
	}

	return nil
}

func (bc *BlockChain) UnspentTxn(pubKey []byte) (map[string]map[int]*transaction.OutputCoin, int, error) {
	spent := make(map[string][]int)
	unspent := make(map[string]map[int]*transaction.OutputCoin)
	balance := 0
	err := storege.Iterate(bc.Store, func(b *block.Block) error {
		for _, txn := range b.Transactions {
			txnID := hex.EncodeToString(txn.ID)
			for outputCoinIndex, OutputCoin := range txn.OutputCoins {
				if _, ok := bc.Spent[txnID+strconv.Itoa(outputCoinIndex)]; !ok &&
					OutputCoin.OwnedBy(pubKey) && !helper.InArray(outputCoinIndex, spent[txnID]) {
					if _, ok := unspent[txnID]; !ok {
						unspent[txnID] = make(map[int]*transaction.OutputCoin)
					}
					unspent[txnID][outputCoinIndex] = OutputCoin
					balance += OutputCoin.Amount
					bc.Spent[txnID+strconv.Itoa(outputCoinIndex)] = struct{}{}
				}
			}

			delete(spent, txnID)

			if txn.IsCoinBase() {
				continue
			}

			for _, inputCoin := range txn.InputCoins {
				if inputCoin.OwnedBy(pubKey) {
					spent[hex.EncodeToString(inputCoin.TxnID)] = append(spent[hex.EncodeToString(inputCoin.TxnID)], inputCoin.OutputCoinIndex)
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
	if len(lastFourthBlocks) == 4 && lastFourthBlocks[3] != nil && lastFourthBlocks[3].Time > tnxRequest.Time {
		return nil, fmt.Errorf("transaction is expired")
	}

	tnxID := transaction.CalculateTxnID(tnxRequest.Signature, tnxRequest.Time)

	for _, oldBlock := range lastFourthBlocks {
		if oldBlock != nil && oldBlock.Contains(tnxID) {
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
		return nil, fmt.Errorf("not enough balance, want %d have %d", required, balance)
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
func NewBlockChain(pubKey []byte, difficulty, transactionCount int, store storege.Store) (*BlockChain, error) {
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

	gbTxn := transaction.NewRewardTxn(pubKey)
	gbTxn.Sig = []byte("Today is same tomorrow that was supposed to be better than yesterday.")
	gbTxn.ID = transaction.CalculateTxnID(gbTxn.Sig, gbTxn.Time)

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
