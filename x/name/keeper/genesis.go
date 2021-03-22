package keeper

import (
	types "github.com/provenance-io/provenance/x/name/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis creates the initial genesis state for the name module.
func (keeper Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	keeper.SetParams(ctx, data.Params)
	for _, record := range data.Bindings {
		addr, err := sdk.AccAddressFromBech32(record.Address)
		if err != nil {
			panic(err)
		}
		if err := keeper.SetNameRecord(ctx, record.Name, addr, record.Restricted); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the current keeper state of the name module.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := keeper.GetParams(ctx)
	// Genesis state data structure.
	records := types.NameRecords{}
	// Callback func that adds records to genesis state.
	appendToRecords := func(record types.NameRecord) error {
		records = append(records, record)
		return nil
	}
	// Collect and return genesis state.
	if err := keeper.IterateRecords(ctx, types.NameKeyPrefix, appendToRecords); err != nil {
		panic(err)
	}
	return types.NewGenesisState(params, records)
}
