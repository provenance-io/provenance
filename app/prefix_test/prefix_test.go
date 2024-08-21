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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
)

func TestSetConfig(t *testing.T) {
	cases := []struct {
		name        string
		isTestnet   bool
		seal        bool
		expPath     string
		expHRP      string
		expCoinType uint32
		expPanic    string
	}{
		{
			name:        "mainnet",
			isTestnet:   false,
			seal:        false,
			expPath:     "m/44'/505'/0'/0/0",
			expHRP:      app.AccountAddressPrefixMainNet,
			expCoinType: app.CoinTypeMainNet,
		},
		{
			name:        "testnet",
			isTestnet:   true,
			seal:        false,
			expPath:     "m/44'/1'/0'/0/0",
			expHRP:      app.AccountAddressPrefixTestNet,
			expCoinType: app.CoinTypeTestNet,
		},
		{
			name:        "back to mainnet",
			isTestnet:   false,
			seal:        false,
			expPath:     "m/44'/505'/0'/0/0",
			expHRP:      app.AccountAddressPrefixMainNet,
			expCoinType: app.CoinTypeMainNet,
		},
		{
			// This is the last valid test, so we're doing seal = true.
			name:        "back to testnet with seal",
			isTestnet:   true,
			seal:        true,
			expPath:     "m/44'/1'/0'/0/0",
			expHRP:      app.AccountAddressPrefixTestNet,
			expCoinType: app.CoinTypeTestNet,
		},
		{
			// Note: A previous test should have had seal = true, making this an
			// attempt to change the config after sealing it.
			name:      "already sealed: mainnet, no reseal",
			isTestnet: false,
			seal:      false,
			expPanic:  "Config is sealed",
		},
		{
			// Note: A previous test should have had seal = true, making this an
			// attempt to change the config after sealing it.
			name:      "already sealed: testnet, no reseal",
			isTestnet: true,
			seal:      false,
			expPanic:  "Config is sealed",
		},
		{
			// Note: A previous test should have had seal = true, making this an
			// attempt to change the config after sealing it.
			name:      "already sealed: mainnet, with reseal",
			isTestnet: false,
			seal:      true,
			expPanic:  "Config is sealed",
		},
		{
			// Note: A previous test should have had seal = true, making this an
			// attempt to change the config after sealing it.
			name:      "already sealed: testnet, with reseal",
			isTestnet: true,
			seal:      true,
			expPanic:  "Config is sealed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testFunc := func() {
				app.SetConfig(tc.isTestnet, tc.seal)
			}
			if len(tc.expPanic) > 0 {
				require.PanicsWithValue(t, tc.expPanic, testFunc, "SetConfig")
			} else {
				require.NotPanics(t, testFunc, "SetConfig")

				expAddrPre := tc.expHRP
				expValPre := tc.expHRP + "valoper"
				expValPubPre := tc.expHRP + "valoperpub"
				expAddrPrePub := tc.expHRP + "pub"
				expConsPre := tc.expHRP + "valcons"
				expConsPubPre := tc.expHRP + "valconspub"

				sdkConfig := sdk.GetConfig()
				fullBIP44Path := sdkConfig.GetFullBIP44Path()
				addrPre := sdkConfig.GetBech32AccountAddrPrefix()
				addrPubPre := sdkConfig.GetBech32AccountPubPrefix()
				valPre := sdkConfig.GetBech32ValidatorAddrPrefix()
				valPubPre := sdkConfig.GetBech32ValidatorPubPrefix()
				consPre := sdkConfig.GetBech32ConsensusAddrPrefix()
				consPubPre := sdkConfig.GetBech32ConsensusPubPrefix()
				coinType := sdkConfig.GetCoinType()
				purpose := sdkConfig.GetPurpose()

				// Using require to check the main HRPs because if they're wrong, the rest will almost certainly also be wrong.
				require.Equal(t, expAddrPre, app.AccountAddressPrefix, "AccountAddressPrefix")
				require.Equal(t, expAddrPre, addrPre, "sdkConfig.GetBech32AccountAddrPrefix()")
				// Asserts from here on out so we get a larger picture upon failure.
				assert.Equal(t, expAddrPrePub, app.AccountPubKeyPrefix, "AccountPubKeyPrefix")
				assert.Equal(t, expAddrPrePub, addrPubPre, "sdkConfig.GetBech32AccountPubPrefix()")
				assert.Equal(t, expValPre, app.ValidatorAddressPrefix, "ValidatorAddressPrefix")
				assert.Equal(t, expValPre, valPre, "sdkConfig.GetBech32ValidatorAddrPrefix()")
				assert.Equal(t, expValPubPre, app.ValidatorPubKeyPrefix, "ValidatorPubKeyPrefix")
				assert.Equal(t, expValPubPre, valPubPre, "sdkConfig.GetBech32ValidatorPubPrefix()")
				assert.Equal(t, expConsPre, app.ConsNodeAddressPrefix, "ConsNodeAddressPrefix")
				assert.Equal(t, expConsPre, consPre, "sdkConfig.GetBech32ConsensusAddrPrefix()")
				assert.Equal(t, expConsPubPre, app.ConsNodePubKeyPrefix, "ConsNodePubKeyPrefix")
				assert.Equal(t, expConsPubPre, consPubPre, "sdkConfig.GetBech32ConsensusPubPrefix()")

				assert.Equal(t, tc.expPath, fullBIP44Path, "sdkConfig.GetFullBIP44Path()")
				assert.Equal(t, tc.expCoinType, app.CoinType, "CoinType - Exp = %d, Act = %d", tc.expCoinType, app.CoinType)
				assert.Equal(t, tc.expCoinType, coinType, "sdkConfig.GetCoinType() - Exp = %d, Act = %d", tc.expCoinType, coinType)
				assert.Equal(t, 44, app.Purpose, "Purpose")
				assert.Equal(t, 44, int(purpose), "sdkConfig.GetPurpose()")
			}
		})
	}
}
