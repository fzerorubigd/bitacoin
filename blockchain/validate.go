package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/repository"
	"github.com/fzerorubigd/bitacoin/storege"
	"github.com/fzerorubigd/bitacoin/transaction"
	"strconv"
)

// Validate all data in the blockchain
func (bc *BlockChain) Validate() error {
	return storege.Iterate(bc.Store, func(b *block.Block) error {
		if err := b.Validate(bc.Mask); err != nil {
			return fmt.Errorf("blockchain is not valid: %w", err)
		}

		return nil
	})
}

func (bc *BlockChain) ValidateIncomingTransactions(transactions []*transaction.Transaction) error {
	rewardCoinCount := 0
	spent := make(map[string]struct{})
	for index, txn := range transactions {
		if len(txn.InputCoins) < 1 {
			return fmt.Errorf("transaction needs at least one inputCoin, transaction index: %d", index)
		}

		if len(txn.OutputCoins) < 1 {
			return fmt.Errorf("transaction needs at least one outputCoin, transaction index: %d", index)
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
			return fmt.Errorf("there must be more than one OutputCoin, transaction index: %d", index)
		} else {
			fromPubKey := txn.InputCoins[0].PubKey
			txnFeeOutputCoin := txn.OutputCoins[0]
			transferredOutputCoin := txn.OutputCoins[1]

			err := transaction.VerifySig(txn.Time, fromPubKey,
				transferredOutputCoin.PubKey, transferredOutputCoin.Amount, txn.Sig)
			if err != nil {
				return fmt.Errorf("verify siniture err: %s, transaction index: %d", err.Error(), index)
			}

			if repository.TransactionFree != txnFeeOutputCoin.Amount {
				return fmt.Errorf("wrong outputCoin transaction fee, expected %d but recieved: %d, transaction index: %d",
					repository.TransactionFree, txnFeeOutputCoin.Amount, index)
			}

			unspent, balance, err := bc.UnspentTxn(fromPubKey)
			if err != nil {
				return fmt.Errorf("balance err: %s, transaction index: %d", err.Error(), index)
			}

			if balance < transferredOutputCoin.Amount {
				return fmt.Errorf("transaction is inposible because pubKey does not own enoght balance, transaction index: %d", index)
			}

			inputTotalAmount := 0
			for _, inputCoin := range txn.InputCoins {
				inputTxnStrId := hex.EncodeToString(inputCoin.TxnID)
				if !inputCoin.OwnedBy(fromPubKey) {
					return fmt.Errorf("spent someone else coin, wrong inputCoin, transaction index: %d", index)
				}
				outputCoin, ok := unspent[inputTxnStrId][inputCoin.OutputCoinIndex]
				if !ok {
					return fmt.Errorf("unspent outputCoinIndex: %d does not exist in transaction: %x, transaction index: %d",
						inputCoin.OutputCoinIndex, inputCoin.TxnID, index)
				}
				if !outputCoin.OwnedBy(fromPubKey) {
					return fmt.Errorf("spent someone else outputCoin, outputCoin Index: %d, transaction index: %d",
						inputCoin.OutputCoinIndex, index)
				}
				inputTotalAmount += unspent[inputTxnStrId][inputCoin.OutputCoinIndex].Amount
				spendKey := inputTxnStrId + strconv.Itoa(inputCoin.OutputCoinIndex)
				if _, ok = spent[spendKey]; ok {
					return fmt.Errorf("double spend err, transaction index: %d", index)
				} else {
					spent[spendKey] = struct{}{}
				}
			}

			restBalance := inputTotalAmount - (transferredOutputCoin.Amount + repository.TransactionFree)
			if restBalance == 0 && len(txn.OutputCoins) > 2 {
				return fmt.Errorf("resetOutputCoin must not exist but it does, transaction index: %d", index)
			} else if restBalance > 0 && len(txn.OutputCoins) < 3 {
				return fmt.Errorf("resetOutputCoin must exist but it does not, transaction index: %d", index)
			} else if restBalance > 0 && len(txn.OutputCoins) > 2 {
				resetOutputCoin := txn.OutputCoins[len(txn.OutputCoins)-1]
				if resetOutputCoin == nil {
					return fmt.Errorf("resetOutputCoin is nil, transaction index: %d", index)
				}

				if !resetOutputCoin.OwnedBy(fromPubKey) {
					return fmt.Errorf("resetOutputCoin has wrong pubKey, transaction index: %d", index)
				}

				if resetOutputCoin.Amount != restBalance {
					return fmt.Errorf("resetOutputCoin has wrong amout, expected %d but recieved: %d, transaction index: %d",
						restBalance, resetOutputCoin.Amount, index)
				}
			}
		}
	}

	return nil
}
