package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/marker/simulation"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func TestParamChanges(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	expected := []struct {
		composedKey string
		key         string
		subspace    string
	}{
		{
			composedKey: "marker/MaxTotalSupply",
			key:         "MaxTotalSupply",
			subspace:    markertypes.ModuleName,
		},
		{
			composedKey: "marker/EnableGovernance",
			key:         "EnableGovernance",
			subspace:    markertypes.ModuleName,
		},
		{
			composedKey: "marker/UnrestrictedDenomRegex",
			key:         "UnrestrictedDenomRegex",
			subspace:    markertypes.ModuleName,
		},
	}

	paramChanges := simulation.ParamChanges(r)

	require.Len(t, paramChanges, 3)

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
