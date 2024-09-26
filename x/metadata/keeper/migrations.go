package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

func (m Migrator) Migrate3to4(_ sdk.Context) error {
	return nil // TODO[2137]: Write Migrate3to4
}
