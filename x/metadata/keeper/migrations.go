package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v042 "github.com/provenance-io/provenance/x/metadata/legacy/v042"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2 to convert attribute keys from 20 to 32 length
func (m *Migrator) Migrate1to2(ctx sdk.Context) error {
	ctx.Logger().Info("Migrating Metadata Module from Version 1 to 2")
	return v042.MigrateAddresses(ctx, m.keeper.storeKey)
}
