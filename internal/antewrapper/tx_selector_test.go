package antewrapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewProvTxSelector(t *testing.T) {
	expC := &provTxSelector{}
	var act baseapp.TxSelector
	testFunc := func() {
		act = NewProvTxSelector()
	}
	require.NotPanics(t, testFunc, "NewProvTxSelector()")
	assertNotNilInterface(t, act, "NewProvTxSelector() result")
	actC, ok := act.(*provTxSelector)
	assert.True(t, ok, "is NewProvTxSelector() result a *provTxSelector")
	assert.Equal(t, expC, actC, "NewProvTxSelector() result underlying *provTxSelector")
}

func TestProvTxSelector_SelectedTxs(t *testing.T) {
	newSelector := func(selectedTxs ...string) *provTxSelector {
		rv := &provTxSelector{}
		// Give it extra capacity so that we can test that a copy is returned.
		rv.selectedTxs = make([][]byte, len(selectedTxs), len(selectedTxs)+10)
		for i, tx := range selectedTxs {
			rv.selectedTxs[i] = []byte(tx)
		}
		return rv
	}

	tests := []struct {
		name string
		ts   *provTxSelector
		exp  [][]byte
	}{
		{
			name: "nil selected",
			ts:   &provTxSelector{selectedTxs: nil},
			exp:  make([][]byte, 0),
		},
		{
			name: "one selected",
			ts:   newSelector("dummyentrytxbz"),
		},
		{
			name: "three selected",
			ts:   newSelector("dummyentrytxonebz", "dummyentrytxtwobz", "dummyentrytxthreebz"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.exp == nil {
				tc.exp = tc.ts.selectedTxs
			}

			var act [][]byte
			testFunc := func() {
				act = tc.ts.SelectedTxs(nil)
			}
			require.NotPanics(t, testFunc, "SelectedTxs")
			assert.Equal(t, tc.exp, act, "SelectedTxs result")

			// Adding an entry in the returned slice must not affect the original.
			act = append(act, []byte("appendedlater"))
			assert.NotEqual(t, tc.ts.selectedTxs, act, "SelectedTxs result after changing a byte")
		})
	}
}

func TestProvTxSelector_Clear(t *testing.T) {
	selector := &provTxSelector{
		totalTxBytes: 23,
		totalTxGas:   3883,
		selectedTxs:  [][]byte{[]byte("txzero"), []byte("txone"), []byte("txtwo"), []byte("txthree")},
	}
	testFunc := func() {
		selector.Clear()
	}
	require.NotPanics(t, testFunc, "Clear()")
	assert.Equal(t, 0, int(selector.totalTxBytes), "totalTxBytes")
	assert.Equal(t, 0, int(selector.totalTxGas), "totalTxGas")
	assert.Empty(t, selector.selectedTxs, "selectedTxs")
}

