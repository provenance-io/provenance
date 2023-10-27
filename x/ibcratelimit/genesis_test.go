package ibcratelimit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	expected := NewGenesisState(NewParams(""))
	genesis := DefaultGenesis()
	require.Equal(t, expected, genesis)
}

func TestGenesisValidate(t *testing.T) {
	testCases := []struct {
		name string
		addr string
		err  string
	}{
		{
			name: "success - valid address",
			addr: "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
		},
		{
			name: "success - empty address",
			addr: "",
		},
		{
			name: "failure - invalid address format",
			addr: "cosmos1234",
			err:  "decoding bech32 failed: invalid separator index 6",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			genesis := NewGenesisState(NewParams(tc.addr))
			err := genesis.Validate()

			if len(tc.err) > 0 {
				require.EqualError(t, err, tc.err, "should have the correct error")
			} else {
				require.NoError(t, err, "should not throw an error")
			}
		})
	}
}

func TestNewGenesisState(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected GenesisState
	}{
		{
			name:     "success - empty contract address can be used",
			expected: GenesisState{Params: NewParams("")},
		},
		{
			name:     "success - params are correctly set.",
			expected: GenesisState{Params: NewParams("cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			genesis := *NewGenesisState(NewParams(tc.expected.Params.ContractAddress))
			require.Equal(t, tc.expected, genesis)
		})
	}
}
