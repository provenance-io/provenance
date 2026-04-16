package keeper

import (
	"cosmossdk.io/log"

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

// Migrate2to3 migrates state from consensus version 2 to 3.
//
// This migration converts the attribute module from raw KV-store access
// to cosmossdk.io/collections.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	logger := m.keeper.Logger(ctx)
	logger.Info(
		"attribute: Migrate2to3 — KV to Collections is a no-op (byte-identical on-disk layout)",
	)
	return nil
}

// Compile-time assurance the keeper exposes a Logger that returns log.Logger.
var _ interface {
	Logger(sdk.Context) log.Logger
} = Keeper{}
