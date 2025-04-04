package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

// SetValue stores a value with a given key.
func (k LedgerKeeper) CreateLedger(ctx sdk.Context, l ledger.Ledger) error {
	if err := validateLedgerBasic(&l); err != nil {
		return err
	}

	if err := k.validateLedgerNotExists(ctx, &l); err != nil {
		return err
	}

	// We omit the nftAddress out of the data we store intentionally as
	// a minor optimization since it is also our data key.
	nftAddress := l.NftAddress
	l.NftAddress = ""

	err := k.Ledgers.Set(ctx, nftAddress, l)
	if err != nil {
		return err
	}

	return nil
}

func (k LedgerKeeper) GetLedger(ctx sdk.Context, nftAddress string) (*ledger.Ledger, error) {
	l, err := k.Ledgers.Get(ctx, nftAddress)

	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (k LedgerKeeper) HasLedger(ctx sdk.Context, nftAddress string) bool {
	l, err := k.GetLedger(ctx, nftAddress)
	if err != nil {
		panic(err)
	}

	return l != nil
}
