package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestFullBIP44PathForTestnet(t *testing.T) {
	cases := []struct {
		name         string
		expected     string
		isTestnet    bool
		seal         bool
		shouldPanic  bool
		panicMessage string
	}{
		{
			"has correct bip44th path for testnet",
			"m/44'/118'/0'/0/0",
			true,
			false,
			false,
			"",
		},
		{
			"has correct bip44th path for mainnet",
			"m/44'/118'/0'/0/0",
			true,
			true,
			false,
			"",
		},
		{
			"cannot double seal",
			"m/44'/118'/0'/0/0",
			true,
			true,
			true,
			"Config is sealed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := sdk.NewConfig()
			if tc.shouldPanic {
				require.Panics(t, func() {
					SetConfig(tc.isTestnet, tc.seal)
				}, tc.name)
			} else {
				SetConfig(tc.isTestnet, tc.seal)
				require.Equal(t, tc.expected, config.GetFullBIP44Path(), tc.name)
			}

		})
	}
}
