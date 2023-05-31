package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// Handler is a name record handler function for use with IterateRecords.
type Handler func(record types.Attribute) error

// Keeper defines the attribute module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set attribute specific params
	paramSpace paramtypes.Subspace

	// Used to ensure accounts exist for addresses.
	authKeeper types.AccountKeeper
	// The keeper used for ensuring names resolve to owners.
	nameKeeper types.NameKeeper

	// Key to access the key-value store from sdk.Context.
	storeKey storetypes.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc codec.BinaryCodec
}

// NewKeeper returns an attribute keeper. It handles:
// - setting attributes against an account
// - removing attributes
// - scanning for existing attributes on an account
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace,
	authKeeper types.AccountKeeper, nameKeeper types.NameKeeper,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	keeper := Keeper{
		storeKey:   key,
		paramSpace: paramSpace,
		authKeeper: authKeeper,
		nameKeeper: nameKeeper,
		cdc:        cdc,
	}
	nameKeeper.SetAttributeKeeper(keeper)
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAllAttributes gets all attributes for an address.
func (k Keeper) GetAllAttributes(ctx sdk.Context, addr string) ([]types.Attribute, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "get_all")

	pred := func(s string) bool { return true }
	return k.prefixScan(ctx, types.AddrStrAttributesKeyPrefix(addr), pred)
}

// GetAllAttributesAddr gets all attributes for an AccAddress or MetadataAddress.
func (k Keeper) GetAllAttributesAddr(ctx sdk.Context, addr []byte) ([]types.Attribute, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "get_all")

	pred := func(s string) bool { return true }
	return k.prefixScan(ctx, types.AddrAttributesKeyPrefix(addr), pred)
}

// GetAttributes gets all attributes with the given name from an account.
func (k Keeper) GetAttributes(ctx sdk.Context, addr string, name string) ([]types.Attribute, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "get")

	name = strings.ToLower(strings.TrimSpace(name))
	if _, err := k.nameKeeper.GetRecordByName(ctx, name); err != nil { // Ensure name exists (ie was bound to an address)
		return nil, err
	}
	pred := func(s string) bool { return strings.EqualFold(s, name) }
	return k.prefixScan(ctx, types.AddrStrAttributesNameKeyPrefix(addr, name), pred)
}

// IterateRecords iterates over all the stored attribute records and passes them to a callback function.
func (k Keeper) IterateRecords(ctx sdk.Context, prefix []byte, handle Handler) error {
	// Init a attribute record iterator
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	// Iterate over records, processing callbacks.
	for ; iterator.Valid(); iterator.Next() {
		record := types.Attribute{}
		// get proto objects for legacy prefix with legacy amino codec.
		if bytes.Equal(types.AttributeKeyPrefixAmino, prefix) {
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

// Stores an attribute under the given account. The attribute name must resolve to the given owner address.
func (k Keeper) SetAttribute(
	ctx sdk.Context, attr types.Attribute, owner sdk.AccAddress,
) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "set")

	if err := k.ValidateExpirationDate(ctx, attr); err != nil {
		return err
	}

	// Ensure attribute is valid
	if err := attr.ValidateBasic(); err != nil {
		return err
	}

	// Ensure attribute value length does not exceed max length value
	maxLength := k.GetMaxValueLength(ctx)
	if int(maxLength) < len(attr.Value) {
		return fmt.Errorf("attribute value length of %v exceeds max length %v", len(attr.Value), maxLength)
	}

	normalizedName, err := k.nameKeeper.Normalize(ctx, attr.Name)
	if err != nil {
		return fmt.Errorf("unable to normalize attribute name %q: %w", attr.Name, err)
	}
	attr.Name = normalizedName
	// Verify an account exists for the given owner address
	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address %q", owner.String())
	}
	// Verify name resolves to owner
	if !k.nameKeeper.ResolvesTo(ctx, attr.Name, owner) {
		return fmt.Errorf("%q does not resolve to address %q", attr.Name, owner.String())
	}
	// Store the sanitized account attribute
	bz, err := k.cdc.Marshal(&attr)
	if err != nil {
		return err
	}

	key := types.AddrAttributeKey(attr.GetAddressBytes(), attr)

	store := ctx.KVStore(k.storeKey)
	store.Set(key, bz)
	k.IncAttrNameAddressLookup(ctx, attr.Name, attr.GetAddressBytes())
	k.addAttributeExpireLookup(store, attr)

	attributeAddEvent := types.NewEventAttributeAdd(attr, owner.String())

	return ctx.EventManager().EmitTypedEvent(attributeAddEvent)
}

