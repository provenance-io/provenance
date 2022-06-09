package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
)

func TestCreateParams(t *testing.T) {
	msgFeeParam := NewParams(sdk.Coin{
		Denom:  "steak",
		Amount: sdk.NewInt(2000),
	}, uint64(7))
	require.Equal(t, sdk.Coin{
		Denom:  "steak",
		Amount: sdk.NewInt(2000),
	}, msgFeeParam.FloorGasPrice)
	require.Equal(t, uint64(7), msgFeeParam.UsdConversionRate)
}

func TestCreateParamSet(t *testing.T) {
	msgFeeParam := NewParams(sdk.Coin{
		Denom:  NhashDenom,
		Amount: sdk.NewInt(3000),
	}, uint64(7))
	paramsetPair := msgFeeParam.ParamSetPairs()
	require.Equal(t, 2, len(paramsetPair))
}

func TestValidateMinGasPriceParamI(t *testing.T) {
	require.NoError(t, validateCoinParam(sdk.Coin{
		Denom:  "steak",
		Amount: sdk.NewInt(2000),
	}))
}

func TestValidateUsdConversionRateParamI(t *testing.T) {
	require.NoError(t, validateUsdConversionRateParam(uint64(7)))
}

func TestMsgFeeParamKeyTable(t *testing.T) {
	keyTable := ParamKeyTable()
	require.Panics(t, func() {
		keyTable.RegisterType(paramtypes.NewParamSetPair(ParamStoreKeyFloorGasPrice, sdk.Coin{
			Denom:  NhashDenom,
			Amount: sdk.NewInt(5000),
		}, validateCoinParam))
	})
	require.Panics(t, func() {
		keyTable.RegisterType(paramtypes.NewParamSetPair(ParamStoreKeyUsdConversionRate, uint64(7), validateUsdConversionRateParam))
	})
}

func TestDefault(t *testing.T) {
	metadataData := DefaultParams()
	require.Equal(t, DefaultFloorGasPrice, metadataData.FloorGasPrice)
	require.Equal(t, DefaultUsdConversionRate, metadataData.UsdConversionRate)
}
