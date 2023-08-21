package simulation_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/hold"
	"github.com/provenance-io/provenance/x/hold/simulation"
)

// startStdoutCapture hijacks os.Stdout and starts capturing it instead.
// Once the returned function is called, capturing will end and the captured
// output will be in the provided buffer.
//
// The returned function is a closer, and MUST be called.
// Calling it multiple times is okay.
func startStdoutCapture(t *testing.T, buffer *bytes.Buffer) func() error {
	pipeReader, pipeWriter, err := os.Pipe()
	require.NoError(t, err, "creating os.Pipe")
	origStdout := os.Stdout
	os.Stdout = pipeWriter
	// Separate go routine for copying so we can return from here without blocking.
	outChan := make(chan bytes.Buffer)
	go func() {
		var readBuffer bytes.Buffer
		_, _ = io.Copy(&readBuffer, pipeReader)
		outChan <- readBuffer
	}()
	var closeErr error
	closed := false
	return func() error {
		if !closed {
			closed = true
			os.Stdout = origStdout
			werr := pipeWriter.Close()
			result := <-outChan
			_, cerr := io.Copy(buffer, &result)
			closeErr = errors.Join(werr, cerr)
		}
		return closeErr
	}
}

func TestRandomAccountHolds(t *testing.T) {
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 3)

	tests := []struct {
		name     string
		seed     int64
		accounts []simtypes.Account
		expected []*hold.AccountHold
	}{
		{
			name:     "nil accounts",
			seed:     0,
			accounts: nil,
			expected: nil,
		},
		{
			name:     "zero accounts",
			seed:     0,
			accounts: []simtypes.Account{},
			expected: nil,
		},
		{
			name:     "1 account not picked",
			seed:     0,
			accounts: accs[0:1],
			expected: nil,
		},
		{
			name:     "1 account picked",
			seed:     1,
			accounts: accs[0:1],
			expected: []*hold.AccountHold{
				{
					Address: accs[0].Address.String(),
					Amount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 552)),
				},
			},
		},
		{
			name:     "2 accounts neither picked",
			seed:     0,
			accounts: accs[0:2],
			expected: nil,
		},
		{
			name:     "2 accounts one picked",
			seed:     2,
			accounts: accs[0:2],
			expected: []*hold.AccountHold{
				{
					Address: accs[1].Address.String(),
					Amount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 543)),
				},
			},
		},
		{
			name:     "2 accounts both picked",
			seed:     1,
			accounts: accs[0:2],
			expected: []*hold.AccountHold{
				{
					Address: accs[0].Address.String(),
					Amount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 822)),
				},
				{
					Address: accs[1].Address.String(),
					Amount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 52)),
				},
			},
		},
		{
			name:     "3 accounts none picked",
			seed:     3,
			accounts: accs,
			expected: nil,
		},
		{
			name:     "3 accounts 2 picked",
			seed:     0,
			accounts: accs,
			expected: []*hold.AccountHold{
				{
					Address: accs[2].Address.String(),
					Amount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 795)),
				},
				{
					Address: accs[1].Address.String(),
					Amount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 203)),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(tc.seed))
			holds := simulation.RandomAccountHolds(r, tc.accounts)
			assert.Equal(t, tc.expected, holds, "RandomAccountHolds result")
		})
	}
}

