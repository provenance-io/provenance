package provutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTernary(t *testing.T) {
	tests := []struct {
		test    bool
		ifTrue  int
		ifFalse int
		exp     int
	}{
		{test: true, ifTrue: 0, ifFalse: 0, exp: 0},
		{test: true, ifTrue: 0, ifFalse: 1, exp: 0},
		{test: true, ifTrue: 1, ifFalse: 0, exp: 1},
		{test: true, ifTrue: -10, ifFalse: 20, exp: -10},
		{test: true, ifTrue: 100, ifFalse: -200, exp: 100},
		{test: false, ifTrue: 0, ifFalse: 0, exp: 0},
		{test: false, ifTrue: 0, ifFalse: 1, exp: 1},
		{test: false, ifTrue: 1, ifFalse: 0, exp: 0},
		{test: false, ifTrue: -10, ifFalse: 20, exp: 20},
		{test: false, ifTrue: 100, ifFalse: -200, exp: -200},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("%t, %d, %d", tc.test, tc.ifTrue, tc.ifFalse)
		t.Run(name, func(t *testing.T) {
			var actual int
			testFunc := func() {
				actual = Ternary(tc.test, tc.ifTrue, tc.ifFalse)
			}
			require.NotPanics(t, testFunc, "Ternary(%s)", name)
			assert.Equal(t, tc.exp, actual, "result of Ternary(%s)", name)
		})
	}
}

func TestPluralize(t *testing.T) {
	ifOne := "ifOne"
	ifOther := "ifOther"
	strs := []string{
		"one", "two", "three", "four", "five",
		"six", "seven", "eight", "nine", "ten",
	}

	tests := []struct {
		name string
		s    []string
		exp  string
	}{
		{
			name: "nil slice",
			s:    nil,
			exp:  ifOther,
		},
		{
			name: "empty slice",
			s:    make([]string, 0),
			exp:  ifOther,
		},
		{
			name: "one entry",
			s:    strs[0:1],
			exp:  ifOne,
		},
		{
			name: "two entries",
			s:    strs[1:3],
			exp:  ifOther,
		},
		{
			name: "three entries",
			s:    strs[3:6],
			exp:  ifOther,
		},
		{
			name: "ten entries",
			s:    strs[0:10],
			exp:  ifOther,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = Pluralize(tc.s, ifOne, ifOther)
			}
			require.NotPanics(t, testFunc, "Pluralize(%d, %q, %q)", len(tc.s), ifOne, ifOther)
			assert.Equal(t, tc.exp, actual, "result of Pluralize(%d, %q, %q)", len(tc.s), ifOne, ifOther)
		})
	}
}

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
			vals := make([]bool, tc.len)
			var actual string
			testFunc := func() {
				actual = PluralEnding(vals)
			}
			require.NotPanics(t, testFunc, "PluralEnding(%d)", tc.len)
			assert.Equal(t, tc.exp, actual, "result of PluralEnding(%d)", tc.len)
		})
	}
}
