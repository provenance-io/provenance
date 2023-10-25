package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/provenance-io/provenance/x/ibcratelimit/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
) Keeper {
	return Keeper{
		storeKey:   key,
		cdc:        cdc,
		paramSpace: paramSpace,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	// This was previously done via i.paramSpace.GetParamSet(ctx, &params). That will
	// panic if the params don't exist. This is a workaround to avoid that panic.
	// Params should be refactored to just use a raw kvstore.
	empty := types.Params{}
	for _, pair := range params.ParamSetPairs() {
		k.paramSpace.GetIfExists(ctx, pair.Key, pair.Value)
	}
	if params == empty {
		return types.DefaultParams()
	}
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
