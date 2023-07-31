package hold

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDefaultGenesisState(t *testing.T) {
	genState := DefaultGenesisState()
	require.NotNil(t, genState, "DefaultGenesisState()")
	assert.Empty(t, genState.Escrows, "Escrows")
}

func TestGenesisState_Validate(t *testing.T) {
	escrows := func(rv ...*AccountEscrow) []*AccountEscrow {
		return rv
	}

	aeGood1 := &AccountEscrow{
		Address: sdk.AccAddress("aeGood1_____________").String(),
		Amount:  sdk.NewCoins(sdk.NewInt64Coin("nhash", 5_000_000_001)),
	}
	aeGood2 := &AccountEscrow{
		Address: sdk.AccAddress("aeGood2_____________").String(),
		Amount:  sdk.NewCoins(sdk.NewInt64Coin("nhash", 35_000), sdk.NewInt64Coin("steak", 88)),
	}
	aeGood3 := &AccountEscrow{
		Address: sdk.AccAddress("aeGood3_____________").String(),
		Amount:  sdk.NewCoins(sdk.NewInt64Coin("nhash", 1_234_567_890)),
	}
	aeBad := &AccountEscrow{
		Address: "",
		Amount:  sdk.NewCoins(sdk.NewInt64Coin("acorn", 321)),
	}

	badErr := func(i int) string {
		return fmt.Sprintf("invalid escrows[%d]: invalid address: empty address string is not allowed", i)
	}
	nilErr := func(i int) string {
		return fmt.Sprintf("invalid escrows[%d]: cannot be nil", i)
	}
	dupErr := func(i, j int) string {
		return fmt.Sprintf("invalid escrows[%d]: duplicate address also at index %d", i, j)
	}

	tests := []struct {
		name     string
		genState GenesisState
		expErr   []string
	}{
		{
			name:     "nil escrows",
			genState: GenesisState{},
		},
		{
			name:     "empty escrows",
			genState: GenesisState{Escrows: nil},
		},
		{
			name:     "one escrow: good",
			genState: GenesisState{Escrows: escrows(aeGood1)},
		},
		{
			name:     "one escrow: bad",
			genState: GenesisState{Escrows: escrows(aeBad)},
			expErr:   []string{badErr(0)},
		},
		{
			name:     "one escrow: nil",
			genState: GenesisState{Escrows: escrows(nil)},
			expErr:   []string{nilErr(0)},
		},
		{
			name:     "two escrows: good good",
			genState: GenesisState{Escrows: escrows(aeGood2, aeGood3)},
		},
		{
			name:     "two escrows: good bad",
			genState: GenesisState{Escrows: escrows(aeGood1, aeBad)},
			expErr:   []string{badErr(1)},
		},
		{
			name:     "two escrows: bad good",
			genState: GenesisState{Escrows: escrows(aeBad, aeGood1)},
			expErr:   []string{badErr(0)},
		},
		{
			name:     "two escrows: good nil",
			genState: GenesisState{Escrows: escrows(aeGood1, nil)},
			expErr:   []string{nilErr(1)},
		},
		{
			name:     "two escrows: nil good",
			genState: GenesisState{Escrows: escrows(nil, aeGood1)},
			expErr:   []string{nilErr(0)},
		},
		{
			name:     "two escrows: bad nil",
			genState: GenesisState{Escrows: escrows(aeBad, nil)},
			expErr:   []string{badErr(0), nilErr(1)},
		},
		{
			name:     "two escrows: nil bad",
			genState: GenesisState{Escrows: escrows(nil, aeBad)},
			expErr:   []string{nilErr(0), badErr(1)},
		},
		{
			name:     "two escrows: same",
			genState: GenesisState{Escrows: escrows(aeGood1, aeGood1)},
			expErr:   []string{dupErr(1, 0)},
		},
		{
			name:     "three escrows: good good good",
			genState: GenesisState{Escrows: escrows(aeGood1, aeGood2, aeGood3)},
		},
		{
			name:     "three escrows: good bad good",
			genState: GenesisState{Escrows: escrows(aeGood1, aeBad, aeGood3)},
			expErr:   []string{badErr(1)},
		},
		{
			name:     "three escrows: good nil bad",
			genState: GenesisState{Escrows: escrows(aeGood1, nil, aeBad)},
			expErr:   []string{nilErr(1), badErr(2)},
		},
		{
			name:     "three escrows: same first and third bad second",
			genState: GenesisState{Escrows: escrows(aeGood1, aeBad, aeGood1)},
			expErr:   []string{badErr(1), dupErr(2, 0)},
		},
		{
			name:     "three escrows: all same",
			genState: GenesisState{Escrows: escrows(aeGood1, aeGood1, aeGood1)},
			expErr:   []string{dupErr(1, 0), dupErr(2, 0)},
		},
		{
			name:     "six escrows: good1 bad good2 nil good1 good3",
			genState: GenesisState{Escrows: escrows(aeGood1, aeBad, aeGood2, nil, aeGood2, aeGood3)},
			expErr:   []string{badErr(1), nilErr(3), dupErr(4, 2)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.genState.Validate()
			}
			require.NotPanics(t, testFunc, "Validate()")
			if len(tc.expErr) > 0 {
				if assert.Error(t, err, "Validate()") {
					for _, exp := range tc.expErr {
						assert.ErrorContains(t, err, exp, "Validate()\nExpected substring: %q", exp)
					}
				}
			} else {
				assert.NoError(t, err, "Validate()")
			}
		})
	}
}