// IncAttrNameAddressLookup increments the count of name to address lookups
func (k Keeper) IncAttrNameAddressLookup(ctx sdk.Context, name string, addrBytes []byte) {
	store := ctx.KVStore(k.storeKey)
	key := types.AttributeNameAddrKeyPrefix(name, addrBytes)
	bz := store.Get(key)
	id := uint64(0)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	bz = sdk.Uint64ToBigEndian(id + 1)
	store.Set(key, bz)
}

// DecAttrNameAddressLookup decrements the name to account lookups and removes value if decremented to 0
func (k Keeper) DecAttrNameAddressLookup(ctx sdk.Context, name string, addrBytes []byte) {
	store := ctx.KVStore(k.storeKey)
	key := types.AttributeNameAddrKeyPrefix(name, addrBytes)
	bz := store.Get(key)
	if bz != nil {
		value := binary.BigEndian.Uint64(bz)
		if value <= uint64(1) {
			store.Delete(key)
		} else {
			store.Set(key, sdk.Uint64ToBigEndian(value-1))
		}
	}
}

// UpdateAttribute updates an attribute under the given account. The attribute name must resolve to the given owner address and value must resolve to an existing attribute.
func (k Keeper) UpdateAttribute(ctx sdk.Context, originalAttribute types.Attribute, updateAttribute types.Attribute, owner sdk.AccAddress,
) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "update")

	var err error

	if err = originalAttribute.ValidateBasic(); err != nil {
		return err
	}

	if err = updateAttribute.ValidateBasic(); err != nil {
		return err
	}
	maxLength := k.GetMaxValueLength(ctx)
	if int(maxLength) < len(updateAttribute.Value) {
		return fmt.Errorf("update attribute value length of %v exceeds max length %v", len(updateAttribute.Value), maxLength)
	}

	normalizedName, err := k.nameKeeper.Normalize(ctx, updateAttribute.Name)
	if err != nil {
		return fmt.Errorf("unable to normalize attribute name %q: %w", updateAttribute.Name, err)
	}

	normalizedOrigName, err := k.nameKeeper.Normalize(ctx, originalAttribute.Name)
	if err != nil {
		return fmt.Errorf("unable to normalize attribute name %q: %w", originalAttribute.Name, err)
	}

	if normalizedName != normalizedOrigName {
		return fmt.Errorf("update and original names must match %s : %s", normalizedName, normalizedOrigName)
	}

	updateAttribute.Name = normalizedName

	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address %q", owner.String())
	}

	if !k.nameKeeper.ResolvesTo(ctx, updateAttribute.Name, owner) {
		return fmt.Errorf("%q does not resolve to address %q", updateAttribute.Name, owner.String())
	}

	store := ctx.KVStore(k.storeKey)
	addrBz := originalAttribute.GetAddressBytes()
	attrKey := types.AddrAttributeKey(addrBz, originalAttribute)
	currentAttr := store.Get(attrKey)

	var found bool
	if currentAttr != nil {
		attr := types.Attribute{}
		if err := k.cdc.Unmarshal(currentAttr, &attr); err != nil {
			return err
		}

		if attr.AttributeType == originalAttribute.AttributeType {
			found = true

			store.Delete(attrKey)
			k.DecAttrNameAddressLookup(ctx, attr.Name, addrBz)
			k.deleteAttributeExpireLookup(store, attr)

			bz, err := k.cdc.Marshal(&updateAttribute)
			if err != nil {
				return err
			}
			updatedKey := types.AddrAttributeKey(addrBz, updateAttribute)
			store.Set(updatedKey, bz)
			k.IncAttrNameAddressLookup(ctx, updateAttribute.Name, updateAttribute.GetAddressBytes())
			k.addAttributeExpireLookup(store, updateAttribute)

			attributeUpdateEvent := types.NewEventAttributeUpdate(originalAttribute, updateAttribute, owner.String())
			if err := ctx.EventManager().EmitTypedEvent(attributeUpdateEvent); err != nil {
				return err
			}
		}
	}
	if !found {
		errorMessage := "no attributes updated"
		ctx.Logger().Error(errorMessage, "name", originalAttribute.Name, "value", string(originalAttribute.Value))
		return fmt.Errorf("%s with name %q : value %q : type: %s", errorMessage, originalAttribute.Name, string(originalAttribute.Value), originalAttribute.AttributeType.String())
	}
	return nil
}

