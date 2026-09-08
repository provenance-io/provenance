package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *ibcratelimit.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &ibcratelimit.GenesisState{
		Params: params,
	}
}

// InitGenesis new ibcratelimit genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *ibcratelimit.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(fmt.Errorf("failed to set ibcratelimit params during InitGenesis: %w", err))
	}
}
