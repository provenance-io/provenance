package keeper

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/provenance-io/provenance/x/name/types"
)

// InitGenesis creates the initial genesis state for the name module.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	// Set module parameters
	k.SetParams(ctx, data.Params)
	for _, record := range data.Bindings {
		addr, err := sdk.AccAddressFromBech32(record.Address)
		if err != nil {
			panic(err)
		}
		if err := k.SetNameRecord(ctx, record.Name, addr, record.Restricted); err != nil {
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

	rng := (&collections.Range[[]byte]{}).
		StartInclusive(types.NameKeyPrefix).
		EndExclusive(storetypes.PrefixEndBytes(types.NameKeyPrefix))

	// Iterate through all name records
	err := k.NameRecords.Walk(ctx, rng, func(key []byte, record types.NameRecord) (bool, error) {
		records = append(records, record)
		return false, nil
	})

	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(params, records)
}
