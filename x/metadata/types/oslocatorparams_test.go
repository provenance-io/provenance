package types

import (
	"testing"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
)

func TestCreateParams(t *testing.T) {
	osParam := NewOSLocatorParams(3000)
	require.Equal(t, int(3000), int(osParam.MaxUriLength))
}

func TestCreateParamSet(t *testing.T) {
	osParam := NewOSLocatorParams(3000)
	paramsetPair := osParam.ParamSetPairs()
	require.Equal(t, 1, len(paramsetPair))
}

func TestValidateMaxURILengthI(t *testing.T) {
	require.NoError(t, validateMaxURILength(uint32(maxURLLength)))
}

func TestOSParamKeyTable(t *testing.T) {
	keyTable := OSParamKeyTable()
	require.Panics(t, func() {
		keyTable.RegisterType(paramtypes.NewParamSetPair(ParamStoreKeyMaxValueLength, "5000", validateMaxURILength))
	})
}

func TestDefault(t *testing.T) {
	metadataData := DefaultOSLocatorParams()
	require.Equal(t, 2048, int(metadataData.MaxUriLength))
}
