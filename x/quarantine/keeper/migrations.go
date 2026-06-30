package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/quarantine"
)

// Migrator handles in-place store migrations for the quarantine module.
type Migrator struct {
	keeper     Keeper
	bankKeeper quarantine.BankKeeper
}

// NewMigrator returns a Migrator for the quarantine module.
// The bankKeeper is passed in because the deactivated keeper no longer holds one.
func NewMigrator(k Keeper, bankKeeper quarantine.BankKeeper) Migrator {
	return Migrator{keeper: k, bankKeeper: bankKeeper}
}

// Migrate1to2 releases all quarantined funds back to their recipients, removes the records,
// and opts everyone out. This runs when the quarantine module is deactivated.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return m.keeper.ReleaseAllQuarantinedFunds(ctx, m.bankKeeper)
}
