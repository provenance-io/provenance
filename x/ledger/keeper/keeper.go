package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

// Keeper defines the mymodule keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	schema   collections.Schema

	Ledgers       collections.Map[string, ledger.Ledger]
	LedgerEntries collections.Map[collections.Pair[string, string], ledger.LedgerEntry]
}

// NewKeeper returns a new mymodule Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	lk := Keeper{
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

func (k Keeper) InitGenesis(ctx sdk.Context, state *ledger.GenesisState) {
	store := k.ledgerStore(ctx)

	for _, l := range state.Ledgers {
		key := []byte(l.NftUuid)
		bz := k.cdc.MustMarshal(&l)
		store.Set(key, bz)
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) {
	// TODO
}

// SetValue stores a value with a given key.
func (k Keeper) CreateLedger(ctx sdk.Context, nftAddress string, denom string) error {
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
func (k Keeper) AppendEntry(ctx sdk.Context, nftAddress string, entryUuid string) error {
	le := ledger.LedgerEntry{
		Uuid: entryUuid,
	}

	key := collections.Join(nftAddress, entryUuid)
	return k.LedgerEntries.Set(ctx, key, le)
}

func (k Keeper) ListLedgerEntries(ctx context.Context, nftAddress string) ([]ledger.LedgerEntry, error) {
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

func (k Keeper) GetLedger(ctx sdk.Context, nftAddress string) (*ledger.Ledger, error) {
	l, err := k.Ledgers.Get(ctx, nftAddress)

	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (k Keeper) ledgerStore(ctx sdk.Context) *prefix.Store {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(ledger.LedgerKeyPrefix))
	return &store
}

func (k Keeper) entryStore(ctx sdk.Context, nftAddress string) *prefix.Store {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(ledger.LedgerKeyPrefix+":entries"+nftAddress))
	return &store
}
