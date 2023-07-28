package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/oracle/types"
	icq "github.com/strangelove-ventures/async-icq/v6/keeper"
)

type Keeper struct {
	storeKey  storetypes.StoreKey
	cdc       codec.BinaryCodec
	icqKeeper icq.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	icqKeeper icq.Keeper,
) Keeper {
	return Keeper{
		storeKey:  key,
		cdc:       cdc,
		icqKeeper: icqKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
