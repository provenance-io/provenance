package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ ConfigKeeper = (*BaseConfigKeeper)(nil)

type ConfigKeeper interface {
	CreateLedger(ctx sdk.Context, l ledger.Ledger) error
	DestroyLedger(ctx sdk.Context, nftAddress string) error
}

type BaseConfigKeeper struct {
	BaseViewKeeper
	BankKeeper
}

// SetValue stores a value with a given key.
func (k BaseConfigKeeper) CreateLedger(ctx sdk.Context, l ledger.Ledger) error {
	if err := ValidateLedgerBasic(&l); err != nil {
		return err
	}

	if k.HasLedger(ctx, l.NftAddress) {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger")
	}

	_, err := getAddress(&l.NftAddress)
	if err != nil {
		return err
	}

	// Validate that the denom exists in the bank keeper to avoid garbage tokens being used.
	if !k.HasSupply(ctx, l.Denom) {
		return NewLedgerCodedError(ErrCodeInvalidField, "denom")
	}

	// TODO validate that the {addr} can be modified by the signer...

	// We omit the nftAddress out of the data we store intentionally as
	// a minor optimization since it is also our data key.
	nftAddress := l.NftAddress
	l.NftAddress = ""

	err = k.Ledgers.Set(ctx, nftAddress, l)
	if err != nil {
		return err
	}

	// Emit the ledger created event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerCreated(nftAddress, l.Denom))

	return nil
}

func (k BaseKeeper) InitGenesis(ctx sdk.Context, state *ledger.GenesisState) {
	// no-op: we start with a clean ledger state.
}

// DestroyLedger removes a ledger from the store by NFT address
func (k BaseConfigKeeper) DestroyLedger(ctx sdk.Context, nftAddress string) error {
	if !k.HasLedger(ctx, nftAddress) {
		return NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	// Remove the ledger from the store
	err := k.Ledgers.Delete(ctx, nftAddress)
	if err != nil {
		return err
	}

	// Emit the ledger destroyed event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerDestroyed(nftAddress))

	return nil
}
