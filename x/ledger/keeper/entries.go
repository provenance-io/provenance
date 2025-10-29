package keeper

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ledger/helper"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// AppendEntries adds multiple new entries to a ledger while maintaining proper sequencing
// and validation. This function handles the core logic for adding ledger entries.
//
// The function performs several critical operations:
// 1. Validates the ledger exists and is accessible
// 2. Checks posted dates are not in the future
// 3. Validates entry types against the ledger class configuration
// 4. Manages sequence numbers to prevent conflicts
// 5. Stores entries with proper correlation ID handling
//
// Parameters:
// - ctx: The SDK context for state operations
// - ledgerKey: The ledger identifier (asset class ID and NFT ID)
// - entries: Array of ledger entries to append
//
// Returns an error if validation fails or if entries cannot be saved.
func (k Keeper) AppendEntries(goCtx context.Context, ledgerKey *types.LedgerKey, entries []*types.LedgerEntry) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Retrieve the ledger to ensure it exists and get its configuration.
	ledger, err := k.RequireGetLedger(goCtx, ledgerKey)
	if err != nil {
		return err
	}

	// Get all existing entries for this NFT to check for conflicts and manage sequencing.
	existingEntries, err := k.ListLedgerEntries(goCtx, ledgerKey)
	if err != nil {
		return err
	}

	// Calculate the current block time in days since epoch for date validation.
	blockTimeDays := helper.DaysSinceEpoch(ctx.BlockTime().UTC())
	for _, newEntry := range entries {
		// Validate that posted dates are not in the future.
		// This prevents entries from being posted with future timestamps.
		if newEntry.PostedDate > blockTimeDays {
			return types.NewErrCodeInvalidField("posted_date", "cannot be in the future")
		}

		// Validate that the LedgerClassEntryType exists for this ledger class.
		// This ensures only valid entry types can be used for this ledger.
		var hasLedgerClassEntryType bool
		hasLedgerClassEntryType, err = k.LedgerClassEntryTypes.Has(goCtx, collections.Join(ledger.LedgerClassId, newEntry.EntryTypeId))
		if err != nil {
			return fmt.Errorf("error getting ledger class entry type for ledger class id %q and entry type %d", ledger.LedgerClassId, newEntry.EntryTypeId)
		}
		if !hasLedgerClassEntryType {
			return types.NewErrCodeInvalidField("entry_type_id", "entry type doesn't exist")
		}

		// Save the individual entry with proper sequencing and conflict resolution.
		existingEntries, err = k.saveNewEntry(goCtx, ledgerKey, existingEntries, newEntry)
		if err != nil {
			return err
		}
	}

	return nil
}

