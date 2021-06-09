package keeper

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/provenance-io/provenance/x/attribute/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc codec.BinaryMarshaler
}

// NewKeeper returns an attribute keeper. It handles:
// - setting attributes against an account
// - removing attributes
// - scanning for existing attributes on an account
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	authKeeper types.AccountKeeper, nameKeeper types.NameKeeper,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   key,
		paramSpace: paramSpace,
		authKeeper: authKeeper,
		nameKeeper: nameKeeper,
		cdc:        cdc,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAllAttributes gets all attributes for account.
func (k Keeper) GetAllAttributes(ctx sdk.Context, acc sdk.AccAddress) ([]types.Attribute, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "get_all")

	pred := func(s string) bool { return true }
	return k.prefixScan(ctx, types.AccountAttributesKeyPrefix(acc), pred)
}

// GetAttributes gets all attributes with the given name from an account.
func (k Keeper) GetAttributes(ctx sdk.Context, acc sdk.AccAddress, name string) ([]types.Attribute, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "get")

	name = strings.ToLower(strings.TrimSpace(name))
	if _, err := k.nameKeeper.GetRecordByName(ctx, name); err != nil { // Ensure name exists (ie was bound to an address)
		return nil, err
	}
	pred := func(s string) bool { return strings.EqualFold(s, name) }
	return k.prefixScan(ctx, types.AccountAttributesNameKeyPrefix(acc, name), pred)
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
			if err := types.ModuleCdc.UnmarshalBinaryBare(iterator.Value(), &record); err != nil {
				return err
			}
		} else {
			if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &record); err != nil {
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
		return fmt.Errorf("unable to normalize attribute name \"%s\": %w", attr.Name, err)
	}
	attr.Name = normalizedName
	// Verify an account exists for the given owner address
	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address \"%s\"", owner.String())
	}
	// Verify name resolves to owner
	if !k.nameKeeper.ResolvesTo(ctx, attr.Name, owner) {
		return fmt.Errorf("\"%s\" does not resolve to address \"%s\"", attr.Name, owner.String())
	}
	// Store the sanitized account attribute
	bz, err := k.cdc.MarshalBinaryBare(&attr)
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(attr.Address)
	if err != nil {
		return err
	}
	key := types.AccountAttributeKey(addr, attr)

	store := ctx.KVStore(k.storeKey)
	store.Set(key, bz)

	attributeAddEvent := types.NewEventAttributeAdd(attr, owner.String())
	if err := ctx.EventManager().EmitTypedEvent(attributeAddEvent); err != nil {
		return err
	}

	return nil
}

// Updates an attribute under the given account. The attribute name must resolve to the given owner address and value must resolve to an existing attribute.
func (k Keeper) UpdateAttribute(ctx sdk.Context, originalAttribute types.Attribute, updateAttribute types.Attribute, owner sdk.AccAddress,
) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "update")

	accountAddress, err := sdk.AccAddressFromBech32(originalAttribute.Address)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("unable to normalize attribute name \"%s\": %w", updateAttribute.Name, err)
	}

	updateAttribute.Name = normalizedName

	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address \"%s\"", owner.String())
	}

	if !k.nameKeeper.ResolvesTo(ctx, updateAttribute.Name, owner) {
		return fmt.Errorf("\"%s\" does not resolve to address \"%s\"", updateAttribute.Name, owner.String())
	}

	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.AccountAttributesNameKeyPrefix(accountAddress, originalAttribute.Name))
	var found bool
	for ; it.Valid(); it.Next() {
		attr := types.Attribute{}
		if err := k.cdc.UnmarshalBinaryBare(it.Value(), &attr); err != nil {
			return err
		}

		if attr.Name == updateAttribute.Name && bytes.Equal(attr.Value, originalAttribute.Value) {
			found = true
			store.Delete(it.Key())

			bz, err := k.cdc.MarshalBinaryBare(&attr)
			if err != nil {
				return err
			}
			updatedKey := types.AccountAttributeKey(accountAddress, updateAttribute)
			store.Set(updatedKey, bz)

			attributeUpdateEvent := types.NewEventAttributeUpdate(originalAttribute, updateAttribute, owner.String())
			if err := ctx.EventManager().EmitTypedEvent(attributeUpdateEvent); err != nil {
				return err
			}
			break
		}
	}
	if !found {
		errorMessage := "no attributes updated"
		ctx.Logger().Error(errorMessage, "name", originalAttribute.Name, "value", string(originalAttribute.Value))
		return fmt.Errorf("%s with name %s and value %s", errorMessage, originalAttribute.Name, string(originalAttribute.Value))
	}
	return nil
}

