package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/marker/types"
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

// Migrate3to4 moves access_control out of every marker account into the markerPerms store.
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	m.keeper.Logger(ctx).Info("Migrating marker permissions to dedicated storage (marker v3->v4).")

	// Collect addresses first — no store mutation while walking the index collection.
	var addrs []sdk.AccAddress
	m.keeper.IterateMarkers(ctx, func(marker types.MarkerAccountI) bool {
		addrs = append(addrs, marker.GetAddress())
		return false
	})

	var moved int
	for _, addr := range addrs {
		// Raw auth read — bypasses GetMarker, which now clears AccessControl.
		acct := m.keeper.authKeeper.GetAccount(ctx, addr)
		ma, ok := acct.(*types.MarkerAccount)
		if !ok || len(ma.AccessControl) == 0 {
			continue
		}
		for _, g := range ma.AccessControl {
			gAddr, err := sdk.AccAddressFromBech32(g.Address)
			if err != nil {
				return fmt.Errorf("marker %s: invalid grant address %q: %w", ma.Denom, g.Address, err)
			}
			if err := m.keeper.SetAccess(ctx, addr, gAddr, g.Permissions); err != nil {
				return err
			}
		}
		ma.AccessControl = nil
		m.keeper.authKeeper.SetAccount(ctx, ma)
		moved++
	}
	m.keeper.Logger(ctx).Info("Marker permissions migration complete.", "markers_migrated", moved)
	return nil
}
