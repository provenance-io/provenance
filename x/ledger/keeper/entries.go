package keeper

import (
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

	if err := validateLedgerEntryBasic(&le); err != nil {
		return err
	}

	// Validate that the ledger exists
	_, err = k.Ledgers.Get(ctx, nftAddress)
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
