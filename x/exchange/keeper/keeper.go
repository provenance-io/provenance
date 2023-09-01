package keeper

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// TODO[1658]: Finish the Keeper struct.
	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	// TODO[1658]: Finish NewKeeper.
	rv := Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
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

// getAllCoins gets all the coin entries from the store with the given prefix.
// The denom comes from the part of the key after the prefix, and the amount comes from the values.
func getAllCoins(store sdk.KVStore, pre []byte) []sdk.Coin {
	// Using a prefixed store here so that iter.Key() doesn't contain the prefix.
	pStore := prefix.NewStore(store, pre)
	iter := pStore.Iterator(nil, nil)
	defer iter.Close()

	var coins []sdk.Coin
	for ; iter.Valid(); iter.Next() {
		denom := string(iter.Key())
		value := string(iter.Value())
		amt, ok := sdkmath.NewIntFromString(value)
		if !ok {
			continue
		}
		coins = append(coins, sdk.Coin{Denom: denom, Amount: amt})
	}

	return coins
}
