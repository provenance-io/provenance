package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenesisState(t *testing.T) {
	port := "random"
	oracle := "oracle"

	genesis := NewGenesisState(port, oracle)
	assert.Equal(t, port, genesis.PortId, "port id must match")
	assert.Equal(t, oracle, genesis.Oracle, "oracle must match")
}

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()

	assert.Equal(t, PortID, genesis.PortId, "port id must match")
	assert.Equal(t, "", genesis.Oracle, "oracle must match")
}

func TestGenesisValidate(t *testing.T) {
	tests := []struct {
		name  string
		state *GenesisState
		err   string
	}{
		{
			name:  "success - all fields are valid",
			state: NewGenesisState(PortID, "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
		},
		{
			name:  "success - all fields are valid with empty oracle",
			state: NewGenesisState(PortID, ""),
		},
		{
			name:  "failure - port id is invalid",
			state: NewGenesisState("x", ""),
			err:   "identifier x has invalid length: 1, must be between 2-128 characters: invalid identifier",
		},
		{
			name:  "failure - oracle is invalid",
			state: NewGenesisState(PortID, "abc"),
			err:   "decoding bech32 failed: invalid bech32 string length 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.state.Validate()
			if len(tc.err) > 0 {
				assert.EqualError(t, res, tc.err, "Genesis.Validate")
			} else {
				assert.NoError(t, res, "Genesis.Validate")
			}
		})
	}
}
