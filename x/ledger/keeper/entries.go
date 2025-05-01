package keeper

import (
	"sort"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ EntriesKeeper = (*BaseEntriesKeeper)(nil)

type EntriesKeeper interface {
	AppendEntries(ctx sdk.Context, authorityAddr sdk.AccAddress, ledgerKey *ledger.LedgerKey, entries []*ledger.LedgerEntry) error
	UpdateEntryBalances(ctx sdk.Context, authorityAddr sdk.AccAddress, ledgerKey *ledger.LedgerKey, correlationId string, bucketBalances []*ledger.BucketBalance) error
}

type BaseEntriesKeeper struct {
	BaseViewKeeper
}

// SetValue stores a value with a given key.
func (k BaseEntriesKeeper) AppendEntries(ctx sdk.Context, authorityAddr sdk.AccAddress, ledgerKey *ledger.LedgerKey, entries []*ledger.LedgerEntry) error {
	// Validate the key
	err := ValidateLedgerKeyBasic(ledgerKey)
	if err != nil {
		return err
	}

	// Need to resolve the ledger class for validation purposes
	ledger, err := k.GetLedger(ctx, ledgerKey)
	if err != nil {
		return err
	}
	if ledger == nil {
		return NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	// TODO validate against the registry that the authority address is the servicer for this NFT.
	// If there isn't a registry entry we'll verify against the owner of the nftId.

	// Validate that the NFT exists
	if !k.NFTKeeper.HasNFT(ctx, ledgerKey.AssetClassId, ledgerKey.NftId) {
		return NewLedgerCodedError(ErrCodeNotFound, "nft")
	}

	// Get all existing entries for this NFT
	existingEntries, err := k.ListLedgerEntries(ctx, ledgerKey)
	if err != nil {
		return err
	}

	for _, le := range entries {
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

		// Validate that the LedgerClassEntryType exists
		hasLedgerClassEntryType, err := k.LedgerClassEntryTypes.Has(ctx, collections.Join(ledger.LedgerClassId, le.EntryTypeId))
		if err != nil {
			return err
		}
		if !hasLedgerClassEntryType {
			return NewLedgerCodedError(ErrCodeInvalidField, "entry_type_id")
		}

		err = k.saveEntry(ctx, ledgerKey, existingEntries, le)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k BaseEntriesKeeper) UpdateEntryBalances(ctx sdk.Context, authorityAddr sdk.AccAddress, ledgerKey *ledger.LedgerKey, correlationId string, bucketBalances []*ledger.BucketBalance) error {
	// Validate the key
	err := ValidateLedgerKeyBasic(ledgerKey)
	if err != nil {
		return err
	}

	// Get the existing entry
	existingEntry, err := k.GetLedgerEntry(ctx, ledgerKey, correlationId)
	if err != nil {
		return err
	}

	if existingEntry == nil {
		return NewLedgerCodedError(ErrCodeNotFound, "entry")
	}

	// Validate the bucket balances
	for _, bb := range bucketBalances {
		if err := ValidateBucketBalance(bb); err != nil {
			return err
		}
	}

	existingEntry.BalanceAmounts = bucketBalances

	ledgerKeyStr, err := LedgerKeyToString(ledgerKey)
	if err != nil {
		return err
	}

	err = k.LedgerEntries.Set(ctx, collections.Join(*ledgerKeyStr, correlationId), *existingEntry)
	if err != nil {
		return err
	}

	return nil
}

func (k BaseEntriesKeeper) saveEntry(ctx sdk.Context, ledgerKey *ledger.LedgerKey, entries []*ledger.LedgerEntry, le *ledger.LedgerEntry) error {
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

	// Get the string representation of the ledger key for use in k/v store
	ledgerKeyStr, err := LedgerKeyToString(ledgerKey)
	if err != nil {
		return err
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
				key := collections.Join(*ledgerKeyStr, entry.CorrelationId)
				if err := k.LedgerEntries.Set(ctx, key, entry); err != nil {
					return err
				}
			}
		}
	}

	// Store the new entry
	entryKey := collections.Join(*ledgerKeyStr, le.CorrelationId)
	err = k.LedgerEntries.Set(ctx, entryKey, *le)
	if err != nil {
		return err
	}

	// Emit the ledger entry added event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerEntryAdded(
		ledgerKey,
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
	totalApplied := math.NewInt(0)
	for _, applied := range le.AppliedAmounts {
		totalApplied = totalApplied.Add(applied.AppliedAmt.Abs())
	}

	if !le.TotalAmt.Equal(totalApplied) {
		return NewLedgerCodedError(ErrCodeInvalidField, "total_amt", "must equal sum of abs(applied amounts)")
	}

	return nil
}
