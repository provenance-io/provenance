package keeper

import (
	"context"

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

	key := collections.Join(nftAddress, le.Uuid)
	return k.LedgerEntries.Set(ctx, key, le)
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
