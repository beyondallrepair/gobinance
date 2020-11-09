package gobinance_test

import (
	"github.com/google/go-cmp/cmp"
	"math/big"
	"testing"
)

var bigFloatComparer = cmp.Comparer(func(a, b *big.Float) bool {
	return a.Cmp(b) == 0
})

func mustParseBigFloat(t *testing.T, str string) *big.Float {
	v, _, err := new(big.Float).Parse(str, 10)
	if err != nil {
		t.Fatalf("unable to parse big float %v: %v", str, err)
	}
	return v
}
