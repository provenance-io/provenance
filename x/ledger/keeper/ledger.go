package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// AddLedgerClass creates a new ledger class with validation checks.
// This function validates that the asset class exists in the registry and that the denom has a supply.
// It ensures that ledger classes are only created for valid asset classes and tokens.
func (k Keeper) AddLedgerClass(ctx sdk.Context, l types.LedgerClass) error {
	// Check if the asset class exists in the registry.
	hasAssetClass := k.RegistryKeeper.AssetClassExists(ctx, &l.AssetClassId)
	if !hasAssetClass {
		return types.NewErrCodeInvalidField("asset_class_id", "asset_class doesn't exist")
	}

	// Check if the ledger class already exists to prevent duplicates.
	has, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}

	if has {
		return types.NewErrCodeAlreadyExists("ledger class entry type")
	}

	// Validate that the denom exists in the bank keeper to avoid garbage tokens being used.
	if !k.BankKeeper.HasSupply(ctx, l.Denom) {
		return types.NewErrCodeInvalidField("denom", "denom doesn't have a supply")
	}

	// Insert the ledger class into the state store.
	if err := k.LedgerClasses.Set(ctx, l.LedgerClassId, l); err != nil {
		return err
	}

	return nil
}

// AddClassEntryType adds a new entry type to an existing ledger class.
// This function validates that the entry type doesn't already exist for the given class.
// Entry types define what kinds of transactions can be recorded in ledgers of this class.
func (k Keeper) AddClassEntryType(ctx sdk.Context, ledgerClassID string, l types.LedgerClassEntryType) error {
	// Create the composite key for the entry type.
	key := collections.Join(ledgerClassID, l.Id)

	// Check if the entry type already exists to prevent duplicates.
	has, err := k.LedgerClassEntryTypes.Has(ctx, key)
	if err != nil {
		return err
	}
	if has {
		return types.NewErrCodeAlreadyExists("ledger class entry type")
	}

	// Store the entry type in the state store.
	if err := k.LedgerClassEntryTypes.Set(ctx, key, l); err != nil {
		return err
	}

	return nil
}

// AddClassStatusType adds a new status type to an existing ledger class.
// This function validates that the status type doesn't already exist for the given class.
// Status types define the possible states that ledger entries can have.
func (k Keeper) AddClassStatusType(ctx sdk.Context, ledgerClassID string, l types.LedgerClassStatusType) error {
	// Create the composite key for the status type.
	key := collections.Join(ledgerClassID, l.Id)

	// Check if the status type already exists to prevent duplicates.
	has, err := k.LedgerClassStatusTypes.Has(ctx, key)
	if err != nil {
		return err
	}

	if has {
		return types.NewErrCodeAlreadyExists("ledger class status type")
	}

	// Store the status type in the state store.
	if err := k.LedgerClassStatusTypes.Set(ctx, key, l); err != nil {
		return err
	}

	return nil
}

// AddClassBucketType adds a new bucket type to an existing ledger class.
// This function validates that the bucket type doesn't already exist for the given class.
// Bucket types define how funds are categorized and organized within ledgers.
func (k Keeper) AddClassBucketType(ctx sdk.Context, ledgerClassID string, l types.LedgerClassBucketType) error {
	// Create the composite key for the bucket type.
	key := collections.Join(ledgerClassID, l.Id)

	// Check if the bucket type already exists to prevent duplicates.
	has, err := k.LedgerClassBucketTypes.Has(ctx, key)
	if err != nil {
		return err
	}
	if has {
		return types.NewErrCodeAlreadyExists("bucket type")
	}

	// Add the bucket type to the state store.
	if err = k.LedgerClassBucketTypes.Set(ctx, key, l); err != nil {
		return err
	}

	return nil
}

