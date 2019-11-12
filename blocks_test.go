package bitacoin

import (
	"strings"
	"testing"
)

func TestBlockCreation(t *testing.T) {
	const data = "Block Data"
	mask := GenerateMask(2)
	prev := EasyHash("Prev hash")

	b := NewBlock(data, mask, prev)
	if err := b.Validate(mask); err != nil {
		t.Errorf("Validation failed: %q", err)
		return
	}

	if !strings.Contains(b.String(), data) {
		t.Errorf("The data is not in string")
	}

	mask2 := GenerateMask(8)
	if err := b.Validate(mask2); err == nil {
		t.Errorf("Block should be invalid with mask %x, but is not", mask2)
	}

	// Invalidate block
	b.Nonce++
	if err := b.Validate(mask); err == nil {
		t.Errorf("Block should be invalid, but is not")
	}
	b.Nonce--

	b.Data = []byte("Something else")
	if err := b.Validate(mask); err == nil {
		t.Errorf("Block should be invalid, but is not")
	}
	b.Data = []byte(data)

	b.PrevHash = EasyHash("Something else")
	if err := b.Validate(mask); err == nil {
		t.Errorf("Block should be invalid, but is not")
	}
	b.PrevHash = prev

	hash := b.Hash
	b.Hash, _ = DifficultHash(mask, "Something else")
	if err := b.Validate(mask); err == nil {
		t.Errorf("Block should be invalid, but is not")
	}
	b.Hash = hash
}
