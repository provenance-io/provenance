package keeper

import (
	"context"
	"errors"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ ViewKeeper = (*BaseViewKeeper)(nil)

type ViewKeeper interface {
	GetLedger(ctx sdk.Context, nftAddress string) (*ledger.Ledger, error)
	HasLedger(ctx sdk.Context, nftAddress string) bool
	ListLedgerEntries(ctx context.Context, nftAddress string) ([]*ledger.LedgerEntry, error)
	GetLedgerEntry(ctx context.Context, nftAddress string, correlationID string) (*ledger.LedgerEntry, error)
	GetBalancesAsOf(ctx context.Context, nftAddress string, asOfDate time.Time) (*ledger.Balances, error)
}

type BaseViewKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	schema   collections.Schema

	Ledgers       collections.Map[string, ledger.Ledger]
	LedgerEntries collections.Map[collections.Pair[string, string], ledger.LedgerEntry]
	FundTransfers collections.Map[collections.Pair[string, string], ledger.FundTransfer]
}

func NewBaseViewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService) BaseViewKeeper {
	sb := collections.NewSchemaBuilder(storeService)

	lk := BaseViewKeeper{
		cdc:      cdc,
		storeKey: storeKey,

		Ledgers: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerPrefix),
			ledgerPrefix,
			collections.StringKey,
			codec.CollValue[ledger.Ledger](cdc),
		),
		LedgerEntries: collections.NewMap(
			sb,
			collections.NewPrefix(entriesPrefix),
			entriesPrefix,
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[ledger.LedgerEntry](cdc),
		),
		FundTransfers: collections.NewMap(
			sb,
			collections.NewPrefix(fundTransfersPrefix),
			fundTransfersPrefix,
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[ledger.FundTransfer](cdc),
		),
	}

	// Build and set the schema
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	lk.schema = schema

	return lk
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
func (k BaseViewKeeper) GetLedger(ctx sdk.Context, nftAddress string) (*ledger.Ledger, error) {
	// Validate the NFT address
	_, err := getAddress(&nftAddress)
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

func (k BaseViewKeeper) HasLedger(ctx sdk.Context, nftAddress string) bool {
	has, _ := k.Ledgers.Has(ctx, nftAddress)
	return has
}

func (k BaseViewKeeper) ListLedgerEntries(ctx context.Context, nftAddress string) ([]*ledger.LedgerEntry, error) {
	// Garbage in, garbage out.
	if !k.HasLedger(sdk.UnwrapSDKContext(ctx), nftAddress) {
		return nil, nil
	}

	// Get all entries for the ledger.
	prefix := collections.NewPrefixedPairRange[string, string](nftAddress)
	iter, err := k.LedgerEntries.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	// Iterate through all entries for the ledger.
	var entries []*ledger.LedgerEntry
	for ; iter.Valid(); iter.Next() {
		le, err := iter.Value()
		if err != nil {
			return nil, err
		}
		entries = append(entries, &le)
	}

	// Sort the entries by effective date and sequence number.
	sortLedgerEntries(entries)

	return entries, nil
}

// GetLedgerEntry retrieves a ledger entry by its correlation ID for a specific NFT address
func (k BaseViewKeeper) GetLedgerEntry(ctx context.Context, nftAddress string, correlationID string) (*ledger.LedgerEntry, error) {
	// Validate the NFT address
	_, err := getAddress(&nftAddress)
	if err != nil {
		return nil, err
	}

	if !isCorrelationIDValid(correlationID) {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "correlation_id")
	}

	entries, err := k.ListLedgerEntries(ctx, nftAddress)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.CorrelationId == correlationID {
			return entry, nil
		}
	}

	// If we get here, the entry was not found
	return nil, nil
}

func (k BaseViewKeeper) ExportGenesis(ctx sdk.Context) *ledger.GenesisState {
	state := &ledger.GenesisState{}

	// Iterate through all ledgers
	iter, err := k.Ledgers.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			panic(err)
		}
		value, err := iter.Value()
		if err != nil {
			panic(err)
		}
		// Set the NftAddress back since it's not stored in the value
		value.NftAddress = key
		// state.Ledgers = append(state.Ledgers, value)
	}

	return state
}

// GetBalancesAsOf returns the principal, interest, and other balances as of a specific effective date
func (k BaseViewKeeper) GetBalancesAsOf(ctx context.Context, nftAddress string, asOfDate time.Time) (*ledger.Balances, error) {
	// Validate the NFT address
	_, err := getAddress(&nftAddress)
	if err != nil {
		return nil, err
	}

	// Get all ledger entries for this NFT
	entries, err := k.ListLedgerEntries(ctx, nftAddress)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "ledger entries")
	}

	// Initialize balances
	balances := &ledger.Balances{
		Principal: math.NewInt(0),
		Interest:  math.NewInt(0),
		Other:     math.NewInt(0),
	}

	sortLedgerEntries(entries)

	// Calculate balances up to the specified date
	for _, entry := range entries {
		// Safe to ignore the error since the sort parses the effective date already.
		// This should also be safe since we should have already validated the date format when the entry was added.
		effectiveDate, _ := parseIS08601Date(entry.EffectiveDate)

		// Skip entries after the asOfDate
		if effectiveDate.After(asOfDate) {
			break
		}

		// Update balances based on the entry
		balances.Principal = balances.Principal.Add(entry.PrinAppliedAmt)
		balances.Interest = balances.Interest.Add(entry.IntAppliedAmt)
		balances.Other = balances.Other.Add(entry.OtherAppliedAmt)
	}

	return balances, nil
}