// IsLedgerClassMaintainer checks if the given address is the maintainer of a ledger class.
// This function validates that the maintainer address matches the one stored in the ledger class.
// Only maintainers can modify ledger class configurations.
func (k Keeper) IsLedgerClassMaintainer(ctx sdk.Context, maintainerAddr string, ledgerClassID string) bool {
	// Validate that the maintainer address matches the one in the ledger class.
	ledgerClass, err := k.LedgerClasses.Get(ctx, ledgerClassID)
	return err == nil && ledgerClass.MaintainerAddress == maintainerAddr
}

// AddLedger creates a new ledger instance with comprehensive validation.
// This function validates the ledger class, asset class, NFT existence, and status type.
// It ensures that ledgers are only created for valid configurations and existing assets.
func (k Keeper) AddLedger(ctx sdk.Context, l types.Ledger) error {
	// Do not allow creating a duplicate ledger.
	if k.HasLedger(ctx, l.Key) {
		return types.NewErrCodeAlreadyExists("ledger")
	}

	// Get the ledger class and verify that the asset class id matches.
	ledgerClass, err := k.RequireGetLedgerClass(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}

	// Ensure the ledger class is compatible with the asset class.
	if ledgerClass.AssetClassId != l.Key.AssetClassId {
		return types.NewErrCodeInvalidField("ledger class", "ledger class not allowed for asset class id")
	}

	// Validate that the ledger class exists in the state store.
	hasLedgerClass, err := k.LedgerClasses.Has(ctx, l.LedgerClassId)
	if err != nil {
		return err
	}
	if !hasLedgerClass {
		return types.NewErrCodeInvalidField("ledger class", "ledger class doesn't exist")
	}

	// Validate that the NFT exists in the registry.
	if !k.RegistryKeeper.HasNFT(ctx, &l.Key.AssetClassId, &l.Key.NftId) {
		return types.NewErrCodeNotFound("nft")
	}

	// Validate that the status type exists for this ledger class.
	hasLedgerClassStatusType, err := k.LedgerClassStatusTypes.Has(ctx, collections.Join(l.LedgerClassId, l.StatusTypeId))
	if err != nil {
		return err
	}
	if !hasLedgerClassStatusType {
		return types.NewErrCodeInvalidField("status_type_id", "status type doesn't exist")
	}

	// We omit the nftAddress out of the data we store intentionally as a minor optimization since it is also our data key.
	keyStr := l.Key.String()

	// Empty out the key to avoid storing it in the ledger and the key field.
	key := l.Key
	l.Key = nil

	// Store the ledger in the state store.
	if err := k.Ledgers.Set(ctx, keyStr, l); err != nil {
		return err
	}

	// Emit the ledger created event to notify other modules.
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventLedgerCreated(key)); err != nil {
		return err
	}

	return nil
}

// UpdateLedgerStatus updates the status type of an existing ledger.
// This function validates that the new status type exists for the ledger's class.
// Status changes can reflect the current state of the ledger (e.g., active, suspended, closed).
func (k Keeper) UpdateLedgerStatus(ctx sdk.Context, lk *types.LedgerKey, statusTypeID int32) error {
	// Retrieve the existing ledger to ensure it exists.
	ledger, err := k.RequireGetLedger(ctx, lk)
	if err != nil {
		return err
	}

	// Validate that the status type exists for this ledger class.
	hasLedgerClassStatusType, err := k.LedgerClassStatusTypes.Has(ctx, collections.Join(ledger.LedgerClassId, statusTypeID))
	if err != nil {
		return err
	}
	if !hasLedgerClassStatusType {
		return types.NewErrCodeInvalidField("status_type_id", "status type doesn't exist")
	}

	// Update the ledger status with the new status type.
	ledger.StatusTypeId = statusTypeID

	keyStr := ledger.Key.String()

	// Empty out the key to avoid storing it in the ledger and the key field.
	key := ledger.Key
	ledger.Key = nil

	// Store the updated ledger in the state store.
	if err := k.Ledgers.Set(ctx, keyStr, *ledger); err != nil {
		return err
	}

	// Emit the ledger updated event.
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventLedgerUpdated(key, types.UpdateType_UPDATE_TYPE_STATUS)); err != nil {
		return err
	}

	return nil
}

