package keeper

import (
	"context"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// A WrappedGovKeeper implements the sanction.GovKeeper interface, allowing for
// mocking of some of the stuff that the gov keeper keeps in fields now.
type WrappedGovKeeper struct {
	Keeper *govkeeper.Keeper
}

// WrapGovKeeper creates a new WrappedGovKeeper around the provided keeper.
func WrapGovKeeper(keeper *govkeeper.Keeper) *WrappedGovKeeper {
	return &WrappedGovKeeper{Keeper: keeper}
}

func (w WrappedGovKeeper) GetProposal(ctx context.Context, propID uint64) *govv1.Proposal {
	prop, err := w.Keeper.Proposals.Get(ctx, propID)
	if err != nil {
		return nil
	}
	return &prop
}
