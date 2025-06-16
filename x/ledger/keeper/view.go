package keeper

import (
	"context"
	"errors"
	"sort"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

var _ ViewKeeper = (*BaseViewKeeper)(nil)

type ViewKeeper interface {
	GetLedger(ctx sdk.Context, key *ledger.LedgerKey) (*ledger.Ledger, error)
	HasLedger(ctx sdk.Context, key *ledger.LedgerKey) bool
	ListLedgerEntries(ctx context.Context, key *ledger.LedgerKey) ([]*ledger.LedgerEntry, error)
	GetLedgerEntry(ctx context.Context, key *ledger.LedgerKey, correlationID string) (*ledger.LedgerEntry, error)
	GetBalancesAsOf(ctx context.Context, key *ledger.LedgerKey, asOfDate time.Time) (*ledger.Balances, error)
	GetLedgerClass(ctx context.Context, ledgerClassId string) (*ledger.LedgerClass, error)
	GetLedgerClassEntryTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassEntryType, error)
	GetLedgerClassStatusTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassStatusType, error)
	GetLedgerClassBucketTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassBucketType, error)
}

type BaseViewKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	schema   collections.Schema

	Ledgers       collections.Map[string, ledger.Ledger]
	LedgerEntries collections.Map[collections.Pair[string, string], ledger.LedgerEntry]
	FundTransfers collections.Map[collections.Pair[string, string], ledger.FundTransfer]

	// LedgerClasses stores the configuration of all ledgers for a given class of asset.
	LedgerClasses          collections.Map[string, ledger.LedgerClass]
	LedgerClassEntryTypes  collections.Map[collections.Pair[string, int32], ledger.LedgerClassEntryType]
	LedgerClassStatusTypes collections.Map[collections.Pair[string, int32], ledger.LedgerClassStatusType]
	LedgerClassBucketTypes collections.Map[collections.Pair[string, int32], ledger.LedgerClassBucketType]

	RegistryKeeper RegistryKeeper
}

func NewBaseViewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, registryKeeper RegistryKeeper) BaseViewKeeper {
	sb := collections.NewSchemaBuilder(storeService)

	lk := BaseViewKeeper{
		cdc:      cdc,
		storeKey: storeKey,

		Ledgers: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerPrefix),
			"ledger",
			collections.StringKey,
			codec.CollValue[ledger.Ledger](cdc),
		),
		LedgerEntries: collections.NewMap(
			sb,
			collections.NewPrefix(entriesPrefix),
			"ledger_entries",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[ledger.LedgerEntry](cdc),
		),
		FundTransfers: collections.NewMap(
			sb,
			collections.NewPrefix(fundTransfersPrefix),
			"fund_transfers",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[ledger.FundTransfer](cdc),
		),

		// Ledger Class configuration data
		LedgerClasses: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassesPrefix),
			"ledger_classes",
			collections.StringKey,
			codec.CollValue[ledger.LedgerClass](cdc),
		),
		LedgerClassEntryTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassEntryTypesPrefix),
			"ledger_class_entry_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[ledger.LedgerClassEntryType](cdc),
		),
		LedgerClassStatusTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassStatusTypesPrefix),
			"ledger_class_status_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[ledger.LedgerClassStatusType](cdc),
		),
		LedgerClassBucketTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassBucketTypesPrefix),
			"ledger_class_bucket_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[ledger.LedgerClassBucketType](cdc),
		),

		RegistryKeeper: registryKeeper,
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
func (k BaseViewKeeper) GetLedger(ctx sdk.Context, key *ledger.LedgerKey) (*ledger.Ledger, error) {
	// Validate the key
	err := ValidateLedgerKeyBasic(key)
	if err != nil {
		return nil, err
	}

	keyStr, err := LedgerKeyToString(key)
	if err != nil {
		return nil, err
	}

	// Lookup the NFT address in the ledger
	l, err := k.Ledgers.Get(ctx, *keyStr)
	if err != nil {
		// Eat the not found error as it is expected, and return nil.
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}

		// Otherwise, return the error.
		return nil, err
	}

	// The NFT address isn't stored in the ledger, so we add it back in.
	l.Key = key
	return &l, nil
}

func (k BaseViewKeeper) HasLedger(ctx sdk.Context, key *ledger.LedgerKey) bool {
	// Validate the key
	err := ValidateLedgerKeyBasic(key)
	if err != nil {
		return false
	}

	keyStr, err := LedgerKeyToString(key)
	if err != nil {
		return false
	}

	has, _ := k.Ledgers.Has(ctx, *keyStr)
	return has
}

