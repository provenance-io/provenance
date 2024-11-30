package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/nav"
)

// InitGenesis initializes the nav module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState *nav.GenesisState) {
	if genState == nil || len(genState.Navs) == 0 {
		return
	}
	err := k.SetNAVRecords(ctx, genState.Navs)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the nav module's genesis state.
func (k Keeper) ExportGenesis(ctx context.Context) *nav.GenesisState {
	return &nav.GenesisState{
		Navs: k.GetAllNAVRecords(ctx),
	}
}
