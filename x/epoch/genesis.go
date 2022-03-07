package epochs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/epoch/keeper"
	"github.com/provenance-io/provenance/x/epoch/types"

)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// set epoch info from genesis
	for _, epoch := range genState.Epochs {
		// Initialize empty epoch values via Cosmos SDK
		if epoch.StartHeight == ctx.BlockHeight() {
			epoch.StartHeight = ctx.BlockHeight()
		}

		epoch.CurrentEpochStartHeight = ctx.BlockHeight()

		k.SetEpochInfo(ctx, epoch)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis(ctx)
	genesis.Epochs = k.AllEpochInfos(ctx)
	return genesis
}