// saveNewEntry handles the individual entry storage logic with sequence number management
// and conflict resolution. This function ensures that entries are stored with proper
// sequencing and that no duplicate correlation IDs exist.
//
// This method is only designed for use by the AppendEntries method.
//
// Returns an updated list of existing entries.
// Returns an error if conflicts are detected or if storage fails.
func (k Keeper) saveNewEntry(goCtx context.Context, ledgerKey *types.LedgerKey, entries []*types.LedgerEntry, newEntry *types.LedgerEntry) ([]*types.LedgerEntry, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure the new entry is valid.
	if err := newEntry.Validate(); err != nil {
		return nil, err
	}

	// Get the string representation of the ledger key for use in key/value store operations.
	ledgerKeyStr := ledgerKey.String()
	// Make sure the entry doesn't already exist.
	newEntryKey := collections.Join(ledgerKeyStr, newEntry.CorrelationId)
	if found, err := k.LedgerEntries.Has(goCtx, newEntryKey); found || err != nil {
		return nil, types.NewErrCodeAlreadyExists(fmt.Sprintf("ledger entry with correlation id %q", newEntry.CorrelationId))
	}

	// Find entries with the same effective date for sequence number management.
	// Entries with the same effective date must have unique sequence numbers.
	var sameDateEntries []*types.LedgerEntry
	for _, entry := range entries {
		if entry.EffectiveDate == newEntry.EffectiveDate {
			sameDateEntries = append(sameDateEntries, entry)
		}
	}

	// If there are entries with the same date, check for sequence number conflicts
	// and manage the sequencing to maintain proper order.
	if len(sameDateEntries) > 0 {
		// Sort entries by sequence number to identify conflicts and gaps.
		slices.SortFunc(sameDateEntries, func(a, b *types.LedgerEntry) int {
			if a.Sequence < b.Sequence {
				return -1
			}
			if a.Sequence > b.Sequence {
				return 1
			}
			return 0
		})

		// Check if the new entry's sequence number conflicts with existing entries.
		// If a conflict is found, increment the sequence numbers of existing entries.
		pushNext := false
		for _, entry := range sameDateEntries {
			if entry.Sequence == newEntry.Sequence {
				pushNext = true
			}
			if pushNext {
				// Prevent overflow of allowed range.
				if err := types.ValidateSequence(entry.Sequence); err != nil {
					return nil, err
				}

				// Update the sequence number of the existing entry to resolve the conflict.
				entry.Sequence++
				key := collections.Join(ledgerKeyStr, entry.CorrelationId)
				if err := k.LedgerEntries.Set(goCtx, key, *entry); err != nil {
					return nil, fmt.Errorf("could not update sequence number of ledger entry with correlation id %q", entry.CorrelationId)
				}
			}
		}
	}

	// Store the new entry with its correlation ID as the key.
	err := k.LedgerEntries.Set(goCtx, newEntryKey, *newEntry)
	if err != nil {
		return nil, fmt.Errorf("could not set ledger entry with correlation id %q", newEntry.CorrelationId)
	}

	// Emit the ledger entry added event to notify other modules of the new entry.
	// This allows for proper event handling and external integrations.
	err = ctx.EventManager().EmitTypedEvent(types.NewEventLedgerEntryAdded(ledgerKey, newEntry.CorrelationId))
	if err != nil {
		return nil, err
	}

	entries = append(entries, newEntry)
	return entries, nil
}

// UpdateEntryBalances updates the balance amounts and applied amounts for an existing ledger entry.
// This function allows for post-entry modifications to reflect changes in bucket balances
// and applied amounts without creating new entries.
//
// The function is typically used when:
// - Settlement instructions modify the final amounts
// - Balance calculations need to be updated after initial entry creation
// - Applied amounts need to be adjusted based on external factors
//
// Parameters:
// - ctx: The SDK context for state operations
// - ledgerKey: The ledger identifier
// - correlationID: The unique identifier of the entry to update
// - balanceAmounts: New bucket balance amounts
// - appliedAmounts: New applied amounts for the entry
//
// Returns an error if the entry doesn't exist or if the update fails.
func (k Keeper) UpdateEntryBalances(ctx context.Context, ledgerKey *types.LedgerKey, correlationID string, totalAmt sdkmath.Int, appliedAmounts []*types.LedgerBucketAmount, balanceAmounts []*types.BucketBalance) error {
	// Retrieve the existing entry to ensure it exists before updating.
	existingEntry, err := k.GetLedgerEntry(ctx, ledgerKey, correlationID)
	if err != nil {
		return err
	}

	// This can only be used to update an entry that already exists.
	if existingEntry == nil {
		return types.NewErrCodeNotFound("entry")
	}

	// Make sure everything provided was okay.
	if err = types.ValidateLedgerEntryAmounts(totalAmt, appliedAmounts, balanceAmounts); err != nil {
		return err
	}

	// Update the entry's fields.
	existingEntry.TotalAmt = totalAmt
	existingEntry.AppliedAmounts = appliedAmounts
	existingEntry.BalanceAmounts = balanceAmounts

	// Store the updated entry back to the state store.
	ledgerKeyStr := ledgerKey.String()
	err = k.LedgerEntries.Set(ctx, collections.Join(ledgerKeyStr, correlationID), *existingEntry)
	if err != nil {
		return err
	}

	return nil
}

