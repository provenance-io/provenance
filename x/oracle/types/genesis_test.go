package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenesisState(t *testing.T) {
	port := "random"
	params := DefaultParams()
	sequence := uint64(10)

	genesis := NewGenesisState(port, params, sequence)
	assert.Equal(t, port, genesis.PortId, "port id must match")
	assert.Equal(t, params, genesis.Params, "params must match")
	assert.Equal(t, int(sequence), int(genesis.Sequence), "sequence must match")
}

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()
	params := DefaultParams()

	assert.Equal(t, PortID, genesis.PortId, "port id must match")
	assert.Equal(t, params, genesis.Params, "params must match")
	assert.Equal(t, int(1), int(genesis.Sequence), "sequence must be 1")
}

func TestGenesisValidate(t *testing.T) {
	tests := []struct {
		name  string
		state *GenesisState
		err   string
	}{
		{
			name:  "success - all fields are valid",
			state: NewGenesisState(PortID, DefaultParams(), 1),
		},
		{
			name:  "failure - port id is invalid",
			state: NewGenesisState("x", DefaultParams(), 1),
			err:   "identifier x has invalid length: 1, must be between 2-128 characters: invalid identifier",
		},
		{
			name:  "failure - sequence id is invalid",
			state: NewGenesisState(PortID, DefaultParams(), 0),
			err:   "sequence 0 is invalid, must be greater than 0",
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