// UpdateAttributeExpiration updates the expiration date on an attribute.
func (k Keeper) UpdateAttributeExpiration(ctx sdk.Context, updateAttribute types.Attribute, owner sdk.AccAddress,
) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "update_expiration")

	if err := k.ValidateExpirationDate(ctx, updateAttribute); err != nil {
		return err
	}

	var err error
	normalizedOrigName, err := k.nameKeeper.Normalize(ctx, updateAttribute.Name)
	if err != nil {
		return fmt.Errorf("unable to normalize attribute name %q: %w", updateAttribute.Name, err)
	}
	updateAttribute.Name = normalizedOrigName

	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address %q", owner.String())
	}

	if !k.nameKeeper.ResolvesTo(ctx, updateAttribute.Name, owner) {
		return fmt.Errorf("%q does not resolve to address %q", updateAttribute.Name, owner.String())
	}

	store := ctx.KVStore(k.storeKey)
	attrKey := types.AddrAttributeKey(updateAttribute.GetAddressBytes(), updateAttribute)
	currentAttr := store.Get(attrKey)
	if currentAttr != nil {
		attr := types.Attribute{}
		if err := k.cdc.Unmarshal(currentAttr, &attr); err != nil {
			return err
		}

		k.deleteAttributeExpireLookup(store, attr)

		originalExpiration := attr.ExpirationDate
		attr.ExpirationDate = updateAttribute.ExpirationDate
		bz, err := k.cdc.Marshal(&attr)
		if err != nil {
			return err
		}
		store.Set(attrKey, bz)

		k.addAttributeExpireLookup(store, attr)

		attributeExpirationUpdateEvent := types.NewEventAttributeExpirationUpdate(attr, originalExpiration, owner.String())
		if err := ctx.EventManager().EmitTypedEvent(attributeExpirationUpdateEvent); err != nil {
			return err
		}
	} else {
		errorMessage := "no attributes updated"
		ctx.Logger().Error(errorMessage, "name", updateAttribute.Name, "value", string(updateAttribute.Value))
		return fmt.Errorf("%s with name %q : value %q : type: %s", errorMessage, updateAttribute.Name, string(updateAttribute.Value), updateAttribute.AttributeType.String())
	}

	return nil
}

// AccountsByAttribute returns a list of sdk.AccAddress that have attribute name assigned
func (k Keeper) AccountsByAttribute(ctx sdk.Context, name string) (addresses []sdk.AccAddress, err error) {
	store := ctx.KVStore(k.storeKey)
	keyPrefix := types.AttributeNameKeyPrefix(name)
	it := sdk.KVStorePrefixIterator(store, keyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		addressBytes, err := types.GetAddressFromKey(it.Key())
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addressBytes)
	}
	return
}

