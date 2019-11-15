package bitacoin

import "time"

const (
	coinBaseReward = 100000
)

type Transaction struct {
	ID []byte

	VOut []TXOutput
	VIn  []TXInput
}

type TXOutput struct {
	Value  int
	PubKey []byte
}

type TXInput struct {
	TXID []byte
	VOut int
	Sig  []byte
}

func calculateTxnID(txn *Transaction) []byte {
	return EasyHash(txn.VOut, txn.VIn)
}

func calculateTxnsHash(txns ...*Transaction) []byte {
	data := make([]interface{}, len(txns))
	for i := range txns {
		data[i] = txns[i].ID
	}

	return EasyHash(data...)
}

func NewCoinBaseTxn(to, data []byte) *Transaction {
	if len(data) == 0 {
		data = EasyHash(to, time.Now())
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
	txn.ID = calculateTxnID(txn)
	return txn
}
