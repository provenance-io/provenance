package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenesisState(t *testing.T) {
	port := "random"
	sequence := uint64(10)
	oracle := "oracle"

	genesis := NewGenesisState(port, sequence, oracle)
	assert.Equal(t, port, genesis.PortId, "port id must match")
	assert.Equal(t, int(sequence), int(genesis.Sequence), "sequence must match")
	assert.Equal(t, oracle, genesis.Oracle, "oracle must match")
}

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()

	assert.Equal(t, PortID, genesis.PortId, "port id must match")
	assert.Equal(t, int(1), int(genesis.Sequence), "sequence must be 1")
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
			state: NewGenesisState(PortID, 1, "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
		},
		{
			name:  "success - all fields are valid with empty oracle",
			state: NewGenesisState(PortID, 1, ""),
		},
		{
			name:  "failure - port id is invalid",
			state: NewGenesisState("x", 1, ""),
			err:   "identifier x has invalid length: 1, must be between 2-128 characters: invalid identifier",
		},
		{
			name:  "failure - sequence id is invalid",
			state: NewGenesisState(PortID, 0, ""),
			err:   "sequence 0 is invalid, must be greater than 0",
		},
		{
			name:  "failure - oracle is invalid",
			state: NewGenesisState(PortID, 1, "abc"),
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
