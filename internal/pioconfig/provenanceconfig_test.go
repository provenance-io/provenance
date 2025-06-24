package pioconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGetProvenanceConfig(t *testing.T) {
	provConfig = nil // Make sure it's not yet set to start these tests.

	tests := []struct {
		name     string
		feeDenom string
		exp      ProvConfig
	}{
		{
			name: "not set yet",
			exp:  ProvConfig{FeeDenom: "nhash", BondDenom: "nhash", ProvMinGasPrices: "0nhash"},
		},
		{
			name:     "water",
			feeDenom: "water",
			exp:      ProvConfig{FeeDenom: "water", BondDenom: "water", ProvMinGasPrices: "0water"},
		},
		{
			name:     "empty fee denom",
			feeDenom: "",
			exp:      ProvConfig{FeeDenom: "nhash", BondDenom: "nhash", ProvMinGasPrices: "0nhash"},
		},
		{
			name:     "atom",
			feeDenom: "atom",
			exp:      ProvConfig{FeeDenom: "atom", BondDenom: "atom", ProvMinGasPrices: "0atom"},
		},
		{
			name:     "nhash",
			feeDenom: "nhash",
			exp:      ProvConfig{FeeDenom: "nhash", BondDenom: "nhash", ProvMinGasPrices: "0nhash"},
		},
		{
			name:     "beans",
			feeDenom: "beans",
			exp:      ProvConfig{FeeDenom: "beans", BondDenom: "beans", ProvMinGasPrices: "0beans"},
		},
		{
			// Shouldn't cause a problem, at least in this test. I'm sure other things would blow up though.
			name:     "invalid denom",
			feeDenom: "$inthebank",
			exp:      ProvConfig{FeeDenom: "$inthebank", BondDenom: "$inthebank", ProvMinGasPrices: "0$inthebank"},
		},
		{
			name:     "empty fee denom again",
			feeDenom: "",
			exp:      ProvConfig{FeeDenom: "nhash", BondDenom: "nhash", ProvMinGasPrices: "0nhash"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.feeDenom != "DO NOT SET" {
				testSet := func() {
					SetProvConfig(tc.feeDenom)
				}
				require.NotPanics(t, testSet, "SetProvConfig(%q)", tc.feeDenom)
			}

			var act ProvConfig
			testGet := func() {
				act = GetProvConfig()
			}
			require.NotPanics(t, testGet, "GetProvConfig after SetProvConfig(%q)", tc.feeDenom)
			assert.Equal(t, tc.exp, act, "GetProvConfig after SetProvConfig(%q)", tc.feeDenom)
		})
	}
}

// all code flows shows set the config for e.g. root cmd etc
func TestGetConfigNotSet(t *testing.T) {
	provConfig = nil // Make sure it's not yet set to start this tests.
	exp := ProvConfig{FeeDenom: "nhash", BondDenom: "nhash", ProvMinGasPrices: "0nhash"}
	var act ProvConfig
	testGet := func() {
		act = GetProvConfig()
	}
	require.NotPanics(t, testGet, "GetProvConfig()")
	assert.Equal(t, exp, act, "GetProvConfig() result")
}

func TestNotChangeableOutsideSet(t *testing.T) {
	SetProvConfig("foo")
	exp := ProvConfig{FeeDenom: "foo", BondDenom: "foo", ProvMinGasPrices: "0foo"}
	orig := GetProvConfig()
	// Requiring (not asserting) this because if this is wrong, the rest will fail too.
	require.Equal(t, exp, orig, "Starting config")

	// Get the config and make some changes to it, then make sure those changes aren't reflected anywhere else.
	toChange := GetProvConfig()
	toChange.FeeDenom = "feedme"
	toChange.BondDenom = "seymour"
	toChange.ProvMinGasPrices = "10drops"

	assert.Equal(t, exp, orig, "Starting config after changing stuff in another config.")

	cfg2 := GetProvConfig()
	assert.Equal(t, exp, cfg2, "GetProvConfig() again after changing stuff in another config.")
}
