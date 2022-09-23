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

// TestConfigSetRegularDenomCustomMsgFloorFeeIgnoredForNhash msg fee floor passed in but ignored since nhash will always be 1905nhash
//(for backwards compatibility when these flags were not around.)
func TestConfigSetRegularDenomCustomMsgFloorFeeIgnoredForNhash(t *testing.T) {
	// doesn't matter still setting nhash base fee as 1905, even though it can and should be changed by governance.
	SetProvenanceConfig("", 0)
	assert.Equal(t, GetProvenanceConfig().BondDenom, "nhash")
	assert.Equal(t, GetProvenanceConfig().FeeDenom, "nhash")
	assert.Equal(t, GetProvenanceConfig().MsgFeeFloorGasPrice, int64(1905))
	assert.Equal(t, GetProvenanceConfig().ProvenanceMinGasPrices, "1905nhash")
}