func TestUpdateBankGenStateForHolds(t *testing.T) {
	denom := "trout"
	coins := func(amt int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom, amt))
	}
	holdGenState := func(holds ...*hold.AccountHold) *hold.GenesisState {
		return &hold.GenesisState{Holds: holds}
	}
	ah := func(addr sdk.AccAddress, amt int64) *hold.AccountHold {
		return &hold.AccountHold{
			Address: addr.String(),
			Amount:  coins(amt),
		}
	}
	bankGenState := func(balances ...banktypes.Balance) *banktypes.GenesisState {
		rv := banktypes.DefaultGenesisState()
		rv.Balances = append(rv.Balances, balances...)
		for _, balance := range rv.Balances {
			rv.Supply = rv.Supply.Add(balance.Coins...)
		}
		return rv
	}
	bal := func(addr sdk.AccAddress, amt int64) banktypes.Balance {
		return banktypes.Balance{
			Address: addr.String(),
			Coins:   coins(amt),
		}
	}
	holdsStrings := func(holdGen *hold.GenesisState) []string {
		if holdGen == nil {
			return nil
		}
		rv := make([]string, len(holdGen.Holds))
		for i, h := range holdGen.Holds {
			// going back to raw address bytes because that'll be easier to identify in failure output.
			rv[i] = fmt.Sprintf("%q:%q", string(sdk.MustAccAddressFromBech32(h.Address)), h.Amount)
		}
		return rv
	}
	balsStrings := func(bankGen *banktypes.GenesisState) []string {
		if bankGen == nil {
			return nil
		}
		rv := make([]string, len(bankGen.Balances))
		for i, b := range bankGen.Balances {
			// going back to raw address bytes because that'll be easier to identify in failure output.
			rv[i] = fmt.Sprintf("%q:%q", string(sdk.MustAccAddressFromBech32(b.Address)), b.Coins)
		}
		return rv
	}
	panicNoBal := func(h *hold.AccountHold) string {
		return fmt.Sprintf("no bank genesis balance found for %s that should have a hold on %s", h.Address, h.Amount)
	}

	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")
	addr3 := sdk.AccAddress("addr3_______________")
	addr4 := sdk.AccAddress("addr4_______________")

	tests := []struct {
		name       string
		bankGen    *banktypes.GenesisState
		holdGen    *hold.GenesisState
		expBankGen *banktypes.GenesisState
		expPanic   string
	}{
		{
			name:       "no holds",
			bankGen:    bankGenState(),
			holdGen:    holdGenState(),
			expBankGen: bankGenState(),
		},
		{
			name:       "one balance no holds",
			bankGen:    bankGenState(bal(addr1, 33)),
			holdGen:    holdGenState(),
			expBankGen: bankGenState(bal(addr1, 33)),
		},
		{
			name:       "one balance with a hold",
			bankGen:    bankGenState(bal(addr1, 33)),
			holdGen:    holdGenState(ah(addr1, 17)),
			expBankGen: bankGenState(bal(addr1, 50)),
		},
		{
			name:       "three balances one hold first",
			bankGen:    bankGenState(bal(addr1, 33), bal(addr2, 66), bal(addr3, 99)),
			holdGen:    holdGenState(ah(addr1, 17)),
			expBankGen: bankGenState(bal(addr1, 50), bal(addr2, 66), bal(addr3, 99)),
		},
		{
			name:       "three balances one hold second",
			bankGen:    bankGenState(bal(addr1, 33), bal(addr2, 66), bal(addr3, 99)),
			holdGen:    holdGenState(ah(addr2, 24)),
			expBankGen: bankGenState(bal(addr1, 33), bal(addr2, 90), bal(addr3, 99)),
		},
		{
			name:       "three balances one hold third",
			bankGen:    bankGenState(bal(addr1, 33), bal(addr2, 66), bal(addr3, 99)),
			holdGen:    holdGenState(ah(addr3, 1)),
			expBankGen: bankGenState(bal(addr1, 33), bal(addr2, 66), bal(addr3, 100)),
		},
		{
			name:       "three balances three holds",
			bankGen:    bankGenState(bal(addr1, 33), bal(addr2, 66), bal(addr3, 99)),
			holdGen:    holdGenState(ah(addr2, 24), ah(addr3, 1), ah(addr1, 17)),
			expBankGen: bankGenState(bal(addr1, 50), bal(addr2, 90), bal(addr3, 100)),
		},
		{
			name:     "no balances one hold",
			bankGen:  bankGenState(),
			holdGen:  holdGenState(ah(addr1, 17)),
			expPanic: panicNoBal(ah(addr1, 17)),
		},
		{
			name:     "three balances one hold not in balances",
			bankGen:  bankGenState(bal(addr1, 33), bal(addr2, 66), bal(addr3, 99)),
			holdGen:  holdGenState(ah(addr4, 111)),
			expPanic: panicNoBal(ah(addr4, 111)),
		},
		{
			name:     "three balances four holds",
			bankGen:  bankGenState(bal(addr1, 33), bal(addr2, 66), bal(addr3, 99)),
			holdGen:  holdGenState(ah(addr2, 24), ah(addr3, 1), ah(addr4, 111), ah(addr1, 17)),
			expPanic: panicNoBal(ah(addr4, 111)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Using sting slices to compare these because it makes failures easier to read.
			expectedHolds := holdsStrings(tc.holdGen)
			expectedBals := balsStrings(tc.expBankGen)

			testFunc := func() {
				simulation.UpdateBankGenStateForHolds(tc.bankGen, tc.holdGen)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "UpdateBankGenStateForHolds")
			if tc.expBankGen != nil {
				actualBals := balsStrings(tc.bankGen)
				assert.Equal(t, expectedBals, actualBals, "resulting bank genesis balances")
				assert.Equal(t, tc.expBankGen.Supply.String(), tc.bankGen.Supply.String(), "resulting bank genesis supply")
			}

			actualHolds := holdsStrings(tc.holdGen)
			assert.Equal(t, expectedHolds, actualHolds, "resulting hold genesis holds")
		})
	}
}

