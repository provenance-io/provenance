package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/provenance-io/provenance/x/name/types"
)

// InitGenesis creates the initial genesis state for the name module.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	k.SetParams(ctx, data.Params)
	for _, record := range data.Bindings {
		// Create record directly without normalization checks
		if err := k.nameRecords.Set(ctx, record.Name, types.NameRecord{
			Name:       record.Name,
			Address:    record.Address,
			Restricted: record.Restricted,
		}); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the current keeper state of the name module.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Get module parameters
	params := k.GetParams(ctx)
	// Collect all name records
	records := make(types.NameRecords, 0)

	// Iterate through all name records
	err := k.nameRecords.Walk(ctx, nil, func(name string, record types.NameRecord) (bool, error) {
		records = append(records, record)
		return false, nil
	})

	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(params, records)
}
