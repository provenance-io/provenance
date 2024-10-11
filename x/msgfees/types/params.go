package types

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
)

// DefaultFloorGasPrice to differentiate between base fee and additional fee when additional fee is in same denom as default base denom i.e nhash
// cannot be a const unfortunately because it's a custom type.
func DefaultFloorGasPrice() sdk.Coin {
	return sdk.Coin{
		Amount: sdkmath.NewInt(pioconfig.GetProvenanceConfig().MsgFeeFloorGasPrice),
		Denom:  pioconfig.GetProvenanceConfig().MsgFloorDenom,
	}
}

var DefaultNhashPerUsdMil = uint64(25_000_000)

// NewParams creates a new parameter object
func NewParams(
	floorGasPrice sdk.Coin,
	nhashPerUsdMil uint64,
	conversionFeeDenom string,
) Params {
	return Params{
		FloorGasPrice:      floorGasPrice,
		NhashPerUsdMil:     nhashPerUsdMil,
		ConversionFeeDenom: conversionFeeDenom,
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(
		DefaultFloorGasPrice(),
		DefaultNhashPerUsdMil,
		pioconfig.GetProvenanceConfig().FeeDenom,
	)
}
