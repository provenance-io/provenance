package ibcratelimit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateContractAddress(t *testing.T) {
	testCases := map[string]struct {
		addr     interface{}
		expected bool
	}{
		"valid_addr": {
			addr:     "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
			expected: true,
		},
		"invalid_addr": {
			addr:     "cosmos1234",
			expected: false,
		},
		"invalid parameter type": {
			addr:     123456,
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := validateContractAddress(tc.addr)

			// Assertions.
			if !tc.expected {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateParams(t *testing.T) {
	testCases := map[string]struct {
		addr     interface{}
		expected bool
	}{
		"valid_addr": {
			addr:     "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
			expected: true,
		},
		"invalid_addr": {
			addr:     "cosmos1234",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			addr, ok := tc.addr.(string)
			require.True(t, ok, "unexpected type of address")

			params := Params{
				ContractAddress: addr,
			}

			err := params.Validate()

			// Assertions.
			if !tc.expected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewParams(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			name: "success - empty contract address can be used",
			addr: "",
		},
		{
			name: "success - address is correctly set.",
			addr: "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := NewParams(tc.addr)
			require.Equal(t, tc.addr, params.ContractAddress)
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.Equal(t, "", params.ContractAddress)
}
