package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	ctx.Logger().Info("Migrating Reward Module from Version 1 to 2")
	ctx.Logger().Info("Finished Migrating Reward Module from Version 1 to 2")
	return nil
}
