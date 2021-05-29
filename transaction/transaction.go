package transaction

import (
	"github.com/fzerorubigd/bitacoin/hasher"
	"time"
)

import "bytes"

const (
	coinBaseReward = 100000
)

type Transaction struct {
	ID []byte

	VOut []TXOutput
	VIn  []TXInput
}

func (txn *Transaction) IsCoinBase() bool {
	return len(txn.VOut) == 1 &&
		len(txn.VIn) == 1 &&
		txn.VIn[0].VOut == -1 &&
		len(txn.VIn[0].TXID) == 0
}

type TXOutput struct {
	Value  int
	PubKey []byte
}

func (txo *TXOutput) TryUnlock(key []byte) bool {
	return bytes.Equal(txo.PubKey, key)
}

type TXInput struct {
	TXID []byte
	VOut int
	Sig  []byte
}

func (txi *TXInput) MatchLock(key []byte) bool {
	return bytes.Equal(txi.Sig, key)
}

func CalculateTxnID(txn *Transaction) []byte {
	return hasher.EasyHash(txn.VOut, txn.VIn)
}

func NewCoinBaseTxn(to, data []byte) *Transaction {
	if len(data) == 0 {
		data = hasher.EasyHash(to, time.Now())
	}

	txi := TXInput{
		TXID: []byte{},
		VOut: -1,
		Sig:  data,
	}

	txo := TXOutput{
		Value:  coinBaseReward,
		PubKey: to,
	}
	txn := &Transaction{
		VOut: []TXOutput{txo},
		VIn:  []TXInput{txi},
	}
	txn.ID = CalculateTxnID(txn)
	return txn
}

func CalculateTxnsHash(txns ...*Transaction) []byte {
	data := make([]interface{}, len(txns))
	for i := range txns {
		data[i] = txns[i].ID
	}

	return hasher.EasyHash(data...)
}
