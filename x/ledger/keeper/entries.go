package keeper

import (
	"fmt"
	"sort"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ EntriesKeeper = (*BaseEntriesKeeper)(nil)

type EntriesKeeper interface {
	AppendEntry(ctx sdk.Context, nftAddress string, le ledger.LedgerEntry) error
}

type BaseEntriesKeeper struct {
	BaseViewKeeper
}

// SetValue stores a value with a given key.
func (k BaseEntriesKeeper) AppendEntry(ctx sdk.Context, nftAddress string, le ledger.LedgerEntry) error {
	if err := ValidateLedgerEntryBasic(&le); err != nil {
		return err
	}

	// TODO validate that the {addr} can be modified by the signer...

	// Validate that the ledger exists
	_, err := k.Ledgers.Get(ctx, nftAddress)
	if err != nil {
		return err
	}

	// Validate dates
	if err := validateEntryDates(&le, ctx); err != nil {
		return err
	}

	// Validate amounts
	if err := validateEntryAmounts(&le); err != nil {
		return err
	}

	// Validate entry type
	if err := validateEntryType(&le); err != nil {
		return err
	}

	// Get all existing entries for this NFT
	entries, err := k.ListLedgerEntries(ctx, nftAddress)
	if err != nil {
		return err
	}

	// Find entries with the same effective date
	var sameDateEntries []ledger.LedgerEntry
	for _, entry := range entries {
		if entry.EffectiveDate == le.EffectiveDate {
			sameDateEntries = append(sameDateEntries, entry)
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
	err = k.LedgerEntries.Set(ctx, key, le)
	if err != nil {
		return err
	}

	// Emit the ledger entry added event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerEntryAdded(
		nftAddress,
		le.CorrelationId,
		le.Type.String(),
		le.PostedDate,
		le.EffectiveDate,
		le.Amt.String(),
	))

	// Emit the balance updated event
	ctx.EventManager().EmitEvent(ledger.NewEventBalanceUpdated(
		nftAddress,
		le.PrinBalAmt.String(),
		le.IntBalAmt.String(),
		le.OtherBalAmt.String(),
	))

	return nil
}

// validateEntryDates checks if the dates are valid
func validateEntryDates(le *ledger.LedgerEntry, ctx sdk.Context) error {
	blockTime := ctx.BlockTime()

	postedDate, err := time.Parse("2006-01-02", le.PostedDate)
	if err != nil {
		return NewLedgerCodedError(ErrCodeInvalidField, "posted date is not a valid date")
	}

	// Check if posted date is in the future
	if postedDate.After(blockTime) {
		return NewLedgerCodedError(ErrCodeInvalidField,
			fmt.Sprintf("posted date cannot be in the future %s (block time: %s)",
				postedDate, blockTime.Format(time.RFC3339)))
	}

	return nil
}

// validateEntryAmounts checks if the amounts are valid
func validateEntryAmounts(le *ledger.LedgerEntry) error {
	// Check if total amount matches sum of applied amounts
	totalApplied := le.PrinAppliedAmt.Add(le.IntAppliedAmt).Add(le.OtherAppliedAmt)
	if !le.Amt.Equal(totalApplied) {
		return NewLedgerCodedError(ErrCodeInvalidField, "total amount must equal sum of applied amounts")
	}

	return nil
}

// validateEntryType checks if the entry type is valid
func validateEntryType(le *ledger.LedgerEntry) error {
	if le.Type == ledger.LedgerEntryType_Unspecified {
		return NewLedgerCodedError(ErrCodeInvalidField, "entry type cannot be unspecified")
	}

	return nil
}
