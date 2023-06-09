package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// MsgServiceHandler defines a function type which handles Msg service message.
type MsgServiceHandler = func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error)

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
