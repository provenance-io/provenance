package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/name/simulation"
	nametypes "github.com/provenance-io/provenance/x/name/types"
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
			composedKey: "name/MinSegmentLength",
			key:         "MinSegmentLength",
			subspace:    nametypes.ModuleName,
		},
		{
			composedKey: "name/MaxSegmentLength",
			key:         "MaxSegmentLength",
			subspace:    nametypes.ModuleName,
		},
		{
			composedKey: "name/MaxNameLevels",
			key:         "MaxNameLevels",
			subspace:    nametypes.ModuleName,
		},
		{
			composedKey: "name/AllowUnrestrictedNames",
			key:         "AllowUnrestrictedNames",
			subspace:    nametypes.ModuleName,
		},
	}

	paramChanges := simulation.ParamChanges(r)

	require.Len(t, paramChanges, 4)

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
