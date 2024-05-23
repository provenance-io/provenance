package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/provenance-io/provenance/x/ibcratelimit/simulation"
)

func TestContractFn(t *testing.T) {
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 3)

	tests := []struct {
		name     string
		seed     int64
		expected string
		accounts []simtypes.Account
	}{
		{
			name:     "success - returns an empty account",
			seed:     3,
			accounts: accs,
			expected: "",
		},
		{
			name:     "success - returns a random account",
			seed:     0,
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
			port := simulation.ContractFn(r, tc.accounts)
			assert.Equal(t, tc.expected, port, "should return correct random contract")
		})
	}
}

func TestRandomizedGenState(t *testing.T) {
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 3)
	tests := []struct {
		name            string
		seed            int64
		expRateLimitGen *ibcratelimit.GenesisState
		accounts        []simtypes.Account
	}{
		{
			name:     "success - can handle no accounts",
			seed:     0,
			accounts: nil,
			expRateLimitGen: &ibcratelimit.GenesisState{
				Params: ibcratelimit.NewParams(""),
			},
		},
		{
			name:     "success - can handle accounts",
			seed:     1,
			accounts: accs,
			expRateLimitGen: &ibcratelimit.GenesisState{
				Params: ibcratelimit.NewParams(""),
			},
		},
		{
			name:     "success - has different output",
			seed:     2,
			accounts: accs,
			expRateLimitGen: &ibcratelimit.GenesisState{
				Params: ibcratelimit.NewParams("cosmos12jszjrc0qhjt0ugt2uh4ptwu0h55pq6qfp9ecl"),
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

			if assert.NotEmpty(t, simState.GenState[ibcratelimit.ModuleName]) {
				rateLimitGenState := &ibcratelimit.GenesisState{}
				err := simState.Cdc.UnmarshalJSON(simState.GenState[ibcratelimit.ModuleName], rateLimitGenState)
				if assert.NoError(t, err, "UnmarshalJSON(ratelimitedibc gen state)") {
					assert.Equal(t, tc.expRateLimitGen, rateLimitGenState, "hold ratelimitedibc state")
				}
			}
		})
	}
}
