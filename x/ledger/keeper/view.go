package keeper

import (
	"context"
	"errors"
	"slices"
	"sort"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/helper"
	"github.com/provenance-io/provenance/x/ledger/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

// GetLedger retrieves a ledger by its NFT address.
//
// Parameters:
//   - ctx: The SDK context
//   - nftAddress: The NFT address to look up the ledger for
//
// Returns:
//   - *ledger.Ledger: A pointer to the found ledger, or nil if not found
//   - error: Any error that occurred during retrieval, or nil if successful
//
// Behavior:
//   - Returns (nil, nil) if the ledger is not found
//   - Returns (nil, err) if an error occurs during retrieval
//   - Returns (&ledger, nil) if the ledger is found successfully
//   - The returned ledger will have its NftAddress field set to the provided nftAddress
func (k Keeper) GetLedger(ctx sdk.Context, key *ledger.LedgerKey) (*ledger.Ledger, error) {
	keyStr := key.String()

	// Lookup the NFT address in the ledger
	l, err := k.Ledgers.Get(ctx, keyStr)
	if err != nil {
		// Eat the not found error as it is expected, and return nil.
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}

		// Otherwise, return the error.
		return nil, err
	}

	// The NFT address isn't stored in the ledger, so we add it back in.
	l.Key = key
	return &l, nil
}

func (k Keeper) RequireGetLedger(ctx sdk.Context, lk *ledger.LedgerKey) (*ledger.Ledger, error) {
	ledger, err := k.GetLedger(ctx, lk)
	if err != nil {
		return nil, err
	}
	if ledger == nil {
		return nil, types.NewLedgerCodedError(types.ErrCodeNotFound, "ledger")
	}

	return ledger, nil
}

func (k Keeper) HasLedger(ctx sdk.Context, key *ledger.LedgerKey) bool {
	keyStr := key.String()

	has, _ := k.Ledgers.Has(ctx, keyStr)
	return has
}

func (k Keeper) ListLedgerEntries(ctx context.Context, key *ledger.LedgerKey) ([]*ledger.LedgerEntry, error) {
	keyStr := key.String()

	// Garbage in, garbage out.
	if !k.HasLedger(sdk.UnwrapSDKContext(ctx), key) {
		return nil, nil
	}

	// Get all entries for the ledger.
	prefix := collections.NewPrefixedPairRange[string, string](keyStr)
	iter, err := k.LedgerEntries.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	// Iterate through all entries for the ledger.
	var entries []*ledger.LedgerEntry
	for ; iter.Valid(); iter.Next() {
		le, err := iter.Value()
		if err != nil {
			return nil, err
		}
		entries = append(entries, &le)
	}

	// Sort the entries by effective date and sequence number.
	slices.SortFunc(entries, func(a, b *ledger.LedgerEntry) int {
		return a.Compare(b)
	})

	return entries, nil
}

// GetLedgerEntry retrieves a ledger entry by its correlation ID for a specific NFT address
func (k Keeper) GetLedgerEntry(ctx context.Context, key *ledger.LedgerKey, correlationID string) (*ledger.LedgerEntry, error) {
	if !k.HasLedger(sdk.UnwrapSDKContext(ctx), key) {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeNotFound, "ledger")
	}

	entries, err := k.ListLedgerEntries(ctx, key)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.CorrelationId == correlationID {
			return entry, nil
		}
	}

	// If we get here, the entry was not found
	return nil, nil
}

func (k Keeper) RequireGetLedgerEntry(ctx sdk.Context, lk *ledger.LedgerKey, correlationID string) (*ledger.LedgerEntry, error) {
	ledgerEntry, err := k.GetLedgerEntry(ctx, lk, correlationID)
	if err != nil {
		return nil, err
	}
	if ledgerEntry == nil {
		return nil, types.NewLedgerCodedError(types.ErrCodeNotFound, "ledger entry")
	}

	return ledgerEntry, nil
}

// GetBalancesAsOf returns the principal, interest, and other balances as of a specific effective date
func (k Keeper) GetBalancesAsOf(ctx context.Context, key *ledger.LedgerKey, asOfDate time.Time) (*ledger.BucketBalances, error) {
	asOfDateInt := helper.DaysSinceEpoch(asOfDate.UTC())

	// Get all ledger entries for this NFT
	entries, err := k.ListLedgerEntries(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeNotFound, "ledger entries")
	}

	// Map of bucket name to list of bucket balances.
	bucketBalances := make(map[int32]*ledger.BucketBalance, 0)

	// Calculate balances up to the specified date
	for _, entry := range entries {
		// Skip entries after the asOfDate
		if entry.EffectiveDate > asOfDateInt {
			break
		}

		// We just find the latest entry as of the asOfDate, and set the balances.
		for _, bucketBalance := range entry.BalanceAmounts {
			bucketBalances[bucketBalance.BucketTypeId] = bucketBalance
		}
	}

	bucketBalancesList := make([]*ledger.BucketBalance, 0)
	for _, balance := range bucketBalances {
		bucketBalancesList = append(bucketBalancesList, balance)
	}

	// Not sure this really helps sort anything since the id's are arbitrary, but at least we're consistent.
	sort.Slice(bucketBalancesList, func(i, j int) bool {
		return bucketBalancesList[i].BucketTypeId < bucketBalancesList[j].BucketTypeId
	})

	return &ledger.BucketBalances{
		BucketBalances: bucketBalancesList,
	}, nil
}

func (k Keeper) GetLedgerClass(ctx context.Context, ledgerClassId string) (*ledger.LedgerClass, error) {
	ledgerClass, err := k.LedgerClasses.Get(ctx, ledgerClassId)
	if err != nil {
		return nil, err
	}
	return &ledgerClass, nil
}

func (k Keeper) GetLedgerClassEntryTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassEntryType, error) {
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassId)
	iter, err := k.LedgerClassEntryTypes.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	entryTypes := make([]*ledger.LedgerClassEntryType, 0)
	for ; iter.Valid(); iter.Next() {
		entryType, err := iter.Value()
		if err != nil {
			return nil, err
		}

		entryTypes = append(entryTypes, &entryType)
	}

	return entryTypes, nil
}

func SliceToMap[T any, K comparable](list []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(list))
	for _, item := range list {
		result[keyFn(item)] = item
	}
	return result
}

func (k Keeper) GetLedgerClassStatusTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassStatusType, error) {
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassId)
	iter, err := k.LedgerClassStatusTypes.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	statusTypes := make([]*ledger.LedgerClassStatusType, 0)
	for ; iter.Valid(); iter.Next() {
		statusType, err := iter.Value()
		if err != nil {
			return nil, err
		}
		statusTypes = append(statusTypes, &statusType)
	}
	return statusTypes, nil
}

func (k Keeper) GetLedgerClassBucketTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassBucketType, error) {
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassId)
	iter, err := k.LedgerClassBucketTypes.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	bucketTypes := make([]*ledger.LedgerClassBucketType, 0)
	for ; iter.Valid(); iter.Next() {
		bucketType, err := iter.Value()
		if err != nil {
			return nil, err
		}
		bucketTypes = append(bucketTypes, &bucketType)
	}
	return bucketTypes, nil
}
