package keeper

import (
	"fmt"
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
	// Validate the NFT address
	_, err := getAddress(&nftAddress)
	if err != nil {
		return err
	}

	if err := ValidateLedgerEntryBasic(&le); err != nil {
		return err
	}

	// Validate that the ledger exists
	_, err = k.Ledgers.Get(ctx, nftAddress)
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

	// TODO validate that the {addr} can be modified by the signer...
	// TODO validate that the ledger entry is not a duplicate

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

	// Check if posted date is in the future
	if le.PostedDate.After(blockTime) {
		return NewLedgerCodedError(ErrCodeInvalidField,
			fmt.Sprintf("posted date cannot be in the future %s (block time: %s)",
				le.PostedDate.Format(time.RFC3339), blockTime.Format(time.RFC3339)))
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

	// Check for negative amounts
	if le.Amt.IsNegative() || le.PrinAppliedAmt.IsNegative() ||
		le.IntAppliedAmt.IsNegative() || le.OtherAppliedAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "amounts cannot be negative")
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