func TestProvTxSelector_SelectTxForProposal(t *testing.T) {
	tests := []struct {
		name        string
		ts          *provTxSelector
		maxTxBytes  uint64
		maxBlockGas uint64
		memTx       sdk.Tx
		txBz        []byte
		exp         bool
		expTS       *provTxSelector
	}{
		{
			name:        "tx puts us over the tx bytes limit",
			ts:          &provTxSelector{totalTxBytes: 53},
			maxTxBytes:  59,
			maxBlockGas: 0,
			memTx:       nil,
			txBz:        []byte("toolong"),
			exp:         false,
			expTS:       &provTxSelector{totalTxBytes: 53},
		},
		{
			name:        "tx puts us exactly at the tx bytes limit",
			ts:          &provTxSelector{totalTxBytes: 53},
			maxTxBytes:  60,
			maxBlockGas: 0,
			memTx:       nil,
			txBz:        []byte("toolong"),
			exp:         true,
			expTS:       &provTxSelector{totalTxBytes: 60, selectedTxs: [][]byte{[]byte("toolong")}},
		},
		{
			name:        "tx fits with byte space left over",
			ts:          &provTxSelector{totalTxBytes: 53},
			maxTxBytes:  61,
			maxBlockGas: 0,
			memTx:       nil,
			txBz:        []byte("toolong"),
			exp:         false,
			expTS:       &provTxSelector{totalTxBytes: 60, selectedTxs: [][]byte{[]byte("toolong")}},
		},
		{
			name:        "tx gas puts us over the block gas limit",
			ts:          &provTxSelector{totalTxGas: 100},
			maxTxBytes:  5000,
			maxBlockGas: 120,
			memTx:       NewMockFeeTx("").WithGas(21),
			txBz:        []byte("randomstr"),
			exp:         false,
			expTS:       &provTxSelector{totalTxGas: 100},
		},
		{
			name:        "tx gas puts us exactly at the block gas limit",
			ts:          &provTxSelector{totalTxGas: 100},
			maxTxBytes:  5000,
			maxBlockGas: 121,
			memTx:       NewMockFeeTx("").WithGas(21),
			txBz:        []byte("randomstr"),
			exp:         true,
			expTS:       &provTxSelector{totalTxGas: 121, totalTxBytes: 9, selectedTxs: [][]byte{[]byte("randomstr")}},
		},
		{
			name:        "tx fits with gas space left over",
			ts:          &provTxSelector{totalTxGas: 100},
			maxTxBytes:  5000,
			maxBlockGas: 122,
			memTx:       NewMockFeeTx("").WithGas(21),
			txBz:        []byte("randomstr"),
			exp:         false,
			expTS:       &provTxSelector{totalTxGas: 121, totalTxBytes: 9, selectedTxs: [][]byte{[]byte("randomstr")}},
		},
		{
			name:        "tx added while max block gas zero",
			ts:          &provTxSelector{totalTxGas: 5},
			maxTxBytes:  5000,
			maxBlockGas: 0,
			memTx:       NewMockFeeTx("").WithGas(300),
			txBz:        []byte("danceparty"),
			exp:         false,
			expTS:       &provTxSelector{totalTxGas: 5, totalTxBytes: 10, selectedTxs: [][]byte{[]byte("danceparty")}},
		},
		{
			name:        "memTx not a fee tx",
			ts:          &provTxSelector{totalTxGas: 3, totalTxBytes: 14},
			maxTxBytes:  5000,
			maxBlockGas: 300,
			memTx:       NewNotFeeTx(""),
			txBz:        []byte("weirdmemtxhuh"),
			exp:         false,
			expTS:       &provTxSelector{totalTxGas: 3, totalTxBytes: 27, selectedTxs: [][]byte{[]byte("weirdmemtxhuh")}},
		},
		{
			name:        "memTx has same gas as fee",
			ts:          &provTxSelector{},
			maxTxBytes:  5000,
			maxBlockGas: 60_000_000,
			memTx:       NewMockFeeTx("").WithGas(350_123).WithFee(sdk.NewCoins(sdk.NewInt64Coin("nhash", 350_123))),
			txBz:        []byte("justanothertx"),
			exp:         false,
			expTS:       &provTxSelector{totalTxGas: DefaultGasLimit, totalTxBytes: 13, selectedTxs: [][]byte{[]byte("justanothertx")}},
		},
		{
			name:        "memTx has other gas",
			ts:          &provTxSelector{},
			maxTxBytes:  5000,
			maxBlockGas: 60_000_000,
			memTx:       NewMockFeeTx("").WithGas(350_124).WithFee(sdk.NewCoins(sdk.NewInt64Coin("nhash", 350_123))),
			txBz:        []byte("justanothertx"),
			exp:         false,
			expTS:       &provTxSelector{totalTxGas: 350_124, totalTxBytes: 13, selectedTxs: [][]byte{[]byte("justanothertx")}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := sdk.Context{}.WithLogger(log.NewNopLogger())
			var act bool
			testFunc := func() {
				act = tc.ts.SelectTxForProposal(ctx, tc.maxTxBytes, tc.maxBlockGas, tc.memTx, tc.txBz)
			}
			require.NotPanics(t, testFunc, "SelectTxForProposal")
			assert.Equal(t, tc.exp, act, "SelectTxForProposal result")
			AssertEqualGas(t, tc.expTS.totalTxGas, tc.ts.totalTxGas, "totalTxGas")
			AssertEqualGas(t, tc.expTS.totalTxBytes, tc.ts.totalTxBytes, "totalTxBytes")
			assert.Equal(t, tc.expTS.selectedTxs, tc.ts.selectedTxs, "selectedTxs")
		})
	}
}
