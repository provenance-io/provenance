package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
	"github.com/provenance-io/provenance/x/registry"
)

func (k Keeper) AddLedgerClass(ctx sdk.Context, maintainerAddr sdk.AccAddress, l types.LedgerClass) error {
	if err := types.ValidateLedgerClassBasic(&l); err != nil {
		return err
	}

	hasAssetClass := k.RegistryKeeper.AssetClassExists(ctx, &l.AssetClassId)
	if !hasAssetClass {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "asset_class_id", "asset_class doesn't exist")
	}

	has, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}
	if has {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "ledger class")
	}

	// Validate that the denom exists in the bank keeper to avoid garbage tokens being used.
	if !k.HasSupply(ctx, l.Denom) {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "denom", "denom doesn't have a supply")
	}

	// Validate that the maintainer in the ledger class is the same as the maintainer address
	// We force them to be the same for now so that a ledger class isn't locked out.
	if l.MaintainerAddress != maintainerAddr.String() {
		return types.NewLedgerCodedError(types.ErrCodeUnauthorized)
	}

	// Insert the ledger class
	err = k.LedgerClasses.Set(ctx, l.LedgerClassId, l)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) AddClassEntryType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l types.LedgerClassEntryType) error {
	if !k.IsLedgerClassMaintainer(ctx, maintainerAddr, ledgerClassId) {
		return types.NewLedgerCodedError(types.ErrCodeUnauthorized)
	}

	key := collections.Join(ledgerClassId, l.Id)

	has, err := k.LedgerClassEntryTypes.Has(ctx, key)
	if err != nil {
		return err
	}

	if has {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "ledger class entry type")
	}

	err = k.LedgerClassEntryTypes.Set(ctx, key, l)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) AddClassStatusType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l types.LedgerClassStatusType) error {
	if !k.IsLedgerClassMaintainer(ctx, maintainerAddr, ledgerClassId) {
		return types.NewLedgerCodedError(types.ErrCodeUnauthorized)
	}

	key := collections.Join(ledgerClassId, l.Id)

	has, err := k.LedgerClassStatusTypes.Has(ctx, key)
	if err != nil {
		return err
	}

	if has {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "ledger class status type")
	}

	err = k.LedgerClassStatusTypes.Set(ctx, key, l)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) AddClassBucketType(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string, l types.LedgerClassBucketType) error {
	// Bucket type validation is performed at the message level in ValidateBasic()

	// Check if the ledger class exists
	ledgerClass, err := k.GetLedgerClass(ctx, ledgerClassId)
	if err != nil {
		return err
	}
	if ledgerClass == nil {
		return types.NewLedgerCodedError(types.ErrCodeNotFound, "ledger class")
	}

	// Check if the maintainer address matches
	if ledgerClass.MaintainerAddress != maintainerAddr.String() {
		return types.NewLedgerCodedError(types.ErrCodeUnauthorized, "maintainer")
	}

	// Check if the bucket type already exists
	has, err := k.LedgerClassBucketTypes.Has(ctx, collections.Join(ledgerClassId, l.Id))
	if err != nil {
		return err
	}
	if has {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "bucket type")
	}

	// Add the bucket type
	err = k.LedgerClassBucketTypes.Set(ctx, collections.Join(ledgerClassId, l.Id), l)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) IsLedgerClassMaintainer(ctx sdk.Context, maintainerAddr sdk.AccAddress, ledgerClassId string) bool {
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

func (k Keeper) AddLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, l types.Ledger) error {
	if err := types.ValidateLedgerBasic(&l); err != nil {
		return err
	}

	if err := RequireAuthority(ctx, k.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
		AssetClassId: l.Key.AssetClassId,
		NftId:        l.Key.NftId,
	}); err != nil {
		return err
	}

	// Do not allow creating a duplicate ledger
	if k.HasLedger(ctx, l.Key) {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "ledger")
	}

	// Get the ledger class and verify that the asset class id matches
	ledgerClass, err := k.GetLedgerClass(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}
	if ledgerClass == nil {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "ledger class", "ledger class doesn't exist")
	}
	if ledgerClass.AssetClassId != l.Key.AssetClassId {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "ledger class", "ledger class not allowed for asset class id")
	}

	// Validate that the LedgerClass exists
	hasLedgerClass, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}
	if !hasLedgerClass {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "ledger class", "ledger class doesn't exist")
	}

	// Validate that the NFT exists
	if !k.RegistryKeeper.HasNFT(ctx, &l.Key.AssetClassId, &l.Key.NftId) {
		return types.NewLedgerCodedError(types.ErrCodeNotFound, "nft")
	}

	// Validate that the LedgerClassStatusType exists
	hasLedgerClassStatusType, err := k.LedgerClassStatusTypes.Has(ctx, collections.Join(l.LedgerClassId, l.StatusTypeId))
	if err != nil {
		return err
	}
	if !hasLedgerClassStatusType {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "status_type_id", "status type doesn't exist")
	}

	// We omit the nftAddress out of the data we store intentionally as
	// a minor optimization since it is also our data key.
	keyStr, err := LedgerKeyToString(l.Key)
	if err != nil {
		return err
	}

	// Emit the ledger created event
	ctx.EventManager().EmitEvent(types.NewEventLedgerCreated(l.Key))

	// Empty out the key to avoid storing it in the ledger and the key field
	l.Key = nil

	// Store the ledger
	err = k.Ledgers.Set(ctx, *keyStr, l)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) UpdateLedgerStatus(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *types.LedgerKey, statusTypeId int32) error {
	if err := RequireAuthority(ctx, k.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
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
		return types.NewLedgerCodedError(types.ErrCodeNotFound, "ledger")
	}

	// Validate that the status type id exists
	hasLedgerClassStatusType, err := k.LedgerClassStatusTypes.Has(ctx, collections.Join(ledger.LedgerClassId, statusTypeId))
	if err != nil {
		return err
	}
	if !hasLedgerClassStatusType {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "status_type_id", "status type doesn't exist")
	}

	// Update the ledger status
	ledger.StatusTypeId = statusTypeId

	// Validate ledger basic fields
	if err := types.ValidateLedgerBasic(ledger); err != nil {
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

func (k Keeper) UpdateLedgerInterestRate(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *types.LedgerKey, interestRate int32, interestDayCountConvention types.DayCountConvention, interestAccrualMethod types.InterestAccrualMethod) error {
	if err := RequireAuthority(ctx, k.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
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
		return types.NewLedgerCodedError(types.ErrCodeNotFound, "ledger")
	}

	// Update the ledger interest rate
	ledger.InterestRate = interestRate
	ledger.InterestDayCountConvention = interestDayCountConvention
	ledger.InterestAccrualMethod = interestAccrualMethod

	// Validate ledger basic fields
	if err := types.ValidateLedgerBasic(ledger); err != nil {
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

func (k Keeper) UpdateLedgerPayment(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *types.LedgerKey, nextPmtAmt int64, nextPmtDate int32, paymentFrequency types.PaymentFrequency) error {
	if err := RequireAuthority(ctx, k.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
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
		return types.NewLedgerCodedError(types.ErrCodeNotFound, "ledger")
	}

	// Update the ledger payment
	ledger.NextPmtAmt = nextPmtAmt
	ledger.NextPmtDate = nextPmtDate
	ledger.PaymentFrequency = paymentFrequency

	// Validate ledger basic fields
	if err := types.ValidateLedgerBasic(ledger); err != nil {
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

func (k Keeper) UpdateLedgerMaturityDate(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *types.LedgerKey, maturityDate int32) error {
	if err := RequireAuthority(ctx, k.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
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
		return types.NewLedgerCodedError(types.ErrCodeNotFound, "ledger")
	}

	// Update the ledger maturity date
	ledger.MaturityDate = maturityDate

	// Validate ledger basic fields
	if err := types.ValidateLedgerBasic(ledger); err != nil {
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

func (k Keeper) InitGenesis(ctx sdk.Context, state *ledger.GenesisState) {
	if state == nil {
		return
	}

	// For new chains, we only set up the basic module structure
	// Actual ledger data should be imported during upgrades, not during genesis
	// This ensures that new chains start with a clean slate and data is only
	// imported when explicitly intended through upgrade handlers
}

// DestroyLedger removes a ledger from the store by NFT address
func (k Keeper) DestroyLedger(ctx sdk.Context, authorityAddr sdk.AccAddress, lk *ledger.LedgerKey) error {
	if err := RequireAuthority(ctx, k.RegistryKeeper, authorityAddr.String(), &registry.RegistryKey{
		AssetClassId: lk.AssetClassId,
		NftId:        lk.NftId,
	}); err != nil {
		return err
	}

	if !k.HasLedger(ctx, lk) {
		return ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "ledger", "ledger doesn't exist")
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