// DeleteAttribute removes attributes under the given account from the state store.
func (k Keeper) DeleteAttribute(ctx sdk.Context, addr string, name string, value *[]byte, owner sdk.AccAddress) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "delete")

	var deleteDistinct bool
	if value != nil {
		deleteDistinct = true
	}

	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address %q", owner.String())
	}

	if !k.nameKeeper.ResolvesTo(ctx, name, owner) {
		if k.nameKeeper.NameExists(ctx, name) {
			return fmt.Errorf("%q does not resolve to address %q", name, owner.String())
		}
		// else name does not exist (anymore) so we can't enforce permission check on delete here, proceed.
	}

	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.AddrStrAttributesNameKeyPrefix(addr, name))
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	attrToDelete := []types.Attribute{} // do delete logic outside of iterator
	for ; iter.Valid(); iter.Next() {
		attr := types.Attribute{}
		if err := k.cdc.Unmarshal(iter.Value(), &attr); err != nil {
			return err
		}

		if attr.Name == name && (!deleteDistinct || bytes.Equal(*value, attr.Value)) {
			attrToDelete = append(attrToDelete, attr)
		}
	}
	iter.Close()
	iter = nil

	for _, attr := range attrToDelete {
		addrBz := attr.GetAddressBytes()
		store.Delete(types.AddrAttributeKey(addrBz, attr))
		k.DecAttrNameAddressLookup(ctx, attr.Name, addrBz)
		k.deleteAttributeExpireLookup(store, attr)
		if !deleteDistinct {
			deleteEvent := types.NewEventAttributeDelete(name, addr, owner.String())
			if err := ctx.EventManager().EmitTypedEvent(deleteEvent); err != nil {
				return err
			}
		} else {
			deleteEvent := types.NewEventDistinctAttributeDelete(name, string(*value), addr, owner.String())
			if err := ctx.EventManager().EmitTypedEvent(deleteEvent); err != nil {
				return err
			}
		}
	}

	errm := "no keys deleted"
	if len(attrToDelete) == 0 && deleteDistinct {
		ctx.Logger().Error(errm, "name", name, "value")
		return fmt.Errorf("%s with name %s value %q", errm, name, string(*value))
	} else if len(attrToDelete) == 0 && !deleteDistinct {
		ctx.Logger().Error(errm, "name", name)
		return fmt.Errorf("%s with name %s", errm, name)
	}
	return nil
}

// PurgeAttribute removes attributes under the given account from the state store.
func (k Keeper) PurgeAttribute(ctx sdk.Context, name string, owner sdk.AccAddress) error {
	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address %q", owner.String())
	}

	if !k.nameKeeper.ResolvesTo(ctx, name, owner) {
		if k.nameKeeper.NameExists(ctx, name) {
			return fmt.Errorf("%q does not resolve to address %q", name, owner.String())
		}
		// else name does not exist (anymore) so we can't enforce permission check on delete here, proceed.
	}

	accts, err := k.AccountsByAttribute(ctx, name)
	if err != nil {
		return err
	}
	for _, acct := range accts {
		attrToDelete := [][]byte{}
		store := ctx.KVStore(k.storeKey)
		it := sdk.KVStorePrefixIterator(store, types.AddrAttributesNameKeyPrefix(acct, name))
		for ; it.Valid(); it.Next() {
			attrToDelete = append(attrToDelete, it.Key())
		}
		it.Close()
		for _, key := range attrToDelete {
			store.Delete(key)
			k.DecAttrNameAddressLookup(ctx, name, acct)
		}
		it.Close()
	}
	return nil
}

// A predicate function for matching names
type namePred = func(string) bool

// Scan all attributes that match the given prefix.
func (k Keeper) prefixScan(ctx sdk.Context, prefix []byte, f namePred) (attrs []types.Attribute, err error) {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		attr := types.Attribute{}
		if err = k.cdc.Unmarshal(it.Value(), &attr); err != nil {
			return
		}
		if f(attr.Name) {
			attrs = append(attrs, attr)
		}
	}
	return
}

