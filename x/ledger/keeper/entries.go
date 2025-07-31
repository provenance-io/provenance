package keeper

import (
	"sort"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/helper"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// SetValue stores a value with a given key.
func (k Keeper) AppendEntries(ctx sdk.Context, ledgerKey *types.LedgerKey, entries []*types.LedgerEntry) error {
	ledger, err := k.RequireGetLedger(ctx, ledgerKey)
	if err != nil {
		return err
	}

	// Get all existing entries for this NFT
	existingEntries, err := k.ListLedgerEntries(ctx, ledgerKey)
	if err != nil {
		return err
	}

	blockTimeDays := helper.DaysSinceEpoch(ctx.BlockTime().UTC())
	for _, le := range entries {
		// Check if posted date is in the future
		if le.PostedDate > blockTimeDays {
			return types.NewErrCodeInvalidField("posted_date", "cannot be in the future")
		}

		// Validate that the LedgerClassEntryType exists
		hasLedgerClassEntryType, err := k.LedgerClassEntryTypes.Has(ctx, collections.Join(ledger.LedgerClassId, le.EntryTypeId))
		if err != nil {
			return err
		}
		if !hasLedgerClassEntryType {
			return types.NewErrCodeInvalidField("entry_type_id", "entry type doesn't exist")
		}

		err = k.saveEntry(ctx, ledgerKey, existingEntries, le)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) UpdateEntryBalances(ctx sdk.Context, ledgerKey *types.LedgerKey, correlationId string, balanceAmounts []*types.BucketBalance, appliedAmounts []*types.LedgerBucketAmount) error {
	// Get the existing entry
	existingEntry, err := k.GetLedgerEntry(ctx, ledgerKey, correlationId)
	if err != nil {
		return err
	}

	if existingEntry == nil {
		return types.NewErrCodeNotFound("entry")
	}

	// Update the entry with the new applied amounts
	existingEntry.AppliedAmounts = appliedAmounts

	// Update the entry with the new bucket balances
	existingEntry.BalanceAmounts = balanceAmounts

	ledgerKeyStr := ledgerKey.String()
	err = k.LedgerEntries.Set(ctx, collections.Join(ledgerKeyStr, correlationId), *existingEntry)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) saveEntry(ctx sdk.Context, ledgerKey *types.LedgerKey, entries []*types.LedgerEntry, le *types.LedgerEntry) error {
	// Find entries with the same effective date
	var sameDateEntries []types.LedgerEntry
	for _, entry := range entries {
		if entry.EffectiveDate == le.EffectiveDate {
			sameDateEntries = append(sameDateEntries, *entry)
		}

		// If the entry's correlation id is already in the list, we need to error
		if entry.CorrelationId == le.CorrelationId {
			return types.NewErrCodeAlreadyExists("correlation_id")
		}
	}

	// Get the string representation of the ledger key for use in k/v store
	ledgerKeyStr := ledgerKey.String()

	// If there are entries with the same date, check for sequence number conflicts
	if len(sameDateEntries) > 0 {
		// Sort entries by sequence number
		sort.Slice(sameDateEntries, func(i, j int) bool {
			return sameDateEntries[i].Sequence < sameDateEntries[j].Sequence
		})

		// Check if the new entry's sequence number conflicts with existing entries
		pushNext := false
		for _, entry := range sameDateEntries {
			if pushNext || entry.Sequence == le.Sequence {
				pushNext = true

				// Update the sequence number of the existing entry
				entry.Sequence++
				key := collections.Join(ledgerKeyStr, entry.CorrelationId)
				if err := k.LedgerEntries.Set(ctx, key, entry); err != nil {
					return err
				}
			}
		}
	}

	// Store the new entry
	entryKey := collections.Join(ledgerKeyStr, le.CorrelationId)
	err := k.LedgerEntries.Set(ctx, entryKey, *le)
	if err != nil {
		return err
	}

	// Emit the ledger entry added event
	ctx.EventManager().EmitTypedEvent(types.NewEventLedgerEntryAdded(
		ledgerKey,
		le.CorrelationId,
	))

	return nil
}
