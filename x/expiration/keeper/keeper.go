package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gogo/protobuf/proto"
	"github.com/provenance-io/provenance/x/expiration/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Handler is a name record handler function for use with IterateExpirations.
type Handler func(record types.Expiration) error

// Keeper defines the name module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set account specific params
	paramSpace paramtypes.Subspace

	// Key to access the key-value store from sdk.Context.
	storeKey sdk.StoreKey

	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// Used to ensure accounts exist for addresses
	authKeeper authkeeper.AccountKeeper

	// To check granter grantee authorization of messages
	authzKeeper authzkeeper.Keeper
}

// todo: we'll most likely need authz keeper for granter/grantee stuff
// 		other module keepers as we?

// NewKeeper returns an expiration keeper. It handles:
// - managing a hierarchy of expiration
// - enforcing permissions for expiration creation/extension/deletion
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	authKeeper authkeeper.AccountKeeper,
	authzKeeper authzkeeper.Keeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:    key,
		paramSpace:  paramSpace,
		cdc:         cdc,
		authKeeper:  authKeeper,
		authzKeeper: authzKeeper,
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

// GetExpiration returns the expiration with the given module asset id.
func (k Keeper) GetExpiration(ctx sdk.Context, moduleAssetId string) (*types.Expiration, error) {
	key, err := types.GetModuleAssetKeyPrefix(moduleAssetId)
	if err != nil {
		return nil, err
	}

	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return nil, types.ErrExpirationNotFound
	}

	b := store.Get(key)
	expiration := &types.Expiration{}
	err = k.cdc.Unmarshal(b, expiration)
	if err != nil {
		return nil, err
	}

	return expiration, nil
}

// SetExpiration creates an expiration record for a module asset
func (k Keeper) SetExpiration(ctx sdk.Context, expiration types.Expiration) error {
	// Run basic expiration data validation
	if err := expiration.ValidateBasic(); err != nil {
		return err
	}

	// Verify owner account exists for the given owner address
	if err := k.isValidOwnerAddress(ctx, expiration.Owner); err != nil {
		k.Logger(ctx).Error("invalid owner address", "err", err, "expiration", expiration.String())
		return err
	}

	// get store key prefix
	store := ctx.KVStore(k.storeKey)
	key, err := types.GetModuleAssetKeyPrefix(expiration.ModuleAssetId)
	if err != nil {
		return err
	}
	// Validate module asset exists
	if err := k.isValidModuleAsset(ctx, expiration.ModuleAssetId); err != nil {

	}
	b, err := k.cdc.Marshal(&expiration)
	if err != nil {
		return err
	}
	store.Set(key, b)

	// emit Add event
	addEvent := types.NewEventExpirationAdd(expiration.ModuleAssetId)
	return k.emitEvent(ctx, addEvent)
}

func (k Keeper) UpdateExpiration(ctx sdk.Context, expiration types.Expiration) error {
	// Run basic expiration data validation
	if err := expiration.ValidateBasic(); err != nil {
		return err
	}

	// get key prefix
	key, err := types.GetModuleAssetKeyPrefix(expiration.ModuleAssetId)
	if err != nil {
		return err
	}

	// lookup old expiration
	store := ctx.KVStore(k.storeKey)

	if oldBytes := store.Get(key); oldBytes != nil {
		oldExpiration := &types.Expiration{}
		if err := k.cdc.Unmarshal(key, oldExpiration); err != nil {
			k.Logger(ctx).Error("could not unmarshal old expiration", "err", err, "expiration", expiration.String(), "oldExpirationBytes", oldBytes)
			return types.ErrExtendExpiration
		}
		// make sure that the new block height is higher than the old block height
		if expiration.BlockHeight <= oldExpiration.BlockHeight {
			k.Logger(ctx).Error("new block height must be higher than old block height", "err", err, "expiration", expiration.String(), "oldExpiration", oldExpiration.String())
			return types.ErrExtendExpiration
		}

		// validate owners are the same
		if expiration.Owner != oldExpiration.Owner {
			k.Logger(ctx).Error("new owner and old owner do not match", "err", err, "expiration", expiration.String(), "oldExpiration", oldExpiration.String())
			return types.ErrNewOwnerNoMatch
		}
		// todo: will we need to validate through authz?
	}

	// marshal expiration record and set
	b, err := k.cdc.Marshal(&expiration)
	if err != nil {
		return err
	}
	store.Set(key, b)

	// emit Extend event
	extendEvent := types.NewEventExpirationExtend(expiration.ModuleAssetId)
	return k.emitEvent(ctx, extendEvent)
}

