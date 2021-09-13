package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/repository"
	"github.com/fzerorubigd/bitacoin/storege"
	"github.com/fzerorubigd/bitacoin/transaction"
	"log"
	"strconv"
)

func (bc *BlockChain) existInPool(txnID string) bool {
	if _, ok := bc.miningPool[txnID]; ok {
		return ok
	}

	if _, ok := bc.stackPool[txnID]; ok {
		return ok
	}

	return false
}

func (bc *BlockChain) AddToMemPool(txn *transaction.Transaction) error {
	txnID := hex.EncodeToString(txn.ID)
	if bc.existInPool(txnID) {
		return fmt.Errorf("transaction alread exist in pool")
	}

	bc.Lock()
	bc.stackPool[txnID] = txn
	bc.Unlock()

	if len(bc.stackPool) >= bc.transactionCount && bc.CancelMining == nil {
		go func() {
			newBlock, err := bc.MineNewBlock()
			if err != nil {
				log.Printf("Mine err: %s\n", err.Error())
			} else {
				log.Printf("new block added to the blockchain successfully.\n%s\n", newBlock.String())
			}
			bc.CancelMining = nil
			bc.miningPool = make(map[string]*transaction.Transaction)
		}()
	}

	return nil
}

func (bc *BlockChain) UnspentTxn(pubKey []byte) (map[string]map[int]*transaction.OutputCoin, int, error) {
	spent := make(map[string]struct{})
	unspent := make(map[string]map[int]*transaction.OutputCoin)
	balance := 0
	err := storege.Iterate(bc.Store, func(b *block.Block) error {
		for _, txn := range b.Transactions {
			txnID := hex.EncodeToString(txn.ID)
			for outputCoinIndex, OutputCoin := range txn.OutputCoins {
				_, spentInOtherBlockTxns := bc.spent[txnID+strconv.Itoa(outputCoinIndex)]
				if _, ok := spent[txnID+strconv.Itoa(outputCoinIndex)]; !spentInOtherBlockTxns &&
					!ok && OutputCoin.OwnedBy(pubKey) {
					if _, ok = unspent[txnID]; !ok {
						unspent[txnID] = make(map[int]*transaction.OutputCoin)
					}
					unspent[txnID][outputCoinIndex] = OutputCoin
					balance += OutputCoin.Amount
				}
			}

			delete(spent, txnID)

			if txn.IsCoinBase() {
				continue
			}

			for _, inputCoin := range txn.InputCoins {
				if inputCoin.OwnedBy(pubKey) {
					spent[hex.EncodeToString(inputCoin.TxnID)+strconv.Itoa(inputCoin.OutputCoinIndex)] = struct{}{}
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

func (bc *BlockChain) NewTxn(txnRequest *transaction.TransactionRequest) (*transaction.Transaction, error) {
	lastFourthBlocks := bc.LastFourthBlocks()
	if len(lastFourthBlocks) == 4 && lastFourthBlocks[3] != nil && lastFourthBlocks[3].Time > txnRequest.Time {
		return nil, fmt.Errorf("transaction is expired")
	}

	txnID := transaction.CalculateTxnID(txnRequest.Signature, txnRequest.Time)

	for _, oldBlock := range lastFourthBlocks {
		if oldBlock != nil && oldBlock.Contains(txnID) {
			return nil, fmt.Errorf("transaction already exist in the blockchain")
		}
	}

	err := transaction.VerifySig(txnRequest.Time, txnRequest.FromPubKey,
		txnRequest.ToPubKey, txnRequest.Amount, txnRequest.Signature)
	if err != nil {
		return nil, fmt.Errorf("vrify signiture err: %s", err.Error())
	}

	unspentTxns, balance, err := bc.UnspentTxn(txnRequest.FromPubKey)
	if err != nil {
		return nil, fmt.Errorf("get unused txn failed: %w", err)
	}

	if txnRequest.Amount <= 0 {
		return nil, fmt.Errorf("amount must be more than 0")
	}

	required := txnRequest.Amount + repository.TransactionFree
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
				PubKey:          txnRequest.FromPubKey,
			})
			bc.spent[strTxnID+strconv.Itoa(outputCoinIndex)] = struct{}{}

			if required <= 0 {
				break bigLoop
			}
		}
	}

	OutputCoins := []*transaction.OutputCoin{
		{
			Amount: repository.TransactionFree,
			PubKey: bc.minerPubKey,
		},
		{
			Amount: txnRequest.Amount,
			PubKey: txnRequest.ToPubKey,
		},
	}

	if required < 0 {
		OutputCoins = append(OutputCoins, &transaction.OutputCoin{
			Amount: -required,
			PubKey: txnRequest.FromPubKey,
		})
	}

	txn := &transaction.Transaction{
		ID:          txnID,
		Time:        txnRequest.Time,
		Sig:         txnRequest.Signature,
		InputCoins:  inputCoins,
		OutputCoins: OutputCoins,
	}

	return txn, nil
}