// UpdateLedgerInterestRate updates the interest rate configuration of an existing ledger.
// This function allows modification of the interest rate, day count convention, and accrual method.
// These parameters affect how interest is calculated and applied to the ledger.
func (k Keeper) UpdateLedgerInterestRate(ctx sdk.Context, lk *types.LedgerKey, interestRate int32, interestDayCountConvention types.DayCountConvention, interestAccrualMethod types.InterestAccrualMethod) error {
	// Retrieve the existing ledger to ensure it exists.
	ledger, err := k.RequireGetLedger(ctx, lk)
	if err != nil {
		return err
	}

	// Update the ledger interest rate configuration.
	ledger.InterestRate = interestRate
	ledger.InterestDayCountConvention = interestDayCountConvention
	ledger.InterestAccrualMethod = interestAccrualMethod

	keyStr := ledger.Key.String()

	// Empty out the key to avoid storing it in the ledger and the key field.
	key := ledger.Key
	ledger.Key = nil

	// Store the updated ledger in the state store.
	if err := k.Ledgers.Set(ctx, keyStr, *ledger); err != nil {
		return err
	}

	// Emit the ledger updated event.
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventLedgerUpdated(key, types.UpdateType_UPDATE_TYPE_INTEREST_RATE)); err != nil {
		return err
	}

	return nil
}

// UpdateLedgerPayment updates the payment configuration of an existing ledger.
// This function allows modification of the next payment amount, date, and frequency.
// These parameters define the payment schedule for the ledger.
func (k Keeper) UpdateLedgerPayment(ctx sdk.Context, lk *types.LedgerKey, nextPmtAmt int64, nextPmtDate int32, paymentFrequency types.PaymentFrequency) error {
	// Retrieve the existing ledger to ensure it exists.
	ledger, err := k.RequireGetLedger(ctx, lk)
	if err != nil {
		return err
	}

	// Update the ledger payment configuration.
	ledger.NextPmtAmt = nextPmtAmt
	ledger.NextPmtDate = nextPmtDate
	ledger.PaymentFrequency = paymentFrequency

	keyStr := ledger.Key.String()

	// Empty out the key to avoid storing it in the ledger and the key field.
	key := ledger.Key
	ledger.Key = nil

	// Store the updated ledger in the state store.
	if err := k.Ledgers.Set(ctx, keyStr, *ledger); err != nil {
		return err
	}

	// Emit the ledger updated event.
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventLedgerUpdated(key, types.UpdateType_UPDATE_TYPE_PAYMENT)); err != nil {
		return err
	}

	return nil
}

// UpdateLedgerMaturityDate updates the maturity date of an existing ledger.
// This function allows modification of when the ledger reaches its final maturity.
// The maturity date is important for calculating final payments and ledger closure.
func (k Keeper) UpdateLedgerMaturityDate(ctx sdk.Context, lk *types.LedgerKey, maturityDate int32) error {
	// Retrieve the existing ledger to ensure it exists.
	ledger, err := k.RequireGetLedger(ctx, lk)
	if err != nil {
		return err
	}

	// Update the ledger maturity date.
	ledger.MaturityDate = maturityDate

	keyStr := ledger.Key.String()

	// Empty out the key to avoid storing it in the ledger and the key field.
	key := ledger.Key
	ledger.Key = nil

	// Store the updated ledger in the state store.
	if err := k.Ledgers.Set(ctx, keyStr, *ledger); err != nil {
		return err
	}

	// Emit the ledger updated event.
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventLedgerUpdated(key, types.UpdateType_UPDATE_TYPE_MATURITY_DATE)); err != nil {
		return err
	}

	return nil
}

