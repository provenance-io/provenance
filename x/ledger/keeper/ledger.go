package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
	"github.com/provenance-io/provenance/x/registry"
)

var _ ConfigKeeper = (*BaseConfigKeeper)(nil)

type ConfigKeeper interface {
	CreateLedgerClass(ctx sdk.Context, maintainerAddr sdk.AccAddress, l ledger.LedgerClass) error
	AddClassEntryType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l ledger.LedgerClassEntryType) error
	AddClassStatusType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l ledger.LedgerClassStatusType) error
	AddClassBucketType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l ledger.LedgerClassBucketType) error

	CreateLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, l ledger.Ledger) error

	UpdateLedgerStatus(ctx sdk.Context, authorityAddr sdk.AccAddress, key *ledger.LedgerKey, statusTypeId int32) error
	UpdateLedgerInterestRate(ctx sdk.Context, authorityAddr sdk.AccAddress, key *ledger.LedgerKey, interestRate int32, interestDayCountConvention ledger.DayCountConvention, interestAccrualMethod ledger.InterestAccrualMethod) error
	UpdateLedgerPayment(ctx sdk.Context, authorityAddr sdk.AccAddress, key *ledger.LedgerKey, nextPmtAmt int64, nextPmtDate int32, paymentFrequency ledger.PaymentFrequency) error
	UpdateLedgerMaturityDate(ctx sdk.Context, authorityAddr sdk.AccAddress, key *ledger.LedgerKey, maturityDate int32) error
	DestroyLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, key *ledger.LedgerKey) error
}

type BaseConfigKeeper struct {
	BaseViewKeeper
	BankKeeper
}

func (k BaseConfigKeeper) CreateLedgerClass(ctx sdk.Context, maintainerAddr sdk.AccAddress, l ledger.LedgerClass) error {
	if err := ValidateLedgerClassBasic(&l); err != nil {
		return err
	}

	hasAssetClass := k.BaseViewKeeper.RegistryKeeper.AssetClassExists(ctx, &l.AssetClassId)
	if !hasAssetClass {
		return NewLedgerCodedError(ErrCodeInvalidField, "asset_class_id", "asset_class already exists")
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
		return NewLedgerCodedError(ErrCodeInvalidField, "denom", "denom doesn't have a supply")
	}

	// Validate that the maintainer in the ledger class is the same as the maintainer address
	// We force them to be the same for now so that a ledger class isn't locked out.
	if l.MaintainerAddress != maintainerAddr.String() {
		return NewLedgerCodedError(ErrCodeUnauthorized)
	}

	// Insert the ledger class
	err = k.LedgerClasses.Set(ctx, l.LedgerClassId, l)
	if err != nil {
		return err
	}

	return nil
}

