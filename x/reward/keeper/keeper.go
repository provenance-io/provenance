package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/provenance-io/provenance/x/reward/types"
)

const StoreKey = types.ModuleName

type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	stakingKeeper types.StakingKeeper
	govKeeper     *govkeeper.Keeper
	bankKeeper    bankkeeper.Keeper
	authkeeper    authkeeper.AccountKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
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
