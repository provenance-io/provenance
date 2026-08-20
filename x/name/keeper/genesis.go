package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/provenance-io/provenance/x/name/types"
)

// InitGenesis creates the initial genesis state for the name module.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	k.SetParams(ctx, data.Params)
	for _, record := range data.Bindings {
		addr, err := sdk.AccAddressFromBech32(record.Address)
		if err != nil {
			panic(fmt.Errorf("invalid address %q for name %q: %w", record.Address, record.Name, err))
		}
		if err = k.SetNameRecord(ctx, record.Name, addr, record.Restricted); err != nil {
			panic(fmt.Errorf("could not bind name %q: %w", record.Name, err))
		}
	}
}

// ExportGenesis exports the current keeper state of the name module.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	records := make(types.NameRecords, 0)

	if err := k.walkRecords(ctx, func(record types.NameRecord) error {
		records = append(records, record)
		return nil
	}); err != nil {
		panic(err)
	}

	// Records are sorted by name to keep exported genesis stable and easy to diff.
	sort.Slice(records, func(i, j int) bool { return records[i].Name < records[j].Name })
	return types.NewGenesisState(params, records)
}
