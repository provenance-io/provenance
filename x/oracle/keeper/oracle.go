package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/oracle/types"
)

// SetTrigger Sets the trigger in the store.
func (k Keeper) SetOracleContract(ctx sdk.Context, contract sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetContractStoreKey(), contract)
}

// GetTrigger Gets a trigger from the store by id.
func (k Keeper) GetOracleContract(ctx sdk.Context) (acc sdk.AccAddress, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractStoreKey()
	acc = store.Get(key)
	if len(acc) == 0 {
		return sdk.AccAddress{}, types.ErrContractAddressDoesNotExist
	}

	return acc, err
}
