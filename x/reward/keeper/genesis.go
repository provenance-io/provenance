package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState()
}

// InitGenesis new msgfees genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {

}
