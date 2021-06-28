package blockchain

import (
	"context"
	"encoding/hex"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/hasher"
	"github.com/fzerorubigd/bitacoin/transaction"
	"strings"
	"testing"
)

func TestBlockCreation(t *testing.T) {
	data := []*transaction.Transaction{transaction.NewRewardTxn([]byte("bita"))}
	mask := hasher.GenerateMask(2)
	prev := hasher.EasyHash("Prev hash")

	b := block.StartMining(context.Background(), data, mask, prev)
	if err := b.Validate(mask); err != nil {
		t.Errorf("Validation failed: %q", err)
		return
	}

	if !strings.Contains(b.String(), hex.EncodeToString(prev)) {
		t.Errorf("The prev hash is not in string")
	}

	mask2 := hasher.GenerateMask(8)
	if err := b.Validate(mask2); err == nil {
		t.Errorf("Block should be invalid with mask %x, but is not", mask2)
	}

	// Invalidate block
	b.Nonce++
	if err := b.Validate(mask); err == nil {
		t.Errorf("Block should be invalid, but is not")
	}
	b.Nonce--

	b.Transactions = []*transaction.Transaction{transaction.NewRewardTxn([]byte("forud"))}
	if err := b.Validate(mask); err == nil {
		t.Errorf("Block should be invalid, but is not")
	}
	b.Transactions = data

	b.PrevHash = hasher.EasyHash("Something else")
	if err := b.Validate(mask); err == nil {
		t.Errorf("Block should be invalid, but is not")
	}
	b.PrevHash = prev

	hash := b.Hash
	b.Hash, _ = hasher.DifficultHash(context.Background(), mask, "Something else")
	if err := b.Validate(mask); err == nil {
		t.Errorf("Block should be invalid, but is not")
	}
	b.Hash = hash
}
