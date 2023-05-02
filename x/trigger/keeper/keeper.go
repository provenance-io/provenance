package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

const StoreKey = types.ModuleName

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
	}
}
