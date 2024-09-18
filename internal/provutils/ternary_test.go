package provutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO[2137]: func TestTernary(t *testing.T)
// TODO[2137]: func TestPluralize(t *testing.T)

func TestPluralEnding(t *testing.T) {
	tests := []struct {
		len int
		exp string
	}{
		{len: 0, exp: "s"},
		{len: 1, exp: ""},
		{len: 2, exp: "s"},
		{len: 3, exp: "s"},
		{len: 5, exp: "s"},
		{len: 50, exp: "s"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.len), func(t *testing.T) {
			actual := PluralEnding(make([]bool, tc.len))
			assert.Equal(t, tc.exp, actual, "pluralEnding(%d)", tc.len)
		})
	}
}
