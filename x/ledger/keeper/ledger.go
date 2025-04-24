package keeper

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ ConfigKeeper = (*BaseConfigKeeper)(nil)

type ConfigKeeper interface {
	CreateLedgerClass(ctx sdk.Context, l ledger.LedgerClass) error
	AddClassEntryType(ctx sdk.Context, ledgerClassId string, l ledger.LedgerClassEntryType) error
	AddClassStatusType(ctx sdk.Context, ledgerClassId string, l ledger.LedgerClassStatusType) error
	AddClassBucketType(ctx sdk.Context, ledgerClassId string, l ledger.LedgerClassBucketType) error

	CreateLedger(ctx sdk.Context, l ledger.Ledger) error
	DestroyLedger(ctx sdk.Context, nftId string) error
}

type BaseConfigKeeper struct {
	BaseViewKeeper
	BankKeeper
	NFTKeeper
}

func (k BaseConfigKeeper) CreateLedgerClass(ctx sdk.Context, l ledger.LedgerClass) error {
	_, found := k.NFTKeeper.GetClass(ctx, l.AssetClassId)
	if !found {
		return NewLedgerCodedError(ErrCodeInvalidField, "asset_class_id")
	}

	has, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}

	if has {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger class")
	}

	// Validate that the denom exists in the bank keeper to avoid garbage tokens being used.
	if !k.HasSupply(ctx, l.Denom) {
		return NewLedgerCodedError(ErrCodeInvalidField, "denom")
	}

	err = k.LedgerClasses.Set(ctx, l.LedgerClassId, l)
	if err != nil {
		return err
	}

	return nil
}

func (k BaseConfigKeeper) AddClassEntryType(ctx sdk.Context, ledgerClassId string, l ledger.LedgerClassEntryType) error {
	// TODO verify that signature has authority over the assetClassId

	key := collections.Join(ledgerClassId, l.Id)

	has, err := k.LedgerClassEntryTypes.Has(ctx, key)
	if err != nil {
		return err
	}

	if has {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger class entry type")
	}

	err = k.LedgerClassEntryTypes.Set(ctx, key, l)
	if err != nil {
		return err
	}

	return nil
}

func (k BaseConfigKeeper) AddClassStatusType(ctx sdk.Context, ledgerClassId string, l ledger.LedgerClassStatusType) error {
	// TODO verify that signature has authority over the assetClassId

	key := collections.Join(ledgerClassId, l.Id)

	has, err := k.LedgerClassStatusTypes.Has(ctx, key)
	if err != nil {
		return err
	}

	if has {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger class status type")
	}

	err = k.LedgerClassStatusTypes.Set(ctx, key, l)
	if err != nil {
		return err
	}
	return nil
}

func (k BaseConfigKeeper) AddClassBucketType(ctx sdk.Context, ledgerClassId string, l ledger.LedgerClassBucketType) error {
	// TODO verify that signature has authority over the assetClassId

	key := collections.Join(ledgerClassId, l.Id)

	has, err := k.LedgerClassBucketTypes.Has(ctx, key)
	if err != nil {
		return err
	}

	if has {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger class bucket type")
	}

	err = k.LedgerClassBucketTypes.Set(ctx, key, l)
	if err != nil {
		return err
	}

	return nil
}

// SetValue stores a value with a given key.
func (k BaseConfigKeeper) CreateLedger(ctx sdk.Context, l ledger.Ledger) error {
	if err := ValidateLedgerBasic(&l); err != nil {
		return err
	}

	if k.HasLedger(ctx, l.NftId) {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger")
	}

	_, err := getAddress(&l.NftId)
	if err != nil {
		return err
	}

	// Validate that the LedgerClass exists
	hasLedgerClass, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}

	if !hasLedgerClass {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger class")
	}

	// Validate that the LedgerClassStatusType exists
	hasLedgerClassStatusType, err := k.LedgerClassStatusTypes.Has(ctx, collections.Join(l.LedgerClassId, l.StatusTypeId))
	if err != nil {
		return err
	}

	if !hasLedgerClassStatusType {
		return NewLedgerCodedError(ErrCodeInvalidField, "status_type_id")
	}

	// TODO validate that the {addr} can be modified by the signer...

	// We omit the nftAddress out of the data we store intentionally as
	// a minor optimization since it is also our data key.
	nftId := l.NftId
	l.NftId = ""

	err = k.Ledgers.Set(ctx, nftId, l)
	if err != nil {
		return err
	}

	// Emit the ledger created event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerCreated(nftId))

	return nil
}

func (k BaseKeeper) InitGenesis(ctx sdk.Context, state *ledger.GenesisState) {
	// no-op: we start with a clean ledger state.
}

// DestroyLedger removes a ledger from the store by NFT address
func (k BaseConfigKeeper) DestroyLedger(ctx sdk.Context, nftId string) error {
	if !k.HasLedger(ctx, nftId) {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger")
	}

	// Remove the ledger from the store
	err := k.Ledgers.Remove(ctx, nftId)
	if err != nil {
		return err
	}

	prefix := collections.NewPrefixedPairRange[string, string](nftId)

	iter, err := k.LedgerEntries.Iterate(ctx, prefix)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		err = k.LedgerEntries.Remove(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}
