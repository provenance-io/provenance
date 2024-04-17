package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// GetParams returns the total set of distribution parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	params = types.DefaultParams() // Initialize with defaults

	bz := store.Get(types.ParamStoreKey)
	if bz != nil {
		k.cdc.MustUnmarshal(bz, &params) // Deserialize parameters
	}
	return params
}

// SetParams sets the account parameters to the store.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params) // Serialize parameters
	store.Set(types.ParamStoreKey, bz)
}

// GetFloorGasPrice returns the current minimum gas price in sdk.Coin used in calculations for charging additional fees
func (k Keeper) GetFloorGasPrice(ctx sdk.Context) sdk.Coin {
	params := k.GetParams(ctx)
	return params.FloorGasPrice
}

// GetNhashPerUsdMil returns the current nhash amount per usd mil.
//
// Conversions:
//   - x nhash/usd-mil = 1,000,000/x usd/hash
//   - y usd/hash = 1,000,000/y nhash/usd-mil
//
// Examples:
//   - 40,000,000 nhash/usd-mil = 1,000,000/40,000,000 usd/hash = $0.025/hash,
//   - $0.040/hash = 1,000,000/0.040 nhash/usd-mil = 25,000,000 nhash/usd-mil
func (k Keeper) GetNhashPerUsdMil(ctx sdk.Context) uint64 {
	params := k.GetParams(ctx)
	return params.NhashPerUsdMil
}

// GetConversionFeeDenom returns the conversion fee denom
func (k Keeper) GetConversionFeeDenom(ctx sdk.Context) string {
	params := k.GetParams(ctx)
	return params.ConversionFeeDenom
}
