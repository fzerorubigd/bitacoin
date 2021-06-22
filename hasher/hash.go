package hasher

import (
	"context"
	"crypto/sha256"
	"fmt"
)

// GenerateMask creates a mask based on the number of zeros required in the hash
func GenerateMask(zeros int) []byte {
	full, half := zeros/2, zeros%2
	var mask []byte
	for i := 0; i < full; i++ {
		mask = append(mask, 0)
	}

	if half > 0 {
		mask = append(mask, 0xf)
	}

	return mask
}

// GoodEnough checks if the hash is good for the current mask
func GoodEnough(mask []byte, hash []byte) bool {
	for i := range mask {
		if hash[i] > mask[i] {
			return false
		}
	}
	return true
}

// EasyHash create hash, the easy way, just a simple sha256 hash
func EasyHash(data ...interface{}) []byte {
	hasher := sha256.New()

	fmt.Fprint(hasher, data...)

	return hasher.Sum(nil)
}

// DifficultHash creates the hash with difficulty mask and conditions,
// return the hash and the nonce used to create the hash
func DifficultHash(ctx context.Context, mask []byte, data ...interface{}) ([]byte, uint64) {
	ln := len(data)
	data = append(data, nil)
	var i uint64
	for {
		select {
		case <-ctx.Done():
			return nil, 0
		default:
			data[ln] = i
			hash := EasyHash(data...)
			if GoodEnough(mask, hash) {
				return hash, i
			}
			i++
		}
	}
}
