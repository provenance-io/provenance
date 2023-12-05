package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	router   baseapp.IMsgServiceRouter
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	router baseapp.IMsgServiceRouter,
) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
		router:   router,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
