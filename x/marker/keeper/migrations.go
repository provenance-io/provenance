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

// Migrate2to3 performs the byte-identity migration from raw KV to collections.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	m.keeper.Logger(sdkCtx).Info("migrating marker module from version 2 to 3 (no-op, byte-identity)")
	return nil
}
