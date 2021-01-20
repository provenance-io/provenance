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
			composedKey: "name/minsegmentlength",
			key:         "minsegmentlength",
			subspace:    nametypes.ModuleName,
		},
		{
			composedKey: "name/maxsegmentlength",
			key:         "maxsegmentlength",
			subspace:    nametypes.ModuleName,
		},
		{
			composedKey: "name/maxnamelevels",
			key:         "maxnamelevels",
			subspace:    nametypes.ModuleName,
		},
		{
			composedKey: "name/unrestricednames",
			key:         "unrestricednames",
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
