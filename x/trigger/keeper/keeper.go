package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)
// Keeper manages the state and logic for the trigger module.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	router   baseapp.IMsgServiceRouter
}
// NewKeeper creates a new Keeper instance for the trigger module.
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
// Logger returns the module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
