package keeper

import (
	"cosmossdk.io/core/codec"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

// Keeper defines the mymodule keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper returns a new mymodule Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// SetValue stores a value with a given key.
func (k Keeper) SetValue(ctx sdk.Context, key string, value string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(ledger.ModuleName+":"))
	store.Set([]byte(key), []byte(value))
}

// GetValue retrieves a value by key.
func (k Keeper) GetValue(ctx sdk.Context, key string) string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(ledger.ModuleName+":"))
	bz := store.Get([]byte(key))
	return string(bz)
}
