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
