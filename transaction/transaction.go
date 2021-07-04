package transaction

import (
	"github.com/fzerorubigd/bitacoin/hasher"
	"github.com/fzerorubigd/bitacoin/repository"
	"time"
)

import "bytes"

type TransactionRequest struct {
	Time       int64
	FromPubKey []byte
	ToPubKey   []byte
	Signature  []byte
	Amount     int
}

type Transaction struct {
	Time        int64
	Sig         []byte
	ID          []byte
	InputCoins  []*InputCoin
	OutputCoins []*OutputCoin
}

type OutputCoin struct {
	Amount int
	PubKey []byte
}

type InputCoin struct {
	TxnID           []byte
	OutputCoinIndex int
	PubKey          []byte
}

func (txn *Transaction) IsCoinBase() bool {
	return len(txn.OutputCoins) == 1 &&
		len(txn.InputCoins) == 1 &&
		txn.InputCoins[0].OutputCoinIndex == -1 &&
		len(txn.InputCoins[0].TxnID) == 0 &&
		bytes.Equal(txn.Sig, repository.Sig)
}

func (txo *OutputCoin) OwnedBy(key []byte) bool {
	return bytes.Equal(txo.PubKey, key)
}

func (txi *InputCoin) OwnedBy(key []byte) bool {
	return bytes.Equal(txi.PubKey, key)
}

func NewRewardTxn(toPubKey []byte) *Transaction {
	txi := InputCoin{
		TxnID:           []byte{},
		OutputCoinIndex: -1,
	}

	txo := OutputCoin{
		Amount: repository.CoinbaseReward,
		PubKey: toPubKey,
	}

	t := time.Now().UnixNano()
	txn := &Transaction{
		Time:        t,
		InputCoins:  []*InputCoin{&txi},
		OutputCoins: []*OutputCoin{&txo},
	}

	txn.Sig = repository.Sig
	txn.ID = CalculateTxnID(txn.Sig, txn.Time)

	return txn
}

func CalculateTxnsHash(txns ...*Transaction) []byte {
	data := make([]interface{}, len(txns))
	for i := range txns {
		data[i] = txns[i].ID
	}

	return hasher.EasyHash(data...)
}

func CalculateTxnID(sin []byte, time int64) []byte {
	return hasher.EasyHash(sin, time)
}
