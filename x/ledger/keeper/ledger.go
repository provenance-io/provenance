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

	CreateLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, l ledger.Ledger) error
	DestroyLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, key *ledger.LedgerKey) error
}

type BaseConfigKeeper struct {
	BaseViewKeeper
	BankKeeper
}

func (k BaseConfigKeeper) CreateLedgerClass(ctx sdk.Context, l ledger.LedgerClass) error {
	hasAssetClass := k.BaseViewKeeper.AssetClassExists(ctx, &l.AssetClassId)
	if !hasAssetClass {
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

func (k BaseConfigKeeper) CreateLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, l ledger.Ledger) error {
	if err := ValidateLedgerBasic(&l); err != nil {
		return err
	}

	// Do not allow creating a duplicate ledger
	if k.HasLedger(ctx, l.Key) {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger")
	}

	// Validate that the LedgerClass exists
	hasLedgerClass, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}
	if !hasLedgerClass {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger class")
	}

	// Validate that the NFT exists
	if !k.NFTKeeper.HasNFT(ctx, l.Key.AssetClassId, l.Key.NftId) {
		return NewLedgerCodedError(ErrCodeNotFound, "nft")
	}

	// Validate that the authority has ownership of the NFT
	nftOwner := k.BaseViewKeeper.GetNFTOwner(ctx, &l.Key.AssetClassId, &l.Key.NftId)
	if nftOwner.String() != authorityAddr.String() {
		return NewLedgerCodedError(ErrCodeUnauthorized, "nft owner")
	}

	// Validate that the LedgerClassStatusType exists
	hasLedgerClassStatusType, err := k.LedgerClassStatusTypes.Has(ctx, collections.Join(l.LedgerClassId, l.StatusTypeId))
	if err != nil {
		return err
	}
	if !hasLedgerClassStatusType {
		return NewLedgerCodedError(ErrCodeInvalidField, "status_type_id")
	}

	// We omit the nftAddress out of the data we store intentionally as
	// a minor optimization since it is also our data key.
	keyStr, err := LedgerKeyToString(l.Key)
	if err != nil {
		return err
	}

	// Emit the ledger created event
	ctx.EventManager().EmitEvent(ledger.NewEventLedgerCreated(l.Key))

	// Empty out the key to avoid storing it in the ledger and the key field
	l.Key = nil

	// Store the ledger
	err = k.Ledgers.Set(ctx, *keyStr, l)
	if err != nil {
		return err
	}

	return nil
}

func (k BaseKeeper) InitGenesis(ctx sdk.Context, state *ledger.GenesisState) {
	// no-op: we start with a clean ledger state.
}

// DestroyLedger removes a ledger from the store by NFT address
func (k BaseConfigKeeper) DestroyLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, key *ledger.LedgerKey) error {
	if !k.HasLedger(ctx, key) {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger")
	}

	keyStr, err := LedgerKeyToString(key)
	if err != nil {
		return err
	}

	// Validate the authority to destroy the ledger
	// TODO: verify against registry

	// Remove the ledger from the store
	err = k.Ledgers.Remove(ctx, *keyStr)
	if err != nil {
		return err
	}

	prefix := collections.NewPrefixedPairRange[string, string](*keyStr)

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
