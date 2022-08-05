package keeper

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/codec"
	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/provenance-io/provenance/x/expiration/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Handler is a name record handler function for use with IterateRecords.
type Handler func(record types.Expiration) error

// Keeper defines the name module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set account specific params
	paramSpace paramtypes.Subspace

	// Key to access the key-value store from sdk.Context.
	storeKey sdk.StoreKey

	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec
}

// NewKeeper returns an expiration keeper. It handles:
// - managing a hierarchy of expiration
// - enforcing permissions for expiration creation/extension/deletion
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key sdk.StoreKey,
	paramSpace paramtypes.Subspace,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   key,
		paramSpace: paramSpace,
		cdc:        cdc,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetDeposit returns the default deposit used in setting module asset expirations
func (k Keeper) GetDeposit(ctx sdk.Context) sdk.Coin {
	deposit := types.DefaultDeposit
	if k.paramSpace.Has(ctx, types.ParamStoreKeyDeposit) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyDeposit, deposit)
	}
	return deposit
}

// SetExpiration binds a name to an address.
func (k Keeper) SetExpiration(
	ctx sdk.Context,
	moduleAssetId string,
	owner string,
	expirationHeight int64,
	deposit sdk.Coin,
	expirationMessages []*types2.Any,
	event sdk.Event,
) error {
	moduleAssetKey, err := types.GetModuleAssetKeyPrefix(moduleAssetId)
	if err != nil {
		return err
	}
	ownerKey, err := types.GetOwnerKeyIndexPrefix(moduleAssetId, owner)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	record := types.Expiration{
		ModuleAssetId:      moduleAssetId,
		Owner:              owner,
		ExpirationHeight:   expirationHeight,
		Deposit:            deposit,
		ExpirationMessages: expirationMessages,
	}
	if err = record.ValidateBasic(); err != nil {
		return err
	}
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	// index by module asset
	store.Set(moduleAssetKey, bz)
	// index by owner
	store.Set(ownerKey, bz)

	nameBoundEvent := types.NewEventNameBound(record.Address, name, record.Restricted)

	if err := ctx.EventManager().EmitTypedEvent(nameBoundEvent); err != nil {
		return err
	}

	return nil
}

// GetExpirationByModuleAssetId resolves a record by module asset id.
func (k Keeper) GetExpirationByModuleAssetId(ctx sdk.Context, moduleAssetId string) (record *types.Expiration, err error) {
	key, err := types.GetModuleAssetKeyPrefix(moduleAssetId)
	if err != nil {
		return nil, err
	}
	return getExpiration(ctx, k, key)
}

// GetExpirationByOwner resolves a record by owner.
func (k Keeper) GetExpirationsByOwner(ctx sdk.Context, owner string) (record *types.Expiration, err error) {
	key, err := types.GetOwnerKeyPrefix(owner)
	if err != nil {
		return nil, err
	}
	return getExpiration(ctx, k, key)
}

func getExpiration(ctx sdk.Context, keeper Keeper, key []byte) (record *types.Expiration, err error) {
	store := ctx.KVStore(keeper.storeKey)
	if !store.Has(key) {
		return nil, types.ErrExpirationNotfound
	}
	bz := store.Get(key)
	record = &types.Expiration{}
	err = keeper.cdc.Unmarshal(bz, record)
	return record, err
}

// ExpirationExists returns true if store contains a record for the given name.
func (k Keeper) ExpirationExists(ctx sdk.Context, moduleAssetId string) bool {
	key, err := types.GetModuleAssetKeyPrefix(moduleAssetId)
	if err != nil {
		return false
	}
	store := ctx.KVStore(k.storeKey)
	return store.Has(key)
}

// GetExpirationsByOwner looks up all names bound to an address.
func (k Keeper) GetExpirationsByOwner(ctx sdk.Context, owner string) ([]types.Expiration, error) {
	// Return value data structure.
	records := make([]types.Expiration, 0)
	// Handler that adds records if owner address matches.
	recordsHandler := func(record types.Expiration) error {
		if record.Owner == owner {
			records = append(records, record)
		}
		return nil
	}
	// Calculate owner key prefix
	key, err := types.GetOwnerKeyPrefix(owner)
	if err != nil {
		return nil, err
	}
	// Collect and return all names that match.
	if err := k.IterateRecords(ctx, key, recordsHandler); err != nil {
		return records, err
	}
	return records, nil
}

// DeleteExpiration removes an expiration record from the kvstore.
func (k Keeper) DeleteExpiration(ctx sdk.Context, moduleAssetId string) error {
	// Need the record to clear the address index
	record, err := k.GetExpirationByModuleAssetId(ctx, moduleAssetId)
	if err != nil {
		return err
	}
	// Delete the module asset index record
	moduleAssetKey, err := types.GetModuleAssetKeyPrefix(moduleAssetId)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(moduleAssetKey)
	// Delete the owner index record
	ownerKey, err := types.GetOwnerKeyPrefix(record.Owner)
	if err != nil {
		return err
	}
	if store.Has(ownerKey) {
		store.Delete(ownerKey)
	}

	nameUnboundEvent := types.NewEventNameUnbound(record.Address, name, record.Restricted)

	if err := ctx.EventManager().EmitTypedEvent(nameUnboundEvent); err != nil {
		return err
	}

	return nil
}

// IterateRecords iterates over all the stored name records and passes them to a callback function.
func (k Keeper) IterateRecords(ctx sdk.Context, prefix []byte, handle Handler) error {
	// Init a name record iterator
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	// Iterate over records, processing callbacks.
	for ; iterator.Valid(); iterator.Next() {
		record := types.NameRecord{}
		// get proto objects for legacy prefix with legacy amino codec.
		if bytes.Equal(prefix, types.NameKeyPrefixAmino) {
			if err := types.ModuleCdc.Unmarshal(iterator.Value(), &record); err != nil {
				return err
			}
		} else {
			if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
				return err
			}
		}
		if err := handle(record); err != nil {
			return err
		}
	}
	return nil
}
