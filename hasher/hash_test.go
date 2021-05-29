package hasher

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGenerateMask(t *testing.T) {
	data := []struct {
		In  int
		Out []byte
	}{
		{
			In:  0,
			Out: []byte{},
		},
		{
			In:  1,
			Out: []byte{0xf},
		},
		{
			In:  2,
			Out: []byte{0},
		},
		{
			In:  3,
			Out: []byte{0, 0xf},
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			t.Parallel()
			out := GenerateMask(d.In)
			if !bytes.Equal(out, d.Out) {
				t.Errorf("Failed for %d, it should be %x but is %x", d.In, d.Out, out)
			}
		})
	}
}

func TestDifficultHash(t *testing.T) {
	mask := GenerateMask(2)
	hash, nonce := DifficultHash(mask, "a", "b", []byte("abc"))
	if !GoodEnough(mask, hash) {
		t.Errorf("Hash is not compatibe with mask")
	}

	easy := EasyHash("a", "b", []byte("abc"), nonce)
	if !bytes.Equal(easy, hash) {
		t.Error("Hash is not valid")
	}
}

func ExampleGenerateMask() {
	mask := GenerateMask(3)
	fmt.Printf("%x", mask)

	// Output:
	// 000f
}

func BenchmarkDifficultHash(b *testing.B) {
	mask := GenerateMask(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DifficultHash(mask, "a", "b", i)
	}

	b.ReportAllocs()
}
