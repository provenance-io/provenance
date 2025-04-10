package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/provenance-io/provenance/x/ledger"
)

// BankKeeper is an interface that allows the ledger keeper to send coins.
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SpendableCoin(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// Keeper defines the mymodule keeper.
type LedgerKeeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	schema     collections.Schema
	bankKeeper BankKeeper

	Ledgers       collections.Map[string, ledger.Ledger]
	LedgerEntries collections.Map[collections.Pair[string, string], ledger.LedgerEntry]
	FundTransfers collections.Map[collections.Pair[string, string], ledger.FundTransfer]
}

const (
	ledgerPrefix        = "ledgers"
	entriesPrefix       = "ledger_entries"
	fundTransfersPrefix = "fund_transfers"
)

// NewKeeper returns a new mymodule Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, bankKeeper bankkeeper.BaseKeeper) LedgerKeeper {
	sb := collections.NewSchemaBuilder(storeService)

	lk := LedgerKeeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,

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

func (k LedgerKeeper) InitGenesis(ctx sdk.Context, state *ledger.GenesisState) {
	for _, l := range state.Ledgers {
		if err := k.CreateLedger(ctx, l); err != nil {
			// May as well panic here as there is no way we should genesis with bad data.
			panic(err)
		}
	}
}

func (k LedgerKeeper) ExportGenesis(ctx sdk.Context) *ledger.GenesisState {
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
		state.Ledgers = append(state.Ledgers, value)
	}

	return state
}
