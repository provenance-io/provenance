package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ibcratelimit/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	storeKey       storetypes.StoreKey
	cdc            codec.BinaryCodec
	ContractKeeper *wasmkeeper.PermissionedKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	contractKeeper *wasmkeeper.PermissionedKeeper,
) Keeper {
	return Keeper{
		storeKey:       key,
		cdc:            cdc,
		ContractKeeper: contractKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.ParamsKey
	bz := store.Get(key)
	if len(bz) == 0 {
		return types.Params{}, nil
	}
	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
}

func (k Keeper) GetContractAddress(ctx sdk.Context) (contract string) {
	params, _ := k.GetParams(ctx)
	return params.ContractAddress
}

func (k Keeper) ContractConfigured(ctx sdk.Context) bool {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false
	}
	return params.ContractAddress != ""
}