// DestroyLedger removes a ledger and all its associated entries from the state store.
// This function performs a complete cleanup by removing the ledger and all its entries.
// It emits an event to notify other modules of the ledger destruction.
func (k Keeper) DestroyLedger(ctx sdk.Context, lk *types.LedgerKey) error {
	// Check if the ledger exists before attempting to destroy it.
	if !k.HasLedger(ctx, lk) {
		return types.NewErrCodeInvalidField("ledger", "ledger doesn't exist")
	}

	keyStr := lk.String()

	// Remove the ledger from the state store.
	if err := k.Ledgers.Remove(ctx, keyStr); err != nil {
		return err
	}

	// Create a prefix to find all entries associated with this ledger.
	prefix := collections.NewPrefixedPairRange[string, string](keyStr)

	// Iterate through all entries to collect their keys for removal.
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

	// Remove all ledger entries associated with this ledger.
	for _, key := range keysToRemove {
		if err := k.LedgerEntries.Remove(ctx, key); err != nil {
			return err
		}
	}

	// Emit the ledger destroyed event to notify other modules.
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventLedgerDestroyed(lk)); err != nil {
		return err
	}

	return nil
}

// GetLedger retrieves a ledger by its NFT address.
// This function looks up a ledger using the ledger key and returns the complete ledger data.
// The NFT address is added back to the ledger since it's not stored in the state.
//
// Parameters:
//   - ctx: The SDK context
//   - key: The ledger key containing asset class ID and NFT ID
//
// Returns:
//   - *ledger.Ledger: A pointer to the found ledger, or nil if not found
//   - error: Any error that occurred during retrieval, or nil if successful
//
// Behavior:
//   - Returns (nil, nil) if the ledger is not found
//   - Returns (nil, err) if an error occurs during retrieval
//   - Returns (&ledger, nil) if the ledger is found successfully
//   - The returned ledger will have its Key field set to the provided key
func (k Keeper) GetLedger(ctx sdk.Context, key *types.LedgerKey) (*types.Ledger, error) {
	keyStr := key.String()

	// Lookup the ledger in the state store using the key string.
	l, err := k.Ledgers.Get(ctx, keyStr)
	if err != nil {
		// Eat the not found error as it is expected, and return nil.
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}

		// Otherwise, return the error.
		return nil, err
	}

	// The key isn't stored in the ledger, so we add it back in.
	l.Key = key
	return &l, nil
}

