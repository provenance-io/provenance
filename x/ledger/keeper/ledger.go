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

	_, err := k.getAddress(&l.NftAddress)
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

	return nil
}

func (k LedgerKeeper) GetLedger(ctx sdk.Context, nftAddress string) (*ledger.Ledger, error) {
	l, err := k.Ledgers.Get(ctx, nftAddress)

	if err != nil {
		return nil, err
	}

	l.NftAddress = nftAddress
	return &l, nil
}

func (k LedgerKeeper) HasLedger(ctx sdk.Context, nftAddress string) bool {
	has, _ := k.Ledgers.Has(ctx, nftAddress)
	return has
}

func (k LedgerKeeper) getAddress(s *string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(*s)
	if err != nil || addr == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "nft_address")
	}

	return addr, nil
}
