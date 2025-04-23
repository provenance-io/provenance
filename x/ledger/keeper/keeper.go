package keeper

import (
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

var _ Keeper = (*BaseKeeper)(nil)

type Keeper interface {
}

// Keeper defines the mymodule keeper.
type BaseKeeper struct {
	BaseViewKeeper
	BaseConfigKeeper
	BaseEntriesKeeper
	BaseFundTransferKeeper
}

const (
	ledgerPrefix                 = "ledgers"
	entriesPrefix                = "ledger_entries"
	fundTransfersPrefix          = "fund_transfers"
	ledgerClassesPrefix          = "ledger_classes"
	ledgerClassEntryTypesPrefix  = "ledger_class_entry_types"
	ledgerClassStatusTypesPrefix = "ledger_class_status_types"
	ledgerClassBucketTypesPrefix = "ledger_class_bucket_types"
)

// NewKeeper returns a new mymodule Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, bankKeeper BankKeeper) BaseKeeper {
	viewKeeper := NewBaseViewKeeper(cdc, storeKey, storeService)

	return BaseKeeper{
		BaseViewKeeper: viewKeeper,
		BaseConfigKeeper: BaseConfigKeeper{
			BaseViewKeeper: viewKeeper,
			BankKeeper:     bankKeeper,
		},
		BaseEntriesKeeper: BaseEntriesKeeper{
			BaseViewKeeper: viewKeeper,
		},
		BaseFundTransferKeeper: BaseFundTransferKeeper{
			BankKeeper: bankKeeper,
		},
	}
}