func (k BaseConfigKeeper) AddClassEntryType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l ledger.LedgerClassEntryType) error {
	if !k.IsLedgerClassMaintainer(ctx, maintainerAddr, ledgerClassId) {
		return NewLedgerCodedError(ErrCodeUnauthorized)
	}

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

func (k BaseConfigKeeper) AddClassStatusType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l ledger.LedgerClassStatusType) error {
	if !k.IsLedgerClassMaintainer(ctx, maintainerAddr, ledgerClassId) {
		return NewLedgerCodedError(ErrCodeUnauthorized)
	}

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

func (k BaseConfigKeeper) AddClassBucketType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l ledger.LedgerClassBucketType) error {
	if !k.IsLedgerClassMaintainer(ctx, maintainerAddr, ledgerClassId) {
		return NewLedgerCodedError(ErrCodeUnauthorized)
	}

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

func (k BaseConfigKeeper) IsLedgerClassMaintainer(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string) bool {
	// Validate that the maintainerAddr is the same as the maintainer address in the ledger class
	ledgerClass, err := k.LedgerClasses.Get(ctx, ledgerClassId)
	if err != nil {
		return false
	}

	if ledgerClass.MaintainerAddress == maintainerAddr.String() {
		return true
	}

	return false
}

func (k BaseConfigKeeper) CreateLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, l ledger.Ledger) error {
	if err := ValidateLedgerBasic(&l); err != nil {
		return err
	}

	if err := RequireAuthority(ctx, k.BaseViewKeeper.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
		AssetClassId: l.Key.AssetClassId,
		NftId:        l.Key.NftId,
	}); err != nil {
		return err
	}

	// Do not allow creating a duplicate ledger
	if k.HasLedger(ctx, l.Key) {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "ledger")
	}

	// Get the ledger class and verify that the asset class id matches
	ledgerClass, err := k.GetLedgerClass(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}
	if ledgerClass == nil {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger class", "ledger class doesn't exist")
	}
	if ledgerClass.AssetClassId != l.Key.AssetClassId {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger class", "ledger class not allowed for asset class id")
	}

	// Validate that the LedgerClass exists
	hasLedgerClass, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}
	if !hasLedgerClass {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger class", "ledger class doesn't exist")
	}

	// Validate that the NFT exists
	if !k.BaseViewKeeper.RegistryKeeper.HasNFT(ctx, &l.Key.AssetClassId, &l.Key.NftId) {
		return NewLedgerCodedError(ErrCodeNotFound, "nft")
	}

	// Validate that the LedgerClassStatusType exists
	hasLedgerClassStatusType, err := k.LedgerClassStatusTypes.Has(ctx, collections.Join(l.LedgerClassId, l.StatusTypeId))
	if err != nil {
		return err
	}
	if !hasLedgerClassStatusType {
		return NewLedgerCodedError(ErrCodeInvalidField, "status_type_id", "status type doesn't exist")
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

func (k BaseConfigKeeper) UpdateLedgerStatus(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *ledger.LedgerKey, statusTypeId int32) error {
	if err := RequireAuthority(ctx, k.BaseViewKeeper.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
		AssetClassId: lk.AssetClassId,
		NftId:        lk.NftId,
	}); err != nil {
		return err
	}

	// Get the ledger
	ledger, err := k.GetLedger(ctx, lk)
	if err != nil {
		return err
	}
	if ledger == nil {
		return NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	// Validate that the status type id exists
	hasLedgerClassStatusType, err := k.LedgerClassStatusTypes.Has(ctx, collections.Join(ledger.LedgerClassId, statusTypeId))
	if err != nil {
		return err
	}
	if !hasLedgerClassStatusType {
		return NewLedgerCodedError(ErrCodeInvalidField, "status_type_id", "status type doesn't exist")
	}

	// Update the ledger status
	ledger.StatusTypeId = statusTypeId

	// Validate ledger basic fields
	if err := ValidateLedgerBasic(ledger); err != nil {
		return err
	}

	keyStr, err := LedgerKeyToString(ledger.Key)
	if err != nil {
		return err
	}

	// Store the ledger
	err = k.Ledgers.Set(ctx, *keyStr, *ledger)
	if err != nil {
		return err
	}
	return fmt.Errorf("not implemented")
}

func (k BaseConfigKeeper) UpdateLedgerInterestRate(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *ledger.LedgerKey, interestRate int32, interestDayCountConvention ledger.DayCountConvention, interestAccrualMethod ledger.InterestAccrualMethod) error {
	if err := RequireAuthority(ctx, k.BaseViewKeeper.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
		AssetClassId: lk.AssetClassId,
		NftId:        lk.NftId,
	}); err != nil {
		return err
	}

	// Get the ledger
	ledger, err := k.GetLedger(ctx, lk)
	if err != nil {
		return err
	}
	if ledger == nil {
		return NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	// Update the ledger interest rate
	ledger.InterestRate = interestRate
	ledger.InterestDayCountConvention = interestDayCountConvention
	ledger.InterestAccrualMethod = interestAccrualMethod

	// Validate ledger basic fields
	if err := ValidateLedgerBasic(ledger); err != nil {
		return err
	}

	keyStr, err := LedgerKeyToString(ledger.Key)
	if err != nil {
		return err
	}

	// Store the ledger
	err = k.Ledgers.Set(ctx, *keyStr, *ledger)
	if err != nil {
		return err
	}

	return nil
}

func (k BaseConfigKeeper) UpdateLedgerPayment(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *ledger.LedgerKey, nextPmtAmt int64, nextPmtDate int32, paymentFrequency ledger.PaymentFrequency) error {
	if err := RequireAuthority(ctx, k.BaseViewKeeper.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
		AssetClassId: lk.AssetClassId,
		NftId:        lk.NftId,
	}); err != nil {
		return err
	}

	// Get the ledger
	ledger, err := k.GetLedger(ctx, lk)
	if err != nil {
		return err
	}
	if ledger == nil {
		return NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	// Update the ledger payment
	ledger.NextPmtAmt = nextPmtAmt
	ledger.NextPmtDate = nextPmtDate
	ledger.PaymentFrequency = paymentFrequency

	// Validate ledger basic fields
	if err := ValidateLedgerBasic(ledger); err != nil {
		return err
	}

	keyStr, err := LedgerKeyToString(ledger.Key)
	if err != nil {
		return err
	}

	// Store the ledger
	err = k.Ledgers.Set(ctx, *keyStr, *ledger)
	if err != nil {
		return err
	}

	return nil
}

func (k BaseConfigKeeper) UpdateLedgerMaturityDate(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *ledger.LedgerKey, maturityDate int32) error {
	if err := RequireAuthority(ctx, k.BaseViewKeeper.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
		AssetClassId: lk.AssetClassId,
		NftId:        lk.NftId,
	}); err != nil {
		return err
	}

	// Get the ledger
	ledger, err := k.GetLedger(ctx, lk)
	if err != nil {
		return err
	}
	if ledger == nil {
		return NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	// Update the ledger maturity date
	ledger.MaturityDate = maturityDate

	// Validate ledger basic fields
	if err := ValidateLedgerBasic(ledger); err != nil {
		return err
	}

	keyStr, err := LedgerKeyToString(ledger.Key)
	if err != nil {
		return err
	}

	// Store the ledger
	err = k.Ledgers.Set(ctx, *keyStr, *ledger)
	if err != nil {
		return err
	}

	return nil
}

func (k BaseKeeper) InitGenesis(ctx sdk.Context, state *ledger.GenesisState) {
	// no-op: we start with a clean ledger state.
}

// DestroyLedger removes a ledger from the store by NFT address
func (k BaseConfigKeeper) DestroyLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *ledger.LedgerKey) error {
	if err := RequireAuthority(ctx, k.BaseViewKeeper.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
		AssetClassId: lk.AssetClassId,
		NftId:        lk.NftId,
	}); err != nil {
		return err
	}

	if !k.HasLedger(ctx, lk) {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger", "ledger doesn't exist")
	}

	keyStr, err := LedgerKeyToString(lk)
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
