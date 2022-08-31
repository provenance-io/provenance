package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/expiration/types"
)

// InitGenesis creates the initial genesis state for the expiration module
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)

	for _, expiration := range data.Expirations {
		err := k.SetExpiration(ctx, expiration)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the current keeper state of the name module
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	// Genesis state data structure.
	records := make([]types.Expiration, 0)
	// Callback func that adds records to genesis state.
	expirationHandler := func(expiration types.Expiration) error {
		records = append(records, expiration)
		return nil
	}
	// Collect and return genesis state.
	if err := k.IterateExpirations(ctx, types.ModuleAssetKeyPrefix, expirationHandler); err != nil {
		panic(err)
	}
	return types.NewGenesisState(params, records)
}
