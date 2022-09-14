package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/provenance-io/provenance/internal/pioconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateParams(t *testing.T) {
	msgFeeParam := NewParams(sdk.Coin{
		Denom:  "steak",
		Amount: sdk.NewInt(2000),
	}, uint64(7), pioconfig.GetProvenanceConfig().FeeDenom)
	assert.Equal(t, sdk.Coin{
		Denom:  "steak",
		Amount: sdk.NewInt(2000),
	}, msgFeeParam.FloorGasPrice)
	assert.Equal(t, uint64(7), msgFeeParam.NhashPerUsdMil)
	assert.Equal(t, "nhash", msgFeeParam.ConversionFeeDenom)

}

func TestCreateParamSet(t *testing.T) {
	msgFeeParam := NewParams(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(3000),
	}, uint64(7), "nhash")
	paramsetPair := msgFeeParam.ParamSetPairs()
	require.Equal(t, 3, len(paramsetPair))
}

func TestValidateMinGasPriceParamI(t *testing.T) {
	require.NoError(t, validateCoinParam(sdk.Coin{
		Denom:  "steak",
		Amount: sdk.NewInt(2000),
	}))
}

func TestValidateUsdConversionRateParamI(t *testing.T) {
	require.NoError(t, validateNhashPerUsdMilParam(uint64(7)))
}

func TestValidateConversionFeeDenomParamI(t *testing.T) {
	require.NoError(t, validateConversionFeeDenomParam("nhash"))
}

func TestMsgFeeParamKeyTable(t *testing.T) {
	keyTable := ParamKeyTable()
	require.Panics(t, func() {
		keyTable.RegisterType(paramtypes.NewParamSetPair(ParamStoreKeyFloorGasPrice, sdk.Coin{
			Denom:  "nhash",
			Amount: sdk.NewInt(5000),
		}, validateCoinParam))
	})
	require.Panics(t, func() {
		keyTable.RegisterType(paramtypes.NewParamSetPair(ParamStoreKeyNhashPerUsdMil, uint64(7), validateNhashPerUsdMilParam))
	})
}

func TestDefault(t *testing.T) {
	msgFeeData := DefaultParams()
	assert.Equal(t, DefaultFloorGasPrice(), msgFeeData.FloorGasPrice)
	assert.Equal(t, DefaultNhashPerUsdMil, msgFeeData.NhashPerUsdMil)
	assert.Equal(t, pioconfig.GetProvenanceConfig().FeeDenom, msgFeeData.ConversionFeeDenom)
}
