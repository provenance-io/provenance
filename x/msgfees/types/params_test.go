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
	})
	require.Equal(t, sdk.Coin{
		Denom:  "steak",
		Amount: sdk.NewInt(2000),
	}, msgFeeParam.FloorGasPrice)
}

func TestCreateParamSet(t *testing.T) {
	msgFeeParam := NewParams(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(3000),
	})
	paramsetPair := msgFeeParam.ParamSetPairs()
	require.Equal(t, 1, len(paramsetPair))
}

func TestValidateMinGasPriceParamI(t *testing.T) {
	require.NoError(t, validateCoinParam(sdk.Coin{
		Denom:  "steak",
		Amount: sdk.NewInt(2000),
	}))
}

func TestMsgFeeParamKeyTable(t *testing.T) {
	keyTable := ParamKeyTable()
	require.Panics(t, func() {
		keyTable.RegisterType(paramtypes.NewParamSetPair(ParamStoreKeyFloorGasPrice, sdk.Coin{
			Denom:  "nhash",
			Amount: sdk.NewInt(5000),
		}, validateCoinParam))
	})
}

func TestDefault(t *testing.T) {
	metadataData := DefaultParams()
	require.Equal(t, DefaultFloorGasPrice, metadataData.FloorGasPrice)
}