// A genesis helper that imports attribute state without owner checks.
func (k Keeper) importAttribute(ctx sdk.Context, attr types.Attribute) error {
	if err := k.ValidateExpirationDate(ctx, attr); err != nil {
		// don't return error, this will ensure this attribute is skipped since it is expired
		return nil
	}

	// Ensure attribute is valid
	err := attr.ValidateBasic()
	if err != nil {
		return err
	}
	// Ensure name is stored in normalized format.
	attrNameOrig := attr.Name
	if attr.Name, err = k.nameKeeper.Normalize(ctx, attr.Name); err != nil {
		return fmt.Errorf("unable to normalize attribute name %q: %w", attrNameOrig, err)
	}
	// Store the sanitized account attribute
	bz, err := k.cdc.Marshal(&attr)
	if err != nil {
		return err
	}
	key := types.AddrAttributeKey(attr.GetAddressBytes(), attr)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, bz)
	k.IncAttrNameAddressLookup(ctx, attr.Name, attr.GetAddressBytes())
	k.addAttributeExpireLookup(store, attr)
	return nil
}

// PopulateAddressAttributeNameTable retrieves all attributes and populates address by attribute name lookup table
// TODO: remove after v1.15.0 upgrade handler is removed
func (k Keeper) PopulateAddressAttributeNameTable(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.AttributeKeyPrefix)
	for ; it.Valid(); it.Next() {
		attr := types.Attribute{}
		if err := k.cdc.Unmarshal(it.Value(), &attr); err != nil {
			return
		}
		k.IncAttrNameAddressLookup(ctx, attr.Name, attr.GetAddressBytes())
	}
}

// DeleteExpiredAttributes find and delete expired attributes returns the total deleted
// limit sets the max amount to delete in a call, 0 for not limit
func (k Keeper) DeleteExpiredAttributes(ctx sdk.Context, limit int) int {
	expirationKeys := [][]byte{}
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(types.AttributeExpirationKeyPrefix, types.GetAttributeExpireTimePrefix(ctx.BlockTime()))
	for ; iterator.Valid(); iterator.Next() {
		expirationKeys = append(expirationKeys, iterator.Key())
	}
	iterator.Close()

	count := 0
	for _, expirationKey := range expirationKeys {
		attrKey := types.GetAddrAttributeKeyFromExpireKey(expirationKey)
		bz := store.Get(attrKey)
		if bz != nil {
			var attribute types.Attribute
			if err := k.cdc.Unmarshal(bz, &attribute); err == nil {
				// delete attribute from store
				store.Delete(attrKey)
				// dec name to address lookup table count
				k.DecAttrNameAddressLookup(ctx, attribute.Name, attribute.GetAddressBytes())

				deleteExpirationEvent := types.NewEventAttributeExpired(attribute)
				if err = ctx.EventManager().EmitTypedEvent(deleteExpirationEvent); err != nil {
					ctx.Logger().Error(fmt.Sprintf("failed to emit typed event %v", err))
				}
				count++
			} else {
				ctx.Logger().Error(fmt.Sprintf("unable to unmarshal attribute to delete key: %v error: %v", attrKey, err))
			}
		}

		// delete the expiration lookup key
		store.Delete(expirationKey)
		if limit != 0 && count >= limit {
			break
		}
	}
	return count
}

// addAttributeExpireLookup safely adds attribute expire key to store, if expire date exists, else no-op
func (k Keeper) addAttributeExpireLookup(store sdk.KVStore, attr types.Attribute) {
	expireKey := types.AttributeExpireKey(attr)
	if expireKey != nil {
		store.Set(expireKey, []byte{})
	}
}

// deleteAttributeExpireLookup safely removes attribute expire key from store if expire date exists, else no-op
func (k Keeper) deleteAttributeExpireLookup(store sdk.KVStore, attr types.Attribute) {
	expireKey := types.AttributeExpireKey(attr)
	if expireKey != nil {
		store.Delete(expireKey)
	}
}

// ValidateExpirationDate returns error if attribute has an expiration date that is in the past of current block time
func (k Keeper) ValidateExpirationDate(ctx sdk.Context, attr types.Attribute) error {
	if attr.ExpirationDate != nil && attr.ExpirationDate.Unix() < ctx.BlockTime().Unix() {
		return fmt.Errorf("attribute expiration date %v is before block time of %v", attr.ExpirationDate.UTC(), ctx.BlockTime().UTC())
	}
	return nil
}
