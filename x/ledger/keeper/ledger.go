package keeper

import (
	"errors"

	"cosmossdk.io/collections"
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

	// Emit the ledger created event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerCreated(nftAddress, l.Denom))

	return nil
}

// GetLedger retrieves a ledger by its NFT address.
//
// Parameters:
//   - ctx: The SDK context
//   - nftAddress: The NFT address to look up the ledger for
//
// Returns:
//   - *ledger.Ledger: A pointer to the found ledger, or nil if not found
//   - error: Any error that occurred during retrieval, or nil if successful
//
// Behavior:
//   - Returns (nil, nil) if the ledger is not found
//   - Returns (nil, err) if an error occurs during retrieval
//   - Returns (&ledger, nil) if the ledger is found successfully
//   - The returned ledger will have its NftAddress field set to the provided nftAddress
func (k LedgerKeeper) GetLedger(ctx sdk.Context, nftAddress string) (*ledger.Ledger, error) {
	// Validate the NFT address
	_, err := k.getAddress(&nftAddress)
	if err != nil {
		return nil, err
	}

	// Lookup the NFT address in the ledger
	l, err := k.Ledgers.Get(ctx, nftAddress)
	if err != nil {
		// Eat the not found error as it is expected, and return nil.
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}

		// Otherwise, return the error.
		return nil, err
	}

	// The NFT address isn't stored in the ledger, so we add it back in.
	l.NftAddress = nftAddress
	return &l, nil
}

func (k LedgerKeeper) HasLedger(ctx sdk.Context, nftAddress string) bool {
	has, _ := k.Ledgers.Has(ctx, nftAddress)
	return has
}
