package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ ConfigKeeper = (*BaseConfigKeeper)(nil)

type ConfigKeeper interface {
	CreateLedger(ctx sdk.Context, l ledger.Ledger) error
}

type BaseConfigKeeper struct {
	BaseViewKeeper
}

// SetValue stores a value with a given key.
func (k BaseConfigKeeper) CreateLedger(ctx sdk.Context, l ledger.Ledger) error {
	if err := validateLedgerBasic(&l); err != nil {
		return err
	}

	if k.HasLedger(ctx, l.NftAddress) {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger")
	}

	_, err := getAddress(&l.NftAddress)
	if err != nil {
		return err
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
	for _, l := range state.Ledgers {
		if err := k.CreateLedger(ctx, l); err != nil {
			// May as well panic here as there is no way we should genesis with bad data.
			panic(err)
		}
	}
}
