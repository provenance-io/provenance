package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/oracle/types"
)

// SetOracle Sets the oracle used by the module.
func (k Keeper) SetOracle(ctx sdk.Context, oracle sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetOracleStoreKey(), oracle)
}

// GetOracle Gets the oracle used by the module.
func (k Keeper) GetOracle(ctx sdk.Context) (oracle sdk.AccAddress, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOracleStoreKey()
	oracle = store.Get(key)
	if len(oracle) == 0 {
		return sdk.AccAddress{}, types.ErrMissingOracleAddress
	}

	return oracle, err
}
