package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

const StoreKey = types.ModuleName

// MessageRouter ADR 031 request type routing
// https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-031-msg-service.md
type MessageRouter interface {
	Handler(msg sdk.Msg) MsgServiceHandler
}

// MsgServiceHandler defines a function type which handles Msg service message.
type MsgServiceHandler = func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	router   MessageRouter
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	router MessageRouter,
) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
		router:   router,
	}
}

func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
