package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEqualDefaultGenesis(t *testing.T) {
	state1 := GenesisState{}
	state2 := GenesisState{}
	require.Equal(t, state1, state2)
}