func TestHoldsString(t *testing.T) {
	denom := "carp"
	ah := func(addr sdk.AccAddress, amt int64) *hold.AccountHold {
		return &hold.AccountHold{
			Address: addr.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin(denom, amt)),
		}
	}

	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")
	addr3 := sdk.AccAddress("addr3_______________")
	addr4 := sdk.AccAddress("addr4_______________")

	tests := []struct {
		name  string
		holds []*hold.AccountHold
	}{
		{name: "nil holds", holds: nil},
		{name: "one hold", holds: []*hold.AccountHold{ah(addr1, 11)}},
		{name: "two holds", holds: []*hold.AccountHold{ah(addr1, 11), ah(addr2, 22)}},
		{name: "two holds reversed", holds: []*hold.AccountHold{ah(addr2, 22), ah(addr1, 11)}},
		{
			name:  "four holds",
			holds: []*hold.AccountHold{ah(addr1, 11), ah(addr2, 22), ah(addr3, 33), ah(addr4, 44)},
		},
		{
			name:  "four shuffled holds",
			holds: []*hold.AccountHold{ah(addr3, 33), ah(addr1, 11), ah(addr4, 44), ah(addr2, 22)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := simulation.HoldsString(tc.holds)
			if len(tc.holds) == 0 {
				assert.Equal(t, "{}", result, "HoldsString result")
			} else {
				expected := make([]string, len(tc.holds)+2)
				expected[0] = "{"
				for i, h := range tc.holds {
					expected[i+1] = ` "` + h.Address + `":"` + h.Amount.String() + `"`
					if i < len(tc.holds)-1 {
						expected[i+1] = expected[i+1] + ","
					}
				}
				expected[len(expected)-1] = "}"

				actual := strings.Split(result, "\n")
				assert.Equal(t, expected, actual, "HoldsString result")
			}
		})
	}
}

