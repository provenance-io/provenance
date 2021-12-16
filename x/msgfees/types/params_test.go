package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateParams(t *testing.T) {
	msgFeeParam := NewParams(true, 2000)
	require.Equal(t, int(2000), int(msgFeeParam.FloorGasPrice))
}

func TestCreateParamSet(t *testing.T) {
	msgFeeParam := NewParams(true, 3000)
	paramsetPair := msgFeeParam.ParamSetPairs()
	require.Equal(t, 2, len(paramsetPair))
}

func TestValidateMinGasPriceParamI(t *testing.T) {
	require.NoError(t, validateIntParam(uint32(1905)))
}

func TestOSParamKeyTable(t *testing.T) {
	keyTable := ParamKeyTable()
	require.Panics(t, func() {
		keyTable.RegisterType(paramtypes.NewParamSetPair(ParamStoreKeyFloorGasPrice, "5000", validateIntParam))
		keyTable.RegisterType(paramtypes.NewParamSetPair(ParamStoreKeyEnableGovernance, "true", validateEnableGovernance))
	})
}

func TestDefault(t *testing.T) {
	metadataData := DefaultParams()
	require.Equal(t, 1905, int(metadataData.FloorGasPrice))
}
