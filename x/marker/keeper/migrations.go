package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v042 "github.com/provenance-io/provenance/x/marker/legacy/v042"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m *Migrator) Migrate1to2(ctx sdk.Context) error {
	ctx.Logger().Info("Migrating Marker Module from Version 1 to 2")
	err := v042.MigrateMarkerAddressKeys(ctx, m.keeper.storeKey)
	if err != nil {
		return err
	}
	err = v042.MigrateMarkerPermissions(ctx, m.keeper)
	ctx.Logger().Info("Finished Migrating Marker Module from Version 1 to 2")
	return err
}
