package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

// SetValue stores a value with a given key.
func (k LedgerKeeper) AppendEntry(ctx sdk.Context, nftAddress string, le ledger.LedgerEntry) error {
	if emptyString(&nftAddress) {
		return NewLedgerCodedError(ErrCodeMissingField, "field[nft_address]")
	}

	if err := validateLedgerEntryBasic(&le); err != nil {
		return err
	}

	_, err := k.getAddress(&nftAddress)
	if err != nil {
		return err
	}

	// TODO validate that the {addr} can be modified by the signer...

	key := collections.Join(nftAddress, le.Uuid)
	err = k.LedgerEntries.Set(ctx, key, le)
	if err != nil {
		return err
	}

	// Emit the ledger entry added event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerEntryAdded(
		nftAddress,
		le.Uuid,
		int32(le.Type),
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

func (k LedgerKeeper) ListLedgerEntries(ctx context.Context, nftAddress string) ([]ledger.LedgerEntry, error) {
	prefix := collections.NewPrefixedPairRange[string, string](nftAddress)

	iter, err := k.LedgerEntries.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var entries []ledger.LedgerEntry
	for ; iter.Valid(); iter.Next() {
		le, err := iter.Value()
		if err != nil {
			return nil, err
		}
		entries = append(entries, le)
	}
	return entries, nil
}

// GetLedgerEntry retrieves a ledger entry by its UUID
func (k LedgerKeeper) GetLedgerEntry(ctx context.Context, uuid string) (*ledger.LedgerEntry, error) {
	// Find the ledger entry by walking through all entries
	var foundEntry *ledger.LedgerEntry
	err := k.LedgerEntries.Walk(ctx, nil, func(key collections.Pair[string, string], value ledger.LedgerEntry) (stop bool, err error) {
		if value.Uuid == uuid {
			foundEntry = &value
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	if foundEntry == nil {
		return nil, fmt.Errorf("ledger entry not found")
	}
	return foundEntry, nil
}
