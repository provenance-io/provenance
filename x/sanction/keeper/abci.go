package keeper

import (
	"math"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlock deletes the temporary entries for all proposals that no longer exist.
// This can happen, e.g., when a proposal is cancelled, since the gov module doesn't call any hooks in that case.
func (k Keeper) BeginBlock(ctx sdk.Context) {
	for _, govPropID := range k.getTempEntryGovPropIDs(ctx) {
		if k.govKeeper.GetProposal(ctx, govPropID) == nil {
			k.DeleteGovPropTempEntries(ctx, govPropID)
		}
	}
}

// getTempEntryGovPropIDs gets the distinct gov prop ids that have temporary entries (in ascending order).
// Rather than reading every index entry, this seeks past each gov prop id once it's been found,
// so the number of reads is proportional to the number of gov props (rather than number of addresses).
func (k Keeper) getTempEntryGovPropIDs(ctx sdk.Context) []uint64 {
	store := ctx.KVStore(k.storeKey)
	start := CreateProposalTempIndexPrefix(nil)
	end := storetypes.PrefixEndBytes(start)

	var rv []uint64
	for {
		govPropID, found := getFirstIndexGovPropID(store, start, end)
		if !found {
			return rv
		}
		rv = append(rv, govPropID)
		if govPropID == math.MaxUint64 {
			return rv
		}
		nextID := govPropID + 1
		start = CreateProposalTempIndexPrefix(&nextID)
	}
}

// getFirstIndexGovPropID gets the gov prop id of the first index entry in the provided range.
// Returns false if there aren't any entries in that range.
func getFirstIndexGovPropID(store storetypes.KVStore, start, end []byte) (uint64, bool) {
	iter := store.Iterator(start, end)
	defer iter.Close() //nolint:errcheck // ignoring close error on iterator: not critical for this context.
	if !iter.Valid() {
		return 0, false
	}
	govPropID, _ := ParseProposalTempIndexKey(iter.Key())
	return govPropID, true
}
