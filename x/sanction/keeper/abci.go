package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlock deletes the temporary entries for all proposals that no longer exist.
// This can happen, e.g., when a proposal is cancelled, since the gov module doesn't call any hooks in that case.
func (k Keeper) BeginBlock(ctx sdk.Context) {
	// The index entries are ordered by gov prop id, so we only need to compare to the previous one to dedup.
	var govPropIDs []uint64
	k.IterateProposalIndexEntries(ctx, nil, func(govPropID uint64, _ sdk.AccAddress) bool {
		if len(govPropIDs) == 0 || govPropIDs[len(govPropIDs)-1] != govPropID {
			govPropIDs = append(govPropIDs, govPropID)
		}
		return false
	})

	for _, govPropID := range govPropIDs {
		if k.govKeeper.GetProposal(ctx, govPropID) == nil {
			k.DeleteGovPropTempEntries(ctx, govPropID)
		}
	}
}