func TestRandomizedGenState(t *testing.T) {
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 3)

	bondCoins := func(amount int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, amount))
	}
	holdGen := func(holds ...*hold.AccountHold) *hold.GenesisState {
		rv := hold.DefaultGenesisState()
		rv.Holds = make([]*hold.AccountHold, len(holds))
		copy(rv.Holds, holds)
		return rv
	}
	ah := func(acc simtypes.Account, amount int64) *hold.AccountHold {
		return &hold.AccountHold{
			Address: acc.Address.String(),
			Amount:  bondCoins(amount),
		}
	}
	bankGen := func(bals ...banktypes.Balance) *banktypes.GenesisState {
		rv := banktypes.DefaultGenesisState()
		rv.Balances = bals
		for _, bal := range rv.Balances {
			rv.Supply = rv.Supply.Add(bal.Coins...)
		}
		return rv
	}
	bal := func(acc simtypes.Account, amount int64) banktypes.Balance {
		return banktypes.Balance{Address: acc.Address.String(), Coins: bondCoins(amount)}
	}
	panicNoBal := func(h *hold.AccountHold) string {
		return fmt.Sprintf("no bank genesis balance found for %s that should have a hold on %s", h.Address, h.Amount)
	}

	// Note: The random expected hold genesis state values are discovered using the TestRandomAccountHolds test.
	// If the way we use the generator changes, those values might change. If that happens, it's okay to change the
	// expected values to what the test is giving now. But only if you're expecting them to change.
	tests := []struct {
		name       string
		seed       int64
		appParams  []*hold.AccountHold
		accounts   []simtypes.Account
		bankGen    *banktypes.GenesisState
		expHoldGen *hold.GenesisState
		expBankGen *banktypes.GenesisState
		expPanic   string
	}{
		{
			name:       "no accounts",
			seed:       0,
			accounts:   nil,
			bankGen:    nil, // Using nil here so it's easier to notice if it's getting set when it shouldn't.
			expHoldGen: holdGen(),
			expBankGen: nil,
		},
		{
			name:     "no bank genesis state",
			seed:     1,
			accounts: accs[0:1],
			bankGen:  nil,
			expPanic: panicNoBal(ah(accs[0], 552)),
		},
		{
			name:     "1 account picked no balance",
			seed:     1,
			accounts: accs[0:1],
			bankGen:  &banktypes.GenesisState{},
			expPanic: panicNoBal(ah(accs[0], 552)),
		},
		{
			name:       "1 account picked already had balance",
			seed:       1,
			accounts:   accs[0:1],
			bankGen:    bankGen(bal(accs[0], 1000)),
			expHoldGen: holdGen(ah(accs[0], 552)),
			expBankGen: bankGen(bal(accs[0], 1552)),
		},
		{
			name:       "3 accounts none picked",
			seed:       3,
			accounts:   accs,
			bankGen:    bankGen(bal(accs[0], 333), bal(accs[1], 50), bal(accs[2], 1000)),
			expHoldGen: holdGen(),
			expBankGen: bankGen(bal(accs[0], 333), bal(accs[1], 50), bal(accs[2], 1000)),
		},
		{
			name:     "3 accounts 2 picked one already had a balance",
			seed:     0,
			accounts: accs,
			bankGen:  bankGen(bal(accs[0], 50), bal(accs[2], 1000)),
			expPanic: panicNoBal(ah(accs[1], 203)),
		},
		{
			name:       "3 accounts 2 picked all already have a balance",
			seed:       0,
			accounts:   accs,
			bankGen:    bankGen(bal(accs[0], 333), bal(accs[1], 50), bal(accs[2], 1000)),
			expHoldGen: holdGen(ah(accs[2], 795), ah(accs[1], 203)),
			expBankGen: bankGen(bal(accs[0], 333), bal(accs[1], 253), bal(accs[2], 1795)),
		},
		{
			name:      "2 from app params only one already had balance",
			appParams: []*hold.AccountHold{ah(accs[0], 123), ah(accs[1], 456)},
			bankGen:   bankGen(bal(accs[1], 44), bal(accs[2], 9876)),
			expPanic:  panicNoBal(ah(accs[0], 123)),
		},
		{
			name:       "2 from app params both already have balance",
			appParams:  []*hold.AccountHold{ah(accs[1], 123), ah(accs[0], 456)},
			bankGen:    bankGen(bal(accs[0], 44), bal(accs[1], 9876)),
			expHoldGen: holdGen(ah(accs[1], 123), ah(accs[0], 456)),
			expBankGen: bankGen(bal(accs[0], 500), bal(accs[1], 9999)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			simState := &module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
				Rand:      rand.New(rand.NewSource(tc.seed)),
				GenState:  make(map[string]json.RawMessage),
				Accounts:  tc.accounts,
			}
			var err error
			if len(tc.appParams) > 0 {
				simState.AppParams[simulation.HoldAccountHolds], err = json.Marshal(tc.appParams)
				require.NoError(t, err, "Marshal(tc.appParams)")
			}
			if tc.bankGen != nil {
				simState.GenState[banktypes.ModuleName], err = simState.Cdc.MarshalJSON(tc.bankGen)
				require.NoError(t, err, "MarshalJSON(tc.bankGen)")
			}

			// Run the function and capture any stdout. Make sure it either panics or doesn't as expected.
			var stdoutBuf bytes.Buffer
			var stdout string
			endStdoutCapture := startStdoutCapture(t, &stdoutBuf)
			defer func() {
				// make sure stdout capturing has ended.
				err = endStdoutCapture()
				if t.Failed() {
					// If the test failed, output stdout to the test log.
					if len(stdout) == 0 {
						stdout = stdoutBuf.String()
					}
					if err != nil {
						stdout = stdout + "\nerror capturing stdout: " + err.Error()
					}
					t.Logf("stdout during RandomizedGenState:\n%s\n", stdout)
				}
			}()
			testFunc := func() {
				simulation.RandomizedGenState(simState)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "RandomizedGenState")
			if len(tc.expPanic) > 0 {
				// No further testing to do.
				return
			}

			// End stdout capturing and get it.
			require.NoError(t, endStdoutCapture(), "ending stdout capturing")
			stdout = stdoutBuf.String()

			// Check the resulting hold genesis state.
			if assert.NotEmpty(t, simState.GenState[hold.ModuleName]) {
				holdGenState := &hold.GenesisState{}
				err = simState.Cdc.UnmarshalJSON(simState.GenState[hold.ModuleName], holdGenState)
				if assert.NoError(t, err, "UnmarshalJSON(hold gen state)") {
					assert.Equal(t, tc.expHoldGen, holdGenState, "hold gen state")
				}
			}

			// check the resulting bank genesis state.
			if tc.expBankGen == nil {
				assert.Empty(t, simState.GenState[banktypes.ModuleName])
			} else if assert.NotEmpty(t, simState.GenState[banktypes.ModuleName]) {
				bankGenState := &banktypes.GenesisState{}
				err = simState.Cdc.UnmarshalJSON(simState.GenState[banktypes.ModuleName], bankGenState)
				if assert.NoError(t, err, "UnmarshalJSON(bank gen state)") {
					assert.Equal(t, tc.expBankGen.Balances, bankGenState.Balances, "bank gen state balances")
					assert.Equal(t, tc.expBankGen.Supply.Sort(), bankGenState.Supply.Sort(), "bank gen state supply")
				}
			}

			// Make sure stdout has the hold genesis state message.
			expStdoutHolds := fmt.Sprintf("Selected %d randomly generated holds:\n", len(tc.expHoldGen.Holds)) +
				simulation.HoldsString(tc.expHoldGen.Holds) + "\n"
			assert.Contains(t, stdout, expStdoutHolds, "stdout message about holds\nExpected:\n%s", expStdoutHolds)
		})
	}
}
