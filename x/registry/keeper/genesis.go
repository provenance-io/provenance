package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/provenance-io/provenance/x/registry/types"
)

func (k Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) {
	for _, entry := range state.Entries {
		if err := k.Registry.Set(ctx, entry.Key.CollKey(), entry); err != nil {
			panic(err) // Genesis should not fail
		}
	}
	for _, change := range state.PendingRoleChanges {
		if err := k.PendingRoleChanges.Set(ctx, change.Id, change); err != nil {
			panic(err) // Genesis should not fail
		}
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	genesis := &types.GenesisState{}

	err := k.Registry.Walk(ctx, nil, func(_ collections.Pair[string, string], value types.RegistryEntry) (stop bool, err error) {
		genesis.Entries = append(genesis.Entries, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	err = k.PendingRoleChanges.Walk(ctx, nil, func(_ string, value types.PendingRoleChange) (stop bool, err error) {
		genesis.PendingRoleChanges = append(genesis.PendingRoleChanges, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return genesis
}
