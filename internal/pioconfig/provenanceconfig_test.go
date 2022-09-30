package pioconfig

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigSetForCustomDenom(t *testing.T) {
	SetProvenanceConfig("hotdog", 100)
	assert.Equal(t, GetProvenanceConfig().BondDenom, "hotdog")
	assert.Equal(t, GetProvenanceConfig().FeeDenom, "hotdog")
	assert.Equal(t, GetProvenanceConfig().MsgFeeFloorGasPrice, int64(100))
	assert.Equal(t, GetProvenanceConfig().ProvenanceMinGasPrices, "100hotdog")
}

func TestConfigSetRegularDenom(t *testing.T) {
	SetProvenanceConfig("", 0)
	assert.Equal(t, GetProvenanceConfig().BondDenom, "nhash")
	assert.Equal(t, GetProvenanceConfig().FeeDenom, "nhash")
	assert.Equal(t, GetProvenanceConfig().MsgFeeFloorGasPrice, int64(1905))
	assert.Equal(t, GetProvenanceConfig().ProvenanceMinGasPrices, "1905nhash")
}

// TestConfigSetRegularDenomCustomMsgFloorFeeIgnoredForNhash msg fee floor passed in as zero and will be set as default == 1905nhash
// (for backwards compatibility when these flags were not around.)
func TestConfigSetRegularDenomCustomMsgFloorFeeIgnoredForNhash(t *testing.T) {
	// doesn't matter still setting nhash base fee as 1905, even though it can and should be changed by governance.
	SetProvenanceConfig("", 0)
	assert.Equal(t, GetProvenanceConfig().BondDenom, "nhash")
	assert.Equal(t, GetProvenanceConfig().FeeDenom, "nhash")
	assert.Equal(t, GetProvenanceConfig().MsgFeeFloorGasPrice, int64(1905))
	assert.Equal(t, GetProvenanceConfig().ProvenanceMinGasPrices, "1905nhash")
}

// TestConfigSetRegularDenomCustomMsgFloorFeeNotIgnoredForNhash msg fee floor passed in as non-zero and will be not be ignored
// (assumes caller knows what they are doing)
// (for backwards compatibility when these flags were not around.)
func TestConfigSetRegularDenomCustomMsgFloorFeeNotIgnoredForNhash(t *testing.T) {
	SetProvenanceConfig("", 18)
	assert.Equal(t, GetProvenanceConfig().BondDenom, "nhash")
	assert.Equal(t, GetProvenanceConfig().FeeDenom, "nhash")
	assert.Equal(t, GetProvenanceConfig().MsgFeeFloorGasPrice, int64(18))
	assert.Equal(t, GetProvenanceConfig().ProvenanceMinGasPrices, "18nhash")
}

// all code flows shows set the config for e.g. root cmd etc
func TestGetConfigNotSet(t *testing.T) {
	provConfig = nil
	assert.Equal(t,
		GetProvenanceConfig(), ProvenanceConfig{}, "Should get empty config if not set, several things in app wiring will fail fast if this not set so not too worried.")
}

func TestNotChangeableOutsideSet(t *testing.T) {
	SetProvenanceConfig("foo", 8)
	cfg1 := GetProvenanceConfig()
	cfg1.FeeDenom = "fee"
	cfg1.BondDenom = "bond"
	cfg1.MsgFloorDenom = "floor"
	cfg1.MsgFeeFloorGasPrice = 50
	cfg1.ProvenanceMinGasPrices = "50floor"

	cfg2 := GetProvenanceConfig()
	assert.Equal(t, "foo", cfg2.FeeDenom)
	assert.Equal(t, "foo", cfg2.BondDenom)
	assert.Equal(t, "foo", cfg2.MsgFloorDenom)
	assert.Equal(t, 8, int(cfg2.MsgFeeFloorGasPrice))
	assert.Equal(t, "8foo", cfg2.ProvenanceMinGasPrices)
}