// Removes attributes under the given account. The attribute name and value if given must resolve to the given owner address.
func (k Keeper) DeleteAttribute(ctx sdk.Context, acc sdk.AccAddress, name string, value *[]byte, owner sdk.AccAddress) error {
	var deleteWithValue bool
	if value != nil {
		deleteWithValue = true
	}

	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "keeper_method", "delete")

	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address \"%s\"", owner.String())
	}

	if !k.nameKeeper.ResolvesTo(ctx, name, owner) {
		if k.nameKeeper.NameExists(ctx, name) {
			return fmt.Errorf("\"%s\" does not resolve to address \"%s\"", name, owner.String())
		}
		// else name does not exist (anymore) so we can't enforce permission check on delete here, proceed.
	}

	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.AccountAttributesNameKeyPrefix(acc, name))
	var count int
	for ; it.Valid(); it.Next() {
		attr := types.Attribute{}
		if err := k.cdc.UnmarshalBinaryBare(it.Value(), &attr); err != nil {
			return err
		}

		if attr.Name == name && (!deleteWithValue || bytes.Equal(*value, attr.Value)) {
			count++
			store.Delete(it.Key())

			if !deleteWithValue {
				deleteEvent := types.NewEventAttributeDelete(name, acc.String(), owner.String())
				if err := ctx.EventManager().EmitTypedEvent(deleteEvent); err != nil {
					return err
				}
			} else {
				deleteEvent := types.NewEventAttributeDeleteWithValue(name, string(*value), acc.String(), owner.String())
				if err := ctx.EventManager().EmitTypedEvent(deleteEvent); err != nil {
					return err
				}
			}
		}
	}
	if count == 0 {
		errm := "no keys deleted"
		ctx.Logger().Error(errm, "name", name)
		return fmt.Errorf("%s with name %s", errm, name)
	}
	return nil
}

// A predicate function for matching names
type namePred = func(string) bool

// Scan all attributes that match the given prefix.
func (k Keeper) prefixScan(ctx sdk.Context, prefix []byte, f namePred) (attrs []types.Attribute, err error) {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, prefix)
	for ; it.Valid(); it.Next() {
		attr := types.Attribute{}
		if err = k.cdc.UnmarshalBinaryBare(it.Value(), &attr); err != nil {
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
	// Ensure attribute is valid
	if err := attr.ValidateBasic(); err != nil {
		return err
	}
	// Attribute must have a valid, non-empty address to import
	if strings.TrimSpace(attr.Address) == "" {
		return fmt.Errorf("unable to import attribute with empty address")
	}
	acc, err := sdk.AccAddressFromBech32(attr.Address)
	if err != nil {
		return err
	}
	// Ensure name is stored in normalized format.
	if attr.Name, err = k.nameKeeper.Normalize(ctx, attr.Name); err != nil {
		return fmt.Errorf("unable to normalize attribute name \"%s\": %w", attr.Name, err)
	}
	// Store the sanitized account attribute
	bz, err := k.cdc.MarshalBinaryBare(&attr)
	if err != nil {
		return err
	}
	key := types.AccountAttributeKey(acc, attr)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, bz)
	return nil
}