// ListLedgerEntries retrieves all ledger entries for a given ledger.
// This function walks through all entries associated with a specific ledger key.
// It returns entries sorted by effective date and sequence number for proper ordering.
func (k Keeper) ListLedgerEntries(ctx context.Context, key *types.LedgerKey) ([]*types.LedgerEntry, error) {
	keyStr := key.String()

	// Check if the ledger exists before attempting to list entries.
	// Garbage in, garbage out.
	if !k.HasLedger(sdk.UnwrapSDKContext(ctx), key) {
		return nil, nil
	}

	// Create a prefix range to find all entries for this ledger.
	prefix := collections.NewPrefixedPairRange[string, string](keyStr)

	// Initialize a slice to collect all ledger entries.
	entries := make([]*types.LedgerEntry, 0)

	// Walk through all entry records that match the ledger prefix.
	err := k.LedgerEntries.Walk(ctx, prefix, func(_ collections.Pair[string, string], value types.LedgerEntry) (stop bool, err error) {
		entries = append(entries, &value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// Sort the entries by effective date and sequence number.
	// This ensures proper chronological ordering of ledger entries.
	slices.SortFunc(entries, (*types.LedgerEntry).Compare)

	return entries, nil
}

// GetLedgerEntry retrieves a ledger entry by its correlation ID for a specific ledger.
// This function looks up a specific entry using the ledger key and correlation ID.
// It returns nil if the entry doesn't exist.
func (k Keeper) GetLedgerEntry(ctx context.Context, key *types.LedgerKey, correlationID string) (*types.LedgerEntry, error) {
	// Create the composite key for the specific ledger entry.
	ledgerEntry, err := k.LedgerEntries.Get(ctx, collections.Join(key.String(), correlationID))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &ledgerEntry, nil
}

// RequireGetLedgerEntry retrieves a ledger entry and requires it to exist.
// This function is similar to GetLedgerEntry but returns an error if the entry is not found.
// It's used when the ledger entry must exist for the operation to proceed.
func (k Keeper) RequireGetLedgerEntry(ctx context.Context, lk *types.LedgerKey, correlationID string) (*types.LedgerEntry, error) {
	ledgerEntry, err := k.GetLedgerEntry(ctx, lk, correlationID)
	if err != nil {
		return nil, err
	}
	if ledgerEntry == nil {
		return nil, types.NewErrCodeNotFound("ledger entry")
	}

	return ledgerEntry, nil
}

// GetBalancesAsOf returns the principal, interest, and other balances as of a specific effective date.
// This function calculates the current state of all buckets by processing ledger entries up to the specified date.
// It provides a point-in-time view of the ledger's financial state.
func (k Keeper) GetBalancesAsOf(ctx context.Context, key *types.LedgerKey, asOfDate time.Time) ([]*types.BucketBalance, error) {
	// Convert the date to days since epoch for comparison with ledger entries.
	asOfDateInt := helper.DaysSinceEpoch(asOfDate.UTC())

	// Get all ledger entries for this ledger.
	entries, err := k.ListLedgerEntries(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, types.NewErrCodeNotFound("ledger entries")
	}

	// Map of bucket type ID to bucket balance.
	// This tracks the latest balance for each bucket type.
	bucketBalances := make(map[int32]*types.BucketBalance, 0)

	// Calculate balances up to the specified date.
	// Process entries in chronological order to build the current state.
	for _, entry := range entries {
		// Skip entries after the asOfDate to get the state as of that date.
		if entry.EffectiveDate > asOfDateInt {
			break
		}

		// Update the latest balance for each bucket type.
		// We just find the latest entry as of the asOfDate, and set the balances.
		for _, bucketBalance := range entry.BalanceAmounts {
			bucketBalances[bucketBalance.BucketTypeId] = bucketBalance
		}
	}

	// Convert the map to a slice for return.
	bucketBalancesList := make([]*types.BucketBalance, len(bucketBalances))
	for i, bt := range slices.Sorted(maps.Keys(bucketBalances)) {
		bucketBalancesList[i] = bucketBalances[bt]
	}

	return bucketBalancesList, nil
}
