package ibcratelimit_test

import (
	"testing"

	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/stretchr/testify/assert"
)

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
			assert.True(t, ok, "unexpected type of address")

			params := ibcratelimit.Params{
				ContractAddress: addr,
			}

			err := params.Validate()

			if !tc.expected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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
			params := ibcratelimit.NewParams(tc.addr)
			assert.Equal(t, tc.addr, params.ContractAddress)
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := ibcratelimit.DefaultParams()
	assert.Equal(t, "", params.ContractAddress)
}
