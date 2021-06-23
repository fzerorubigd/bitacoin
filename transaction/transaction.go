package transaction

import (
	"fmt"
	"github.com/fzerorubigd/bitacoin/hasher"
	"github.com/fzerorubigd/bitacoin/repository"
	"time"
)

import "bytes"

type TransactionRequest struct {
	Time       time.Time
	FromPubKey []byte
	ToPubKey   []byte
	Signature  []byte
	Amount     int
}

type Transaction struct {
	Time        time.Time
	ID          []byte
	Sig         []byte
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
		len(txn.InputCoins[0].TxnID) == 0
}

func (txo *OutputCoin) OwnedBy(key []byte) bool {
	return bytes.Equal(txo.PubKey, key)
}

func (txi *InputCoin) OwnedBy(key []byte) bool {
	return bytes.Equal(txi.PubKey, key)
}

func NewCoinBaseTxn(toPubKey []byte) *Transaction {
	txi := InputCoin{
		TxnID:           []byte{},
		OutputCoinIndex: -1,
	}

	txo := OutputCoin{
		Amount: repository.CoinbaseReward,
		PubKey: toPubKey,
	}
	txn := &Transaction{
		InputCoins:  []*InputCoin{&txi},
		OutputCoins: []*OutputCoin{&txo},
	}
	txn.ID = ExtractTxnID(nil, time.Now())
	return txn
}

func CalculateTxnsHash(txns ...*Transaction) []byte {
	data := make([]interface{}, len(txns))
	for i := range txns {
		data[i] = txns[i].ID
	}

	return hasher.EasyHash(data...)
}

func ExtractTxnID(sin []byte, time time.Time) []byte {
	buf := bytes.NewBuffer(make([]byte, len(sin)+len(time.String())))
	_, _ = fmt.Fprint(buf, sin, time)
	return buf.Bytes()
}
