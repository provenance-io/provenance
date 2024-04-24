package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"

	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/gogoproto/proto"
)

// MutateGenesisState will extract a GenesisState object from the provided config and run it through
// the provided mutator. Then it will update the config to have that new genesis state.
//
// G is the type of the specific genesis state struct being used here.
func MutateGenesisState[G proto.Message](t *testing.T, cfg *testnet.Config, moduleName string, emptyState G, mutator func(state G) G) {
	t.Helper()
	err := cfg.Codec.UnmarshalJSON(cfg.GenesisState[moduleName], emptyState)
	if err != nil {
		t.Logf("initial %s genesis state:\n%s", moduleName, string(cfg.GenesisState[moduleName]))
	}
	require.NoError(t, err, "UnmarshalJSON %s genesis state as %T", moduleName, emptyState)

	emptyState = mutator(emptyState)

	cfg.GenesisState[moduleName], err = cfg.Codec.MarshalJSON(emptyState)
	if err != nil {
		t.Logf("%s genesis state after mutator:\n%#v", moduleName, emptyState)
	}
	require.NoError(t, err, "MarshalJSON %s genesis state from %T after mutator", moduleName, emptyState)
}
