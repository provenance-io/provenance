package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold"
)

// InitGenesis loads the provided GenesisState into the state store.
// Panics if there's an error.
func (k Keeper) InitGenesis(origCtx sdk.Context, genState *escrow.GenesisState) {
	if genState == nil {
		return
	}

	// We don't want the events from this, so use a context with a throw-away event manager.
	ctx := origCtx.WithEventManager(sdk.NewEventManager())

	for i, ae := range genState.Escrows {
		// Not worrying about wrapping any bech32 error because I'm assuming
		// genState.Validate() was called before this.
		addr := sdk.MustAccAddressFromBech32(ae.Address)
		if err := k.AddEscrow(ctx, addr, ae.Amount); err != nil {
			panic(fmt.Errorf("escrows[%d]: %w", i, err))
		}
	}
}

// ExportGenesis creates a GenesisState from the current state store.
func (k Keeper) ExportGenesis(ctx sdk.Context) *escrow.GenesisState {
	var err error
	rv := &escrow.GenesisState{}

	rv.Escrows, err = k.GetAllAccountEscrows(ctx)
	if err != nil {
		panic(err)
	}

	return rv
}
