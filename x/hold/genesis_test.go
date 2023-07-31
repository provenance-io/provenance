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
	assert.Empty(t, genState.Holds, "Holds")
}

func TestGenesisState_Validate(t *testing.T) {
	holds := func(rv ...*AccountHold) []*AccountHold {
		return rv
	}

	ahGood1 := &AccountHold{
		Address: sdk.AccAddress("ahGood1_____________").String(),
		Amount:  sdk.NewCoins(sdk.NewInt64Coin("nhash", 5_000_000_001)),
	}
	ahGood2 := &AccountHold{
		Address: sdk.AccAddress("ahGood2_____________").String(),
		Amount:  sdk.NewCoins(sdk.NewInt64Coin("nhash", 35_000), sdk.NewInt64Coin("steak", 88)),
	}
	ahGood3 := &AccountHold{
		Address: sdk.AccAddress("ahGood3_____________").String(),
		Amount:  sdk.NewCoins(sdk.NewInt64Coin("nhash", 1_234_567_890)),
	}
	ahBad := &AccountHold{
		Address: "",
		Amount:  sdk.NewCoins(sdk.NewInt64Coin("acorn", 321)),
	}

	badErr := func(i int) string {
		return fmt.Sprintf("invalid holds[%d]: invalid address: empty address string is not allowed", i)
	}
	nilErr := func(i int) string {
		return fmt.Sprintf("invalid holds[%d]: cannot be nil", i)
	}
	dupErr := func(i, j int) string {
		return fmt.Sprintf("invalid holds[%d]: duplicate address also at index %d", i, j)
	}

	tests := []struct {
		name     string
		genState GenesisState
		expErr   []string
	}{
		{
			name:     "nil holds",
			genState: GenesisState{Holds: nil},
		},
		{
			name:     "empty holds",
			genState: GenesisState{Holds: []*AccountHold{}},
		},
		{
			name:     "one holds: good",
			genState: GenesisState{Holds: holds(ahGood1)},
		},
		{
			name:     "one hold: bad",
			genState: GenesisState{Holds: holds(ahBad)},
			expErr:   []string{badErr(0)},
		},
		{
			name:     "one hold: nil",
			genState: GenesisState{Holds: holds(nil)},
			expErr:   []string{nilErr(0)},
		},
		{
			name:     "two holds: good good",
			genState: GenesisState{Holds: holds(ahGood2, ahGood3)},
		},
		{
			name:     "two holds: good bad",
			genState: GenesisState{Holds: holds(ahGood1, ahBad)},
			expErr:   []string{badErr(1)},
		},
		{
			name:     "two holds: bad good",
			genState: GenesisState{Holds: holds(ahBad, ahGood1)},
			expErr:   []string{badErr(0)},
		},
		{
			name:     "two holds: good nil",
			genState: GenesisState{Holds: holds(ahGood1, nil)},
			expErr:   []string{nilErr(1)},
		},
		{
			name:     "two holds: nil good",
			genState: GenesisState{Holds: holds(nil, ahGood1)},
			expErr:   []string{nilErr(0)},
		},
		{
			name:     "two holds: bad nil",
			genState: GenesisState{Holds: holds(ahBad, nil)},
			expErr:   []string{badErr(0), nilErr(1)},
		},
		{
			name:     "two holds: nil bad",
			genState: GenesisState{Holds: holds(nil, ahBad)},
			expErr:   []string{nilErr(0), badErr(1)},
		},
		{
			name:     "two holds: same",
			genState: GenesisState{Holds: holds(ahGood1, ahGood1)},
			expErr:   []string{dupErr(1, 0)},
		},
		{
			name:     "three holds: good good good",
			genState: GenesisState{Holds: holds(ahGood1, ahGood2, ahGood3)},
		},
		{
			name:     "three holds: good bad good",
			genState: GenesisState{Holds: holds(ahGood1, ahBad, ahGood3)},
			expErr:   []string{badErr(1)},
		},
		{
			name:     "three holds: good nil bad",
			genState: GenesisState{Holds: holds(ahGood1, nil, ahBad)},
			expErr:   []string{nilErr(1), badErr(2)},
		},
		{
			name:     "three holds: same first and third bad second",
			genState: GenesisState{Holds: holds(ahGood1, ahBad, ahGood1)},
			expErr:   []string{badErr(1), dupErr(2, 0)},
		},
		{
			name:     "three holds: all same",
			genState: GenesisState{Holds: holds(ahGood1, ahGood1, ahGood1)},
			expErr:   []string{dupErr(1, 0), dupErr(2, 0)},
		},
		{
			name:     "six holds: good1 bad good2 nil good1 good3",
			genState: GenesisState{Holds: holds(ahGood1, ahBad, ahGood2, nil, ahGood2, ahGood3)},
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