// DeleteExpiration removes an expiration record from the kvstore.
func (k Keeper) DeleteExpiration(ctx sdk.Context, moduleAssetId string) error {
	key, err := types.GetModuleAssetKeyPrefix(moduleAssetId)
	if err != nil {
		return err
	}

	// delete record from store
	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		store.Delete(key)
	}

	// todo: are we going to delete the asset when the expiration is deleted?

	// emit Delete event
	deleteEvent := types.NewEventExpirationDelete(moduleAssetId)
	return k.emitEvent(ctx, deleteEvent)
}

// GetExpirationByModuleAssetId resolves a record by module asset id.
func (k Keeper) GetExpirationByModuleAssetId(ctx sdk.Context, moduleAssetId string) (*types.Expiration, error) {
	key, err := types.GetModuleAssetKeyPrefix(moduleAssetId)
	if err != nil {
		return nil, err
	}
	return getExpiration(ctx, k, key)
}

func getExpiration(ctx sdk.Context, keeper Keeper, key []byte) (*types.Expiration, error) {
	store := ctx.KVStore(keeper.storeKey)
	if !store.Has(key) {
		return nil, types.ErrExpirationNotFound
	}
	bz := store.Get(key)
	record := &types.Expiration{}
	err := keeper.cdc.Unmarshal(bz, record)
	return record, err
}

func (k Keeper) emitEvent(ctx sdk.Context, message proto.Message) error {
	if err := ctx.EventManager().EmitTypedEvent(message); err != nil {
		ctx.Logger().Error("unable to emit event", "error", err, "event", message)
		return err
	}
	return nil
}

func (k Keeper) isValidOwnerAddress(ctx sdk.Context, owner string) error {
	accAddress, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress
	}
	if owner := k.authKeeper.GetAccount(ctx, accAddress); owner == nil {
		return sdkerrors.ErrUnknownAddress
	}
	return nil
}

func (k Keeper) isValidModuleAsset(ctx sdk.Context, moduleAssetId string) error {
	// todo: we need to know the type of module asset we are dealing with so we can check with its Keeper
	panic("implement me")
}

//// GetExpirationByOwner resolves a record by owner.
//func (k Keeper) GetExpirationsByOwner(ctx sdk.Context, owner string) (record *types.Expiration, err error) {
//	key, err := types.GetOwnerKeyPrefix(owner)
//	if err != nil {
//		return nil, err
//	}
//	return getExpiration(ctx, k, key)
//}

//// GetExpirationsByOwner looks up all names bound to an address.
//func (k Keeper) getExpirationsByOwner(ctx sdk.Context, owner string) ([]types.Expiration, error) {
//	// Return value data structure.
//	records := make([]types.Expiration, 0)
//	// Handler that adds records if owner address matches.
//	recordsHandler := func(record types.Expiration) error {
//		if record.Owner == owner {
//			records = append(records, record)
//		}
//		return nil
//	}
//	// Calculate owner key prefix
//	key, err := types.GetOwnerKeyPrefix(owner)
//	if err != nil {
//		return nil, err
//	}
//	// Collect and return all names that match.
//	if err := k.IterateExpirations(ctx, key, recordsHandler); err != nil {
//		return records, err
//	}
//	return records, nil
//}

// IterateExpirations iterates over all the stored name records and passes them to a callback function.
func (k Keeper) IterateExpirations(ctx sdk.Context, prefix []byte, handle Handler) error {
	// Init a name record iterator
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	// Iterate over records, processing callbacks.
	for ; iterator.Valid(); iterator.Next() {
		record := types.Expiration{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if err := handle(record); err != nil {
			return err
		}
	}
	return nil
}