// RequireGetLedger retrieves a ledger and requires it to exist.
// This function is similar to GetLedger but returns an error if the ledger is not found.
// It's used when the ledger must exist for the operation to proceed.
func (k Keeper) RequireGetLedger(ctx sdk.Context, lk *types.LedgerKey) (*types.Ledger, error) {
	ledger, err := k.GetLedger(ctx, lk)
	if err != nil {
		return nil, err
	}
	if ledger == nil {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	return ledger, nil
}

// HasLedger checks if a ledger exists for the given key.
// This function provides a quick existence check without retrieving the full ledger data.
func (k Keeper) HasLedger(ctx sdk.Context, key *types.LedgerKey) bool {
	keyStr := key.String()

	has, _ := k.Ledgers.Has(ctx, keyStr)
	return has
}

// GetLedgerClass retrieves a ledger class by its ID.
// This function looks up a ledger class and returns the complete class configuration.
// It returns nil if the ledger class doesn't exist.
func (k Keeper) GetLedgerClass(ctx context.Context, ledgerClassID string) (*types.LedgerClass, error) {
	ledgerClass, err := k.LedgerClasses.Get(ctx, ledgerClassID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ledgerClass, nil
}

// RequireGetLedgerClass retrieves a ledger class and requires it to exist.
// This function is similar to GetLedgerClass but returns an error if the class is not found.
// It's used when the ledger class must exist for the operation to proceed.
func (k Keeper) RequireGetLedgerClass(ctx context.Context, ledgerClassID string) (*types.LedgerClass, error) {
	ledgerClass, err := k.GetLedgerClass(ctx, ledgerClassID)
	if err != nil {
		return nil, err
	}
	if ledgerClass == nil {
		return nil, types.NewErrCodeNotFound("ledger class")
	}
	return ledgerClass, nil
}

// GetAllLedgerClasses retrieves all ledger classes with pagination support.
// This function provides paginated access to all ledger classes in the system.
// It uses the collections.Paginate function to handle pagination efficiently.
func (k Keeper) GetAllLedgerClasses(ctx context.Context, pageRequest *query.PageRequest) ([]*types.LedgerClass, *query.PageResponse, error) {
	// Use query.CollectionPaginate to handle pagination for the ledger classes collection.
	ledgerClasses, pageRes, err := query.CollectionPaginate(ctx, k.LedgerClasses, pageRequest, func(_ string, value types.LedgerClass) (*types.LedgerClass, error) {
		return &value, nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Ensure ledgerClasses is never nil, return empty slice instead
	if ledgerClasses == nil {
		ledgerClasses = []*types.LedgerClass{}
	}

	return ledgerClasses, pageRes, nil
}

// GetLedgerClassEntryTypes retrieves all entry types for a given ledger class.
// This function walks through all entry type definitions associated with the ledger class.
// Entry types define what kinds of transactions can be recorded in ledgers of this class.
func (k Keeper) GetLedgerClassEntryTypes(ctx context.Context, ledgerClassID string) ([]*types.LedgerClassEntryType, error) {
	// Create a prefix range to find all entry types for this ledger class.
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassID)

	// Initialize a slice to collect all entry types.
	entryTypes := make([]*types.LedgerClassEntryType, 0)

	// Walk through all entry type records that match the ledger class prefix.
	err := k.LedgerClassEntryTypes.Walk(ctx, prefix, func(_ collections.Pair[string, int32], value types.LedgerClassEntryType) (stop bool, err error) {
		entryTypes = append(entryTypes, &value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return entryTypes, nil
}

// GetLedgerClassStatusTypes retrieves all status types for a given ledger class.
// This function walks through all status type definitions associated with the ledger class.
// Status types define the possible states that ledger entries can have.
func (k Keeper) GetLedgerClassStatusTypes(ctx context.Context, ledgerClassID string) ([]*types.LedgerClassStatusType, error) {
	// Create a prefix range to find all status types for this ledger class.
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassID)

	// Initialize a slice to collect all status types.
	statusTypes := make([]*types.LedgerClassStatusType, 0)

	// Walk through all status type records that match the ledger class prefix.
	err := k.LedgerClassStatusTypes.Walk(ctx, prefix, func(_ collections.Pair[string, int32], value types.LedgerClassStatusType) (stop bool, err error) {
		statusTypes = append(statusTypes, &value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return statusTypes, nil
}

// GetLedgerClassBucketTypes retrieves all bucket types for a given ledger class.
// This function walks through all bucket type definitions associated with the ledger class.
// Bucket types define how funds are categorized and organized within ledgers.
func (k Keeper) GetLedgerClassBucketTypes(ctx context.Context, ledgerClassID string) ([]*types.LedgerClassBucketType, error) {
	// Create a prefix range to find all bucket types for this ledger class.
	prefix := collections.NewPrefixedPairRange[string, int32](ledgerClassID)

	// Initialize a slice to collect all bucket types.
	bucketTypes := make([]*types.LedgerClassBucketType, 0)

	// Walk through all bucket type records that match the ledger class prefix.
	err := k.LedgerClassBucketTypes.Walk(ctx, prefix, func(_ collections.Pair[string, int32], value types.LedgerClassBucketType) (stop bool, err error) {
		bucketTypes = append(bucketTypes, &value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return bucketTypes, nil
}
