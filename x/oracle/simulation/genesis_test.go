package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/provenance-io/provenance/x/oracle/simulation"
	"github.com/provenance-io/provenance/x/oracle/types"
	"github.com/stretchr/testify/assert"
)

func TestPortFn(t *testing.T) {
	tests := []struct {
		name     string
		seed     int64
		expected string
	}{
		{
			name:     "success - returns a random port that is not the oracle",
			seed:     0,
			expected: "vipxlpbshz",
		},
		{
			name:     "success - returns the oracle port",
			seed:     1,
			expected: "oracle",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(tc.seed))
			port := simulation.PortFn(r)
			assert.Equal(t, tc.expected, port, "should return correct random port")
		})
	}
}

func TestOracleFn(t *testing.T) {
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 3)

	tests := []struct {
		name     string
		seed     int64
		expected string
		accounts []simtypes.Account
	}{
		{
			name:     "success - returns an empty account",
			seed:     0,
			accounts: accs,
			expected: "",
		},
		{
			name:     "success - returns a random account",
			seed:     3,
			accounts: accs,
			expected: "cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc",
		},
		{
			name:     "success - returns a different random account",
			seed:     2,
			accounts: accs,
			expected: "cosmos12jszjrc0qhjt0ugt2uh4ptwu0h55pq6qfp9ecl",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(tc.seed))
			port := simulation.OracleFn(r, tc.accounts)
			assert.Equal(t, tc.expected, port, "should return correct random oracle")
		})
	}
}

func TestRandomizedGenState(t *testing.T) {
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 3)
	tests := []struct {
		name         string
		seed         int64
		expOracleGen *types.GenesisState
		accounts     []simtypes.Account
	}{
		{
			name:     "success - can handle no accounts",
			seed:     0,
			accounts: nil,
			expOracleGen: &types.GenesisState{
				PortId: "vipxlpbshz",
				Oracle: "",
			},
		},
		{
			name:     "success - can handle accounts",
			seed:     1,
			accounts: accs,
			expOracleGen: &types.GenesisState{
				PortId: "oracle",
				Oracle: "",
			},
		},
		{
			name:     "success - has different output",
			seed:     2,
			accounts: accs,
			expOracleGen: &types.GenesisState{
				PortId: "knxndtw",
				Oracle: "cosmos10gqqppkly524p6v7hypvvl8sn7wky85jajrph0",
			},
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
			simulation.RandomizedGenState(simState)

			if assert.NotEmpty(t, simState.GenState[types.ModuleName]) {
				oracleGenState := &types.GenesisState{}
				err := simState.Cdc.UnmarshalJSON(simState.GenState[types.ModuleName], oracleGenState)
				if assert.NoError(t, err, "UnmarshalJSON(oracle gen state)") {
					assert.Equal(t, tc.expOracleGen, oracleGenState, "hold oracle state")
				}
			}
		})
	}
}

func TestRandomChannel(t *testing.T) {

	tests := []struct {
		name     string
		seed     int64
		expected string
	}{
		{
			name:     "success - returns a random channel",
			seed:     3,
			expected: "channel-8",
		},
		{
			name:     "success - returns a different random channel",
			seed:     2,
			expected: "channel-786",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(tc.seed))
			port := simulation.RandomChannel(r)
			assert.Equal(t, tc.expected, port, "should return correct random channel")
		})
	}
}
