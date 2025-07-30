package keeper

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

func (k Keeper) AddLedgerClass(ctx sdk.Context, l types.LedgerClass) error {
	hasAssetClass := k.RegistryKeeper.AssetClassExists(ctx, &l.AssetClassId)
	if !hasAssetClass {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "asset_class_id", "asset_class doesn't exist")
	}

	has, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}

	if has {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "ledger class entry type")
	}

	// Validate that the denom exists in the bank keeper to avoid garbage tokens being used.
	if !k.HasSupply(ctx, l.Denom) {
		return types.NewLedgerCodedError(types.ErrCodeInvalidField, "denom", "denom doesn't have a supply")
	}

	// Insert the ledger class
	if err := k.LedgerClasses.Set(ctx, l.LedgerClassId, l); err != nil {
		return err
	}

	return nil
}

func (k Keeper) AddClassEntryType(ctx sdk.Context, ledgerClassId string, l types.LedgerClassEntryType) error {
	key := collections.Join(ledgerClassId, l.Id)

	has, err := k.LedgerClassEntryTypes.Has(ctx, key)
	if err != nil {
		return err
	}
	if has {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "ledger class entry type")
	}

	if err := k.LedgerClassEntryTypes.Set(ctx, key, l); err != nil {
		return err
	}

	return nil
}

func (k Keeper) AddClassStatusType(ctx sdk.Context, ledgerClassId string, l types.LedgerClassStatusType) error {
	key := collections.Join(ledgerClassId, l.Id)

	has, err := k.LedgerClassStatusTypes.Has(ctx, key)
	if err != nil {
		return err
	}

	if has {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "ledger class status type")
	}

	if err := k.LedgerClassStatusTypes.Set(ctx, key, l); err != nil {
		return err
	}

	return nil
}

func (k Keeper) AddClassBucketType(ctx sdk.Context, ledgerClassId string, l types.LedgerClassBucketType) error {
	key := collections.Join(ledgerClassId, l.Id)

	// Check if the bucket type already exists
	has, err := k.LedgerClassBucketTypes.Has(ctx, key)
	if err != nil {
		return err
	}
	if has {
		return types.NewLedgerCodedError(types.ErrCodeAlreadyExists, "bucket type")
	}

	// Add the bucket type
	if err = k.LedgerClassBucketTypes.Set(ctx, key, l); err != nil {
		return err
	}

	return nil
}

func (k Keeper) IsLedgerClassMaintainer(ctx sdk.Context, maintainerAddr string, ledgerClassId string) bool {
	// Validate that the maintainerAddr is the same as the maintainer address in the ledger class
	ledgerClass, err := k.LedgerClasses.Get(ctx, ledgerClassId)
	return err == nil && ledgerClass.MaintainerAddress == maintainerAddr
}

func (k Keeper) AddLedger(ctx sdk.Context, l types.Ledger) error {
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
	keyStr := l.Key.String()

	// Emit the ledger created event
	ctx.EventManager().EmitEvent(types.NewEventLedgerCreated(l.Key))

	// Empty out the key to avoid storing it in the ledger and the key field
	l.Key = nil

	// Store the ledger
	if err := k.Ledgers.Set(ctx, keyStr, l); err != nil {
		return err
	}

	return nil
}

func (k Keeper) UpdateLedgerStatus(ctx sdk.Context, lk *types.LedgerKey, statusTypeId int32) error {
	ledger, err := k.RequireGetLedger(ctx, lk)
	if err != nil {
		return err
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

	keyStr := ledger.Key.String()

	// Store the ledger
	if err := k.Ledgers.Set(ctx, keyStr, *ledger); err != nil {
		return err
	}

	return nil
}

func (k Keeper) UpdateLedgerInterestRate(ctx sdk.Context, lk *types.LedgerKey, interestRate int32, interestDayCountConvention types.DayCountConvention, interestAccrualMethod types.InterestAccrualMethod) error {
	ledger, err := k.RequireGetLedger(ctx, lk)
	if err != nil {
		return err
	}

	// Update the ledger interest rate
	ledger.InterestRate = interestRate
	ledger.InterestDayCountConvention = interestDayCountConvention
	ledger.InterestAccrualMethod = interestAccrualMethod

	keyStr := ledger.Key.String()

	// Store the ledger
	if err := k.Ledgers.Set(ctx, keyStr, *ledger); err != nil {
		return err
	}

	return nil
}

func (k Keeper) UpdateLedgerPayment(ctx sdk.Context, lk *types.LedgerKey, nextPmtAmt int64, nextPmtDate int32, paymentFrequency types.PaymentFrequency) error {
	ledger, err := k.RequireGetLedger(ctx, lk)
	if err != nil {
		return err
	}

	// Update the ledger payment
	ledger.NextPmtAmt = nextPmtAmt
	ledger.NextPmtDate = nextPmtDate
	ledger.PaymentFrequency = paymentFrequency

	keyStr := ledger.Key.String()

	// Store the ledger
	if err := k.Ledgers.Set(ctx, keyStr, *ledger); err != nil {
		return err
	}

	return nil
}

func (k Keeper) UpdateLedgerMaturityDate(ctx sdk.Context, lk *types.LedgerKey, maturityDate int32) error {
	ledger, err := k.RequireGetLedger(ctx, lk)
	if err != nil {
		return err
	}

	// Update the ledger maturity date
	ledger.MaturityDate = maturityDate

	keyStr := ledger.Key.String()

	// Store the ledger
	if err := k.Ledgers.Set(ctx, keyStr, *ledger); err != nil {
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
func (k Keeper) DestroyLedger(ctx sdk.Context, lk *ledger.LedgerKey) error {
	if !k.HasLedger(ctx, lk) {
		return ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "ledger", "ledger doesn't exist")
	}

	keyStr := lk.String()

	// Remove the ledger from the store
	if err := k.Ledgers.Remove(ctx, keyStr); err != nil {
		return err
	}

	prefix := collections.NewPrefixedPairRange[string, string](keyStr)

	iter, err := k.LedgerEntries.Iterate(ctx, prefix)
	if err != nil {
		return err
	}
	defer iter.Close()

	// Store the keys that we need to remove.
	keysToRemove := make([]collections.Pair[string, string], 0)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		keysToRemove = append(keysToRemove, key)
	}

	// Remove the ledger entries
	for _, key := range keysToRemove {
		if err := k.LedgerEntries.Remove(ctx, key); err != nil {
			return err
		}
	}

	return nil
}