func (k BaseViewKeeper) ListLedgerEntries(ctx context.Context, key *ledger.LedgerKey) ([]*ledger.LedgerEntry, error) {
	// Validate the key
	err := ValidateLedgerKeyBasic(key)
	if err != nil {
		return nil, err
	}

	keyStr, err := LedgerKeyToString(key)
	if err != nil {
		return nil, err
	}

	// Garbage in, garbage out.
	if !k.HasLedger(sdk.UnwrapSDKContext(ctx), key) {
		return nil, nil
	}

	// Get all entries for the ledger.
	prefix := collections.NewPrefixedPairRange[string, string](*keyStr)
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
func (k BaseViewKeeper) GetLedgerEntry(ctx context.Context, key *ledger.LedgerKey, correlationID string) (*ledger.LedgerEntry, error) {
	// Validate the key
	err := ValidateLedgerKeyBasic(key)
	if err != nil {
		return nil, err
	}

	if !k.HasLedger(sdk.UnwrapSDKContext(ctx), key) {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	if !isCorrelationIDValid(correlationID) {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "correlation_id")
	}

	entries, err := k.ListLedgerEntries(ctx, key)
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
		_, err := iter.Key()
		if err != nil {
			panic(err)
		}
		_, err = iter.Value()
		if err != nil {
			panic(err)
		}

	}

	return state
}

// GetBalancesAsOf returns the principal, interest, and other balances as of a specific effective date
func (k BaseViewKeeper) GetBalancesAsOf(ctx context.Context, key *ledger.LedgerKey, asOfDate time.Time) (*ledger.Balances, error) {
	// Validate the key
	err := ValidateLedgerKeyBasic(key)
	if err != nil {
		return nil, err
	}

	asOfDateInt := DaysSinceEpoch(asOfDate.UTC())

	// Get all ledger entries for this NFT
	entries, err := k.ListLedgerEntries(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "ledger entries")
	}

	// Map of bucket name to list of bucket balances.
	bucketBalances := make(map[int32]*ledger.BucketBalance, 0)

	// Calculate balances up to the specified date
	for _, entry := range entries {
		// Skip entries after the asOfDate
		if entry.EffectiveDate > asOfDateInt {
			break
		}

		// We just find the latest entry as of the asOfDate, and set the balances.
		for _, bucketBalance := range entry.BalanceAmounts {
			bucketBalances[bucketBalance.BucketTypeId] = bucketBalance
		}
	}

	bucketBalancesList := make([]*ledger.BucketBalance, 0)
	for _, balance := range bucketBalances {
		bucketBalancesList = append(bucketBalancesList, balance)
	}

	// Not sure this really helps sort anything since the id's are arbitrary, but at least we're consistent.
	sort.Slice(bucketBalancesList, func(i, j int) bool {
		return bucketBalancesList[i].BucketTypeId < bucketBalancesList[j].BucketTypeId
	})

	return &ledger.Balances{
		BucketBalances: bucketBalancesList,
	}, nil
}

func (k BaseViewKeeper) GetLedgerClass(ctx context.Context, ledgerClassId string) (*ledger.LedgerClass, error) {
	ledgerClass, err := k.LedgerClasses.Get(ctx, ledgerClassId)
	if err != nil {
		return nil, err
	}
	return &ledgerClass, nil
}

func (k BaseViewKeeper) GetLedgerClassEntryTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassEntryType, error) {
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassId)
	iter, err := k.LedgerClassEntryTypes.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	entryTypes := make([]*ledger.LedgerClassEntryType, 0)
	for ; iter.Valid(); iter.Next() {
		entryType, err := iter.Value()
		if err != nil {
			return nil, err
		}

		entryTypes = append(entryTypes, &entryType)
	}

	return entryTypes, nil
}

func SliceToMap[T any, K comparable](list []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(list))
	for _, item := range list {
		result[keyFn(item)] = item
	}
	return result
}

func (k BaseViewKeeper) GetLedgerClassStatusTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassStatusType, error) {
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassId)
	iter, err := k.LedgerClassStatusTypes.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	statusTypes := make([]*ledger.LedgerClassStatusType, 0)
	for ; iter.Valid(); iter.Next() {
		statusType, err := iter.Value()
		if err != nil {
			return nil, err
		}
		statusTypes = append(statusTypes, &statusType)
	}
	return statusTypes, nil
}

func (k BaseViewKeeper) GetLedgerClassBucketTypes(ctx context.Context, ledgerClassId string) ([]*ledger.LedgerClassBucketType, error) {
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassId)
	iter, err := k.LedgerClassBucketTypes.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	bucketTypes := make([]*ledger.LedgerClassBucketType, 0)
	for ; iter.Valid(); iter.Next() {
		bucketType, err := iter.Value()
		if err != nil {
			return nil, err
		}
		bucketTypes = append(bucketTypes, &bucketType)
	}
	return bucketTypes, nil
}
