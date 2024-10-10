package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
)

func TestCreateParams(t *testing.T) {
	msgFeeParam := NewParams(sdk.Coin{
		Denom:  "steak",
		Amount: sdkmath.NewInt(2000),
	}, uint64(7), "nhash")
	assert.Equal(t, sdk.Coin{
		Denom:  "steak",
		Amount: sdkmath.NewInt(2000),
	}, msgFeeParam.FloorGasPrice)
	assert.Equal(t, uint64(7), msgFeeParam.NhashPerUsdMil)
	assert.Equal(t, "nhash", msgFeeParam.ConversionFeeDenom)
}

func TestDefault(t *testing.T) {
	msgFeeData := DefaultParams()
	assert.Equal(t, DefaultFloorGasPrice(), msgFeeData.FloorGasPrice)
	assert.Equal(t, DefaultNhashPerUsdMil, msgFeeData.NhashPerUsdMil)
	assert.Equal(t, pioconfig.GetProvenanceConfig().FeeDenom, msgFeeData.ConversionFeeDenom)
}
