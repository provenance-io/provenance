package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/escrow"
	"github.com/provenance-io/provenance/x/escrow/simulation"
)

func TestRandomAccountEscrows(t *testing.T) {
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 3)

	tests := []struct {
		name     string
		seed     int64
		accounts []simtypes.Account
		expected []*escrow.AccountEscrow
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
			expected: []*escrow.AccountEscrow{
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
			expected: []*escrow.AccountEscrow{
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
			expected: []*escrow.AccountEscrow{
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
			name:     "3 accounts 2 picked",
			seed:     0,
			accounts: accs,
			expected: []*escrow.AccountEscrow{
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
			escrows := simulation.RandomAccountEscrows(r, tc.accounts)
			assert.Equal(t, tc.expected, escrows, "RandomAccountEscrows result")
		})
	}
}

func TestRandomizedGenState(t *testing.T) {
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 3)

	gs := func(escrows ...*escrow.AccountEscrow) *escrow.GenesisState {
		rv := &escrow.GenesisState{
			Escrows: make([]*escrow.AccountEscrow, len(escrows)),
		}
		copy(rv.Escrows, escrows)
		return rv
	}
	bondCoins := func(amount int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, amount))
	}
	ae := func(acc simtypes.Account, amount int64) *escrow.AccountEscrow {
		return &escrow.AccountEscrow{
			Address: acc.Address.String(),
			Amount:  bondCoins(amount),
		}
	}
	bgs := func(bals ...banktypes.Balance) *banktypes.GenesisState {
		rv := &banktypes.GenesisState{
			Supply:        sdk.Coins{},
			Balances:      bals,
			SendEnabled:   make([]banktypes.SendEnabled, 0),
			DenomMetadata: make([]banktypes.Metadata, 0),
			Params: banktypes.Params{
				SendEnabled:        make([]*banktypes.SendEnabled, 0),
				DefaultSendEnabled: false,
			},
		}
		for _, bal := range rv.Balances {
			rv.Supply = rv.Supply.Add(bal.Coins...)
		}
		return rv
	}
	bal := func(acc simtypes.Account, amount int64) banktypes.Balance {
		return banktypes.Balance{Address: acc.Address.String(), Coins: bondCoins(amount)}
	}

	tests := []struct {
		name         string
		seed         int64
		appParams    []*escrow.AccountEscrow
		accounts     []simtypes.Account
		bankState    *banktypes.GenesisState
		expGenState  *escrow.GenesisState
		expBankState *banktypes.GenesisState
	}{
		{
			name:         "no accounts",
			seed:         0,
			expGenState:  gs(),
			expBankState: nil,
		},
		{
			name:         "1 account picked no balance",
			seed:         1,
			accounts:     accs[0:1],
			bankState:    &banktypes.GenesisState{},
			expGenState:  gs(ae(accs[0], 552)),
			expBankState: bgs(bal(accs[0], 552)),
		},
		{
			name:         "1 account picked already had balance",
			seed:         1,
			accounts:     accs[0:1],
			bankState:    bgs(bal(accs[0], 1000)),
			expGenState:  gs(ae(accs[0], 552)),
			expBankState: bgs(bal(accs[0], 1552)),
		},
		{
			name:         "3 accounts 2 picked one already had a balance",
			seed:         0,
			accounts:     accs,
			bankState:    bgs(bal(accs[0], 50), bal(accs[2], 1000)),
			expGenState:  gs(ae(accs[2], 795), ae(accs[1], 203)),
			expBankState: bgs(bal(accs[0], 50), bal(accs[2], 1795), bal(accs[1], 203)),
		},
		{
			name:         "2 from app params one already had balance",
			appParams:    []*escrow.AccountEscrow{ae(accs[0], 123), ae(accs[1], 456)},
			bankState:    bgs(bal(accs[1], 44), bal(accs[2], 9876)),
			expGenState:  gs(ae(accs[0], 123), ae(accs[1], 456)),
			expBankState: bgs(bal(accs[1], 500), bal(accs[2], 9876), bal(accs[0], 123)),
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
				simState.AppParams[simulation.EscrowAccountEscrows], err = json.Marshal(tc.appParams)
				require.NoError(t, err, "Marshal(tc.appParams)")
			}
			if tc.bankState != nil {
				simState.GenState[banktypes.ModuleName], err = simState.Cdc.MarshalJSON(tc.bankState)
				require.NoError(t, err, "MarshalJSON(tc.bankState)")
			}

			testFunc := func() {
				simulation.RandomizedGenState(simState)
			}
			require.NotPanics(t, testFunc, "RandomizedGenState")

			if assert.NotEmpty(t, simState.GenState[escrow.ModuleName]) {
				escrowGenState := &escrow.GenesisState{}
				err = simState.Cdc.UnmarshalJSON(simState.GenState[escrow.ModuleName], escrowGenState)
				if assert.NoError(t, err, "UnmarshalJSON(escrow gen state)") {
					assert.Equal(t, tc.expGenState, escrowGenState, "escrow gen state")
				}
			}

			if tc.expBankState == nil {
				assert.Empty(t, simState.GenState[banktypes.ModuleName])
			} else if assert.NotEmpty(t, simState.GenState[banktypes.ModuleName]) {
				bankGenState := &banktypes.GenesisState{}
				err = simState.Cdc.UnmarshalJSON(simState.GenState[banktypes.ModuleName], bankGenState)
				if assert.NoError(t, err, "UnmarshalJSON(bank gen state)") {
					assert.Equal(t, tc.expBankState, bankGenState, "bank gen state")
				}
			}
		})
	}
}
