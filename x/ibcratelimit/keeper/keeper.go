package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// Keeper for the ibcratelimit module
type Keeper struct {
	storeKey           storetypes.StoreKey
	cdc                codec.BinaryCodec
	PermissionedKeeper ibcratelimit.PermissionedKeeper
}

// NewKeeper Creates a new Keeper for the module.
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	permissionedKeeper ibcratelimit.PermissionedKeeper,
) Keeper {
	return Keeper{
		storeKey:           key,
		cdc:                cdc,
		PermissionedKeeper: permissionedKeeper,
	}
}

// Logger Creates a new logger for the module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcratelimit.ModuleName)
}

// GetParams Gets the params for the module.
func (k Keeper) GetParams(ctx sdk.Context) (params ibcratelimit.Params, err error) {
	store := ctx.KVStore(k.storeKey)
	key := ibcratelimit.ParamsKey
	bz := store.Get(key)
	if len(bz) == 0 {
		return ibcratelimit.Params{}, nil
	}
	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}

// SetParams Sets the params for the module.
func (k Keeper) SetParams(ctx sdk.Context, params ibcratelimit.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(ibcratelimit.ParamsKey, bz)
}

// GetContractAddress Gets the current value of the module's contract address.
func (k Keeper) GetContractAddress(ctx sdk.Context) (contract string) {
	params, _ := k.GetParams(ctx)
	return params.ContractAddress
}

// ContractConfigured Checks if the contract has been configured for the module.
func (k Keeper) ContractConfigured(ctx sdk.Context) bool {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false
	}
	return params.ContractAddress != ""
}
