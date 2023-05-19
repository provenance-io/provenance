package prefix_test

// These tests are for things in the app/prefix.go file.
// They're in this other folder and package so that they don't interfere with
// the other tests in the app folder, e.g. the stuff in upgrade_test.go.
//
// These tests alter some global variables in a way that can't easily be undone.
// For example, the bech32 address HRP is changed from the default "cosmos" to
// "tp" during these tests.
//
// When this file was in the app/ folder, any tests that created an app would
// pass when run individually, but fail when being run with all tests.
// What was happening is that, during app setup, a genesis file is created using
// the default "cosmos" HRP on all the addresses. But by the time it tries to
// read/parse that file, the tests in this file have run, changing the HRP to tp
// and locking the config (so it can't be changed back).
//
// By putting these tests in their own folder, they are invoked on their own,
// separately from any other tests, and without modifying global variables used
// by any other tests.

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
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
					app.SetConfig(tc.isTestnet, tc.seal)
				}, "SetConfig")
			} else {
				require.NotPanics(t, func() {
					app.SetConfig(tc.isTestnet, tc.seal)
				}, "SetConfig")
				config := sdk.GetConfig()
				fullBIP44Path := config.GetFullBIP44Path()
				require.Equal(t, tc.expected, fullBIP44Path, "GetFullBIP44Path")
			}
		})
	}
}
