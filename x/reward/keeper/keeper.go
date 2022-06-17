package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/provenance-io/provenance/x/reward/types"
)

const StoreKey = types.ModuleName

type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           codec.BinaryCodec
	stakingKeeper types.StakingKeeper
	govKeeper     *govkeeper.Keeper
	bankKeeper    bankkeeper.Keeper
	authkeeper    authkeeper.AccountKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key sdk.StoreKey,
	stakingKeeper types.StakingKeeper,
	govKeeper *govkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	authKeeper authkeeper.AccountKeeper,
) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		stakingKeeper: stakingKeeper,
		govKeeper:     govKeeper,
		bankKeeper:    bankKeeper,
		authkeeper:    authKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetActionDelegate sets the action delegate in the keeper
func (k Keeper) SetActionDelegate(ctx sdk.Context, actionDelegate types.ActionDelegate) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&actionDelegate)
	store.Set(types.GetActionDelegateKey(), bz)
}

// GetActionDelegate returns a action delegate
func (k Keeper) GetActionDelegate(ctx sdk.Context) (actionDelegate types.ActionDelegate, err error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetActionDelegateKey())
	if len(bz) == 0 {
		return actionDelegate, nil
	}
	err = k.cdc.Unmarshal(bz, &actionDelegate)
	return actionDelegate, err
}

// SetActionTransferDelegations sets the action transfer delegations in the keeper
func (k Keeper) SetActionTransferDelegations(ctx sdk.Context, actionTransferDelegations types.ActionTransferDelegations) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&actionTransferDelegations)
	store.Set(types.GetActionTransferDelegationsKey(), bz)
}

// GetActionTransferDelegations returns a action transfer delegations
func (k Keeper) GetActionTransferDelegations(ctx sdk.Context) (actionTransferDelegations types.ActionTransferDelegations, err error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetActionTransferDelegationsKey())
	if len(bz) == 0 {
		return actionTransferDelegations, nil
	}
	err = k.cdc.Unmarshal(bz, &actionTransferDelegations)
	return actionTransferDelegations, err
}
