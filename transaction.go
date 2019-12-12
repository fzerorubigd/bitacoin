package bitacoin

import "time"

import "bytes"

import "fmt"

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

func NewTransaction(bc *BlockChain, from, to []byte, amount int) (*Transaction, error) {
	txns, txom, acc, err := bc.UnspentTxn(from)
	if err != nil {
		return nil, fmt.Errorf("get unused txn failed: %w", err)
	}

	if amount <= 0 {
		return nil, fmt.Errorf("negative transfer?")
	}

	if acc < amount {
		return nil, fmt.Errorf("not enough money, want %d have %d", amount, acc)
	}

	var (
		vin      []TXInput
		required = amount
	)

bigLoop:
	for id, txn := range txns {
		for _, v := range txom[id] {
			required -= txn.VOut[v].Value
			vin = append(vin, TXInput{
				TXID: txn.ID,
				VOut: v,
				Sig:  from, // TODO : real sign
			})

			if required <= 0 {
				break bigLoop
			}
		}
	}

	vout := []TXOutput{
		TXOutput{
			Value:  amount,
			PubKey: to,
		},
	}
	if required < 0 {
		vout = append(vout, TXOutput{
			Value:  -required,
			PubKey: from,
		})
	}

	txn := &Transaction{
		VIn:  vin,
		VOut: vout,
	}

	txn.ID = calculateTxnID(txn)

	return txn, nil
}
