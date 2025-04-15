package keeper

import (
	"sort"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ EntriesKeeper = (*BaseEntriesKeeper)(nil)

type EntriesKeeper interface {
	AppendEntries(ctx sdk.Context, nftAddress string, les []*ledger.LedgerEntry) error
}

type BaseEntriesKeeper struct {
	BaseViewKeeper
}

// SetValue stores a value with a given key.
func (k BaseEntriesKeeper) AppendEntries(ctx sdk.Context, nftAddress string, les []*ledger.LedgerEntry) error {
	// TODO validate that the {addr} can be modified by the signer...

	if len(les) == 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "entries", "cannot be nil or empty")
	}

	if !k.HasLedger(ctx, nftAddress) {
		return NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	// Get all existing entries for this NFT
	entries, err := k.ListLedgerEntries(ctx, nftAddress)
	if err != nil {
		return err
	}

	for _, le := range les {
		if err := ValidateLedgerEntryBasic(le); err != nil {
			return err
		}

		// Validate dates
		if err := validateEntryDates(le, ctx); err != nil {
			return err
		}

		// Validate amounts
		if err := validateEntryAmounts(le); err != nil {
			return err
		}

		// Validate entry type
		if err := validateEntryType(le); err != nil {
			return err
		}

		err := k.saveEntry(ctx, nftAddress, entries, le)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k BaseEntriesKeeper) saveEntry(ctx sdk.Context, nftAddress string, entries []*ledger.LedgerEntry, le *ledger.LedgerEntry) error {
	// Find entries with the same effective date
	var sameDateEntries []ledger.LedgerEntry
	for _, entry := range entries {
		if entry.EffectiveDate == le.EffectiveDate {
			sameDateEntries = append(sameDateEntries, *entry)
		}

		// If the entry's correlation id is already in the list, we need to error
		if entry.CorrelationId == le.CorrelationId {
			return NewLedgerCodedError(ErrCodeAlreadyExists, "correlation_id")
		}
	}

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
				key := collections.Join(nftAddress, entry.CorrelationId)
				if err := k.LedgerEntries.Set(ctx, key, entry); err != nil {
					return err
				}
			}
		}
	}

	// Store the new entry
	key := collections.Join(nftAddress, le.CorrelationId)
	err := k.LedgerEntries.Set(ctx, key, *le)
	if err != nil {
		return err
	}

	// Emit the ledger entry added event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerEntryAdded(
		nftAddress,
		le.CorrelationId,
	))

	return nil
}

// validateEntryDates checks if the dates are valid
func validateEntryDates(le *ledger.LedgerEntry, ctx sdk.Context) error {
	blockTimeDays := DaysSinceEpoch(ctx.BlockTime().UTC())

	if le.PostedDate <= 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "posted_date", "is not a valid date")
	}

	// Check if posted date is in the future
	if le.PostedDate > blockTimeDays {
		return NewLedgerCodedError(ErrCodeInvalidField, "posted_date", "cannot be in the future")
	}

	return nil
}

// validateEntryAmounts checks if the amounts are valid
func validateEntryAmounts(le *ledger.LedgerEntry) error {
	// Check if total amount matches sum of applied amounts
	totalApplied := int64(0)
	for _, applied := range le.AppliedAmounts {
		totalApplied += applied.AppliedAmt
	}

	if le.TotalAmt != totalApplied {
		return NewLedgerCodedError(ErrCodeInvalidField, "amount", "must equal sum of applied amounts")
	}

	return nil
}

// validateEntryType checks if the entry type is valid
func validateEntryType(le *ledger.LedgerEntry) error {
	if le.Type == ledger.LedgerEntryType_Unspecified {
		return NewLedgerCodedError(ErrCodeInvalidField, "entry_type", "cannot be unspecified")
	}

	return nil
}
