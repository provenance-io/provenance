package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestFullBIP44Path(t *testing.T) {
	cases := []struct {
		name         string
		expected     string
		isTestnet    bool
		seal         bool
		panicMessage string
	}{
		{
			// Note: The way our SetConfig function works, calling it with isTestnet = true makes some changes that aren't undone
			// when later calling it with isTestnet = false. So to test the mainnet BIP44, we need to do it first.
			name:         "has correct bip44th path for mainnet",
			expected:     "m/44'/505'/0'/0/0",
			isTestnet:    false,
			seal:         false,
			panicMessage: "",
		},
		{
			// Note: This is the 2nd to last test, so we're doing seal = true.
			name:         "has correct bip44th path for testnet",
			expected:     "m/44'/1'/0'/0/0",
			isTestnet:    true,
			seal:         true,
			panicMessage: "",
		},
		{
			// Note: The previous test should have had seal = true, making this an
			// attempt to change the config after sealing it.
			name:         "cannot double seal",
			expected:     "",
			isTestnet:    false,
			seal:         true,
			panicMessage: "Config is sealed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.panicMessage) > 0 {
				require.PanicsWithValue(t, tc.panicMessage, func() {
					SetConfig(tc.isTestnet, tc.seal)
				}, "SetConfig")
			} else {
				require.NotPanics(t, func() {
					SetConfig(tc.isTestnet, tc.seal)
				}, "SetConfig")
				config := sdk.GetConfig()
				fullBIP44Path := config.GetFullBIP44Path()
				require.Equal(t, tc.expected, fullBIP44Path, "GetFullBIP44Path")
			}

		})
	}
}
