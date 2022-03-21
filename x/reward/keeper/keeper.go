package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	epochkeeper "github.com/provenance-io/provenance/x/epoch/keeper"
	"github.com/provenance-io/provenance/x/reward/types"
)

const StoreKey = types.ModuleName

type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         codec.BinaryCodec
	epochKeeper epochkeeper.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key sdk.StoreKey,
	epochKeeper epochkeeper.Keeper,
) Keeper {
	return Keeper{
		storeKey:    key,
		cdc:         cdc,
		epochKeeper: epochKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
