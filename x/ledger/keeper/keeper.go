package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

// Keeper defines the mymodule keeper.
type LedgerKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	schema   collections.Schema

	Ledgers       collections.Map[string, ledger.Ledger]
	LedgerEntries collections.Map[collections.Pair[string, string], ledger.LedgerEntry]
}

// NewKeeper returns a new mymodule Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService) LedgerKeeper {
	sb := collections.NewSchemaBuilder(storeService)

	lk := LedgerKeeper{
		cdc:      cdc,
		storeKey: storeKey,

		Ledgers: collections.NewMap(
			sb,
			collections.NewPrefix("ledgers"),
			"ledgers",
			collections.StringKey,
			codec.CollValue[ledger.Ledger](cdc),
		),
		LedgerEntries: collections.NewMap(
			sb,
			collections.NewPrefix("ledger_entries"),
			"ledger_entries",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[ledger.LedgerEntry](cdc),
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

func (k LedgerKeeper) InitGenesis(ctx sdk.Context, state *ledger.GenesisState) {
	for _, l := range state.Ledgers {
		k.CreateLedger(ctx, l.NftAddress, l.Denom)
	}
}

func (k LedgerKeeper) ExportGenesis(ctx sdk.Context) {
	// TODO
}

// SetValue stores a value with a given key.
func (k LedgerKeeper) CreateLedger(ctx sdk.Context, nftAddress string, denom string) error {
	l := ledger.Ledger{
		Denom: denom,
	}

	err := k.Ledgers.Set(ctx, nftAddress, l)
	if err != nil {
		return err
	}

	return nil
}

// SetValue stores a value with a given key.
func (k LedgerKeeper) AppendEntry(ctx sdk.Context, nftAddress string, entryUuid string) error {
	le := ledger.LedgerEntry{
		Uuid: entryUuid,
	}

	key := collections.Join(nftAddress, entryUuid)
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

func (k LedgerKeeper) GetLedger(ctx sdk.Context, nftAddress string) (*ledger.Ledger, error) {
	l, err := k.Ledgers.Get(ctx, nftAddress)

	if err != nil {
		return nil, err
	}

	return &l, nil
}
