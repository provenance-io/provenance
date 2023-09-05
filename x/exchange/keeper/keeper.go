package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/provenance-io/provenance/x/exchange"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	accountKeeper exchange.AccountKeeper

	// TODO[1658]: Finish the Keeper struct.
	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, accountKeeper exchange.AccountKeeper) Keeper {
	// TODO[1658]: Finish NewKeeper.
	rv := Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		authority:     authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
	return rv
}

// GetAuthority gets the address (as bech32) that has governance authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// getAllKeys gets all the keys in the store with the given prefix.
func getAllKeys(store sdk.KVStore, pre []byte) [][]byte {
	// Using a prefix iterator so that iter.Key() is the whole key (including the prefix).
	iter := sdk.KVStorePrefixIterator(store, pre)
	defer iter.Close()

	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}

	return keys
}

// deleteAll deletes all keys that have the given prefix.
func deleteAll(store sdk.KVStore, pre []byte) {
	keys := getAllKeys(store, pre)
	for _, key := range keys {
		store.Delete(key)
	}
}

// iterate iterates over all the entries in the store with the given prefix.
// The key provided to the callback will NOT have the provided prefix; it will be everything after it.
// The callback should return false to continue iteration, or true to stop.
func iterate(store sdk.KVStore, pre []byte, cb func(key, value []byte) bool) {
	// Using an open iterator on a prefixed store here so that iter.Key() doesn't contain the prefix.
	pStore := prefix.NewStore(store, pre)
	iter := pStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if cb(iter.Key(), iter.Value()) {
			break
		}
	}
}

// getStore gets the store for the exchange module.
func (k Keeper) getStore(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(k.storeKey)
}

// iterate iterates over all the entries in the store with the given prefix.
// The key provided to the callback will NOT have the provided prefix; it will be everything after it.
// The callback should return false to continue iteration, or true to stop.
func (k Keeper) iterate(ctx sdk.Context, pre []byte, cb func(key, value []byte) bool) {
	iterate(k.getStore(ctx), pre, cb)
}
