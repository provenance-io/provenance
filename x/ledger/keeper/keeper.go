package keeper

import (
	"fmt"

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
}

// NewKeeper returns a new mymodule Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
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
		Denom:  denom,
		Ledger: []*ledger.LedgerEntry{},
	}

	v, err := k.cdc.Marshal(&l)
	if err != nil {
		return err
	}

	store := k.ledgerStore(ctx)
	store.Set([]byte(nftAddress), []byte(v))
	return nil
}

func (k Keeper) GetLedger(ctx sdk.Context, nftAddress string) (*ledger.Ledger, error) {
	store := k.ledgerStore(ctx)

	bz := store.Get([]byte(nftAddress))
	if bz == nil {
		return nil, fmt.Errorf("ledger not found for nft %s", nftAddress)
	}

	var l ledger.Ledger
	err := k.cdc.Unmarshal(bz, &l)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (k Keeper) ledgerStore(ctx sdk.Context) *prefix.Store {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(ledger.LedgerKeyPrefix))
	return &store
}
