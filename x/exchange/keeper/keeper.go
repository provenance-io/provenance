package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// TODO[1658]: Finish the Keeper struct.
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	// TODO[1658]: Finish NewKeeper.
	rv := Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
	return rv
}
