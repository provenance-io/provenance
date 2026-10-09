package keeper

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// Handler is a name record handler function for use with IterateRecords.
type Handler func(record types.Attribute) error

// Keeper defines the attribute module Keeper
type Keeper struct {
	// Used to ensure accounts exist for addresses.
	authKeeper types.AccountKeeper
	// The keeper used for ensuring names resolve to owners.
	nameKeeper types.NameKeeper

	// storeService abstracts access to the module's KVStore.
	storeService store.KVStoreService
	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	modAddr sdk.AccAddress

	authority string
	// Collections schema.
	schema collections.Schema

	// attributes stores each Attribute.
	// Key prefix: 0x02  Layout: [len(addr)][addr][sha256(name)][sha256(value)]
	attributes collections.Map[types.AttrTriple, types.Attribute]

	// nameAddrCounts holds the per-(name, addr) attribute reference counter.
	// Key prefix: 0x03  Layout: [sha256(name)][len(addr)][addr]
	nameAddrCounts collections.Map[types.NameAddrPair, uint64]

	// expirationIndex is a sentinel index ordered by expiration epoch.
	// Key prefix: 0x04  Layout: [8-byte epoch][len(addr)][addr][sha256(name)][sha256(value)]
	expirationIndex collections.Map[types.ExpireTriple, bool]

	// params holds the module Params (MaxValueLength).
	// Key prefix: 0x05 — identical to old AttributeParamPrefix.
	params collections.Item[types.Params]
}

// NewKeeper returns an attribute keeper. It handles:
// - setting attributes against an account
// - removing attributes
// - scanning for existing attributes on an account
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService,
	authKeeper types.AccountKeeper, nameKeeper types.NameKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		authKeeper:      authKeeper,
		nameKeeper:      nameKeeper,
		cdc:             cdc,
		storeService:    storeService,
		modAddr:         authtypes.NewModuleAddress(types.ModuleName),
		authority:       authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		attributes:      collections.NewMap(sb, collections.NewPrefix(types.AttributeKeyPrefix), "attributes", types.AttrTripleKey, codec.CollValue[types.Attribute](cdc)),
		nameAddrCounts:  collections.NewMap(sb, types.AttributeAddrLookupKeyPrefix, "name_addr_counts", types.NameAddrPairKey, types.Uint64Value),
		expirationIndex: collections.NewMap(sb, types.AttributeExpirationKeyPrefix, "expiration_index", types.ExpireTripleKey, types.SentinelValue),
		params:          collections.NewItem(sb, types.AttributeParamPrefix, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(fmt.Errorf("attribute: failed to build collections schema: %w", err))
	}
	k.schema = schema

	nameKeeper.SetAttributeKeeper(k)
	return k
}

// GetAuthority is signer of the proposal
func (k Keeper) GetAuthority() string {
	return k.authority
}

// IsAuthority returns true if the provided address bech32 string is the authority address.
func (k Keeper) IsAuthority(addr string) bool {
	return strings.EqualFold(k.authority, addr)
}

// ValidateAuthority returns an error if the provided address is not the authority.
func (k Keeper) ValidateAuthority(addr string) error {
	if !k.IsAuthority(addr) {
		return govtypes.ErrInvalidSigner.Wrapf("expected %q got %q", k.GetAuthority(), addr)
	}
	return nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAllAttributes gets all attributes for an address.
func (k Keeper) GetAllAttributes(ctx sdk.Context, addr string) ([]types.Attribute, error) {
	defer telemetry.MeasureSince(telemetry.Now(), types.ModuleName, "keeper_method", "get_all")

	return k.attrsForAddr(ctx, types.GetAttributeAddressBytes(addr))
}

// GetAllAttributesAddr gets all attributes for an AccAddress or MetadataAddress.
func (k Keeper) GetAllAttributesAddr(ctx sdk.Context, addr []byte) ([]types.Attribute, error) {
	defer telemetry.MeasureSince(telemetry.Now(), types.ModuleName, "keeper_method", "get_all")

	return k.attrsForAddr(ctx, addr)
}

// GetAttributes gets all attributes with the given name from an account.
func (k Keeper) GetAttributes(ctx sdk.Context, addr string, name string) ([]types.Attribute, error) {
	defer telemetry.MeasureSince(telemetry.Now(), types.ModuleName, "keeper_method", "get")

	name = strings.ToLower(strings.TrimSpace(name))
	if _, err := k.nameKeeper.GetRecordByName(ctx, name); err != nil { // Ensure name exists (ie was bound to an address)
		return nil, err
	}
	return k.attrsForAddrName(ctx, types.GetAttributeAddressBytes(addr), name)
}

// FindMissingAttributes returns the subset of reqAttrs for which addr does not have a matching,
// unexpired attribute. Each of the provided reqAttrs can be either a full/exact name, or a name
// that start with a wildcard "*." that matches any attribute with a name with the same suffix.
// It's assumed that the provided reqAttrs have each been normalized.
func (k Keeper) FindMissingAttributes(ctx sdk.Context, addr []byte, reqAttrs []string) ([]string, error) {
	if len(reqAttrs) == 0 {
		return nil, nil
	}

	// Only look each required attribute up once, even if it's provided more than once.
	found := make(map[string]bool, len(reqAttrs))
	unique := make([]string, 0, len(reqAttrs))
	for _, reqAttr := range reqAttrs {
		if _, seen := found[reqAttr]; !seen {
			found[reqAttr] = false
			unique = append(unique, reqAttr)
		}
	}

	for _, reqAttr := range unique {
		// A wildcard is satisfied by anything under the name, an exact name only by itself.
		// Both are a prefix scan; the exact one still needs the name checked since, e.g., the
		// attributes named "aaa.bbb" and "aaa.bbbb" share the reversed prefix "bbb".
		name, isWildcard := strings.CutPrefix(reqAttr, "*.")
		revName := types.ReverseName(name)
		if isWildcard {
			revName += "."
		}

		has := false
		err := k.attributes.Walk(ctx, attrAddrNameRange(addr, revName)(nil),
			func(_ types.AttrTriple, attr types.Attribute) (bool, error) {
				if isExpired(ctx, attr) {
					return false, nil
				}
				if !isWildcard && attr.Name != reqAttr {
					return false, nil
				}
				has = true
				return true, nil // Stop at the first match.
			})
		if err != nil {
			return nil, fmt.Errorf("could not look up the %q attribute on %s: %w", reqAttr, sdk.AccAddress(addr), err)
		}
		found[reqAttr] = has
	}

	var missing []string
	for _, reqAttr := range reqAttrs {
		if !found[reqAttr] {
			missing = append(missing, reqAttr)
		}
	}
	return missing, nil
}

// IterateRecords iterates over all the stored attribute records and passes them to a callback function.
func (k Keeper) IterateRecords(ctx sdk.Context, handle Handler) error {
	// Init an attribute record iterator
	return k.attributes.Walk(ctx, nil, func(_ types.AttrTriple, record types.Attribute) (stop bool, err error) {
		if err = handle(record); err != nil {
			return true, err
		}
		return false, nil
	})
}

// SetAttribute stores an attribute under the given account. The attribute name must resolve to the given owner address.
func (k Keeper) SetAttribute(
	ctx sdk.Context, attr types.Attribute, owner sdk.AccAddress,
) error {
	defer telemetry.MeasureSince(telemetry.Now(), types.ModuleName, "keeper_method", "set")

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

	if ownerAcc := k.authKeeper.GetAccount(ctx, owner); ownerAcc == nil {
		return fmt.Errorf("no account found for owner address %q", owner.String())
	}

	if !k.nameKeeper.ResolvesTo(ctx, attr.Name, owner) {
		return fmt.Errorf("%q does not resolve to address %q", attr.Name, owner.String())
	}
	key := types.BuildAttrTriple(attr)
	oldAttr, exists, err := k.getAttr(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		// The attribute key doesn't include the expiration date, so changing only the expiration
		// reuses the key. Drop the old entry, or it's orphaned in the expiration index.
		if err = k.removeExpireEntry(ctx, oldAttr); err != nil {
			return err
		}
	}

	if err = k.attributes.Set(ctx, key, attr); err != nil {
		return err
	}
	if !exists {
		if err = k.incNameAddrCount(ctx, attr.Name, attr.GetAddressBytes()); err != nil {
			return err
		}
	}
	if err = k.addExpireEntry(ctx, attr); err != nil {
		return err
	}
	attributeAddEvent := types.NewEventAttributeAdd(attr, owner.String())
	return ctx.EventManager().EmitTypedEvent(attributeAddEvent)
}

// IncAttrNameAddressLookup increments the count of name to address lookups
func (k Keeper) IncAttrNameAddressLookup(ctx sdk.Context, name string, addrBytes []byte) {
	if err := k.incNameAddrCount(ctx, name, addrBytes); err != nil {
		k.Logger(ctx).Error("IncAttrNameAddressLookup failed", "error", err)
	}
}

func (k Keeper) incNameAddrCount(ctx sdk.Context, name string, addrBytes []byte) error {
	key := types.BuildNameAddrPair(name, addrBytes)
	current, err := k.nameAddrCounts.Get(ctx, key)
	switch {
	case errors.Is(err, collections.ErrNotFound):
		current = 0
	case err != nil:
		// A decode or store failure. Treating it as 0 would reset a real count.
		return err
	}
	return k.nameAddrCounts.Set(ctx, key, current+1)
}

// DecAttrNameAddressLookup decrements the name to account lookups and removes value if decremented to 0
func (k Keeper) DecAttrNameAddressLookup(ctx sdk.Context, name string, addrBytes []byte) {
	if err := k.decNameAddrCount(ctx, name, addrBytes); err != nil {
		k.Logger(ctx).Error("DecAttrNameAddressLookup failed", "error", err)
	}
}

func (k Keeper) decNameAddrCount(ctx sdk.Context, name string, addrBytes []byte) error {
	key := types.BuildNameAddrPair(name, addrBytes)
	current, err := k.nameAddrCounts.Get(ctx, key)
	switch {
	case errors.Is(err, collections.ErrNotFound):
		return nil // Nothing to decrement.
	case err != nil:
		// A decode or store failure. Removing the entry here would drop a real count.
		return err
	case current <= 1:
		return k.nameAddrCounts.Remove(ctx, key)
	default:
		return k.nameAddrCounts.Set(ctx, key, current-1)
	}
}

// UpdateAttribute updates an attribute under the given account. The attribute name must resolve to the given owner address and value must resolve to an existing attribute.
func (k Keeper) UpdateAttribute(ctx sdk.Context, originalAttribute types.Attribute, updateAttribute types.Attribute, owner sdk.AccAddress,
) error {
	defer telemetry.MeasureSince(telemetry.Now(), types.ModuleName, "keeper_method", "update")

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

	addrBz := originalAttribute.GetAddressBytes()
	origKey := types.BuildAttrTriple(originalAttribute)

	currentAttr, exists, err := k.getAttr(ctx, origKey)
	if err != nil {
		return err
	}

	var found bool
	if exists {
		if currentAttr.AttributeType == originalAttribute.AttributeType {
			found = true

			if err = k.attributes.Remove(ctx, origKey); err != nil {
				return err
			}
			k.DecAttrNameAddressLookup(ctx, currentAttr.Name, addrBz)
			if err = k.removeExpireEntry(ctx, currentAttr); err != nil {
				return err
			}
			// Preserve the existing expiration date if the update doesn't specify one.
			// MsgUpdateAttributeRequest has no expiration_date field, so updateAttribute.ExpirationDate
			// is nil when this arrives via the msg server.
			if updateAttribute.ExpirationDate == nil {
				updateAttribute.ExpirationDate = currentAttr.ExpirationDate
			}
			if err = k.attributes.Set(ctx, types.BuildAttrTriple(updateAttribute), updateAttribute); err != nil {
				return err
			}
			if err = k.incNameAddrCount(ctx, updateAttribute.Name, updateAttribute.GetAddressBytes()); err != nil {
				return err
			}
			if err = k.addExpireEntry(ctx, updateAttribute); err != nil {
				return err
			}

			attributeUpdateEvent := types.NewEventAttributeUpdate(originalAttribute, updateAttribute, owner.String())
			if err = ctx.EventManager().EmitTypedEvent(attributeUpdateEvent); err != nil {
				return err
			}
		}
	}
	if !found {
		return fmt.Errorf("no attributes updated with name %q : value %q : type: %s", originalAttribute.Name, string(originalAttribute.Value), originalAttribute.AttributeType.String())
	}
	return nil
}

// UpdateAttributeExpiration updates the expiration date on an attribute.
func (k Keeper) UpdateAttributeExpiration(ctx sdk.Context, updateAttribute types.Attribute, owner sdk.AccAddress,
) error {
	defer telemetry.MeasureSince(telemetry.Now(), types.ModuleName, "keeper_method", "update_expiration")

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

	origKey := types.BuildAttrTriple(updateAttribute)
	attr, getErr := k.attributes.Get(ctx, origKey)
	if getErr != nil {
		errorMessage := "no attributes updated"
		ctx.Logger().Error(errorMessage, "name", updateAttribute.Name, "value", string(updateAttribute.Value))
		return fmt.Errorf("%s with name %q : value %q : type: %s",
			errorMessage, updateAttribute.Name, string(updateAttribute.Value), updateAttribute.AttributeType.String())
	}

	if err := k.removeExpireEntry(ctx, attr); err != nil {
		return err
	}

	originalExpiration := attr.ExpirationDate
	attr.ExpirationDate = updateAttribute.ExpirationDate

	if err := k.attributes.Set(ctx, origKey, attr); err != nil {
		return err
	}
	if err := k.addExpireEntry(ctx, attr); err != nil {
		return err
	}

	attributeExpirationUpdateEvent := types.NewEventAttributeExpirationUpdate(attr, originalExpiration, owner.String())
	if err := ctx.EventManager().EmitTypedEvent(attributeExpirationUpdateEvent); err != nil {
		return err
	}

	return nil
}

// AccountsByAttribute returns a list of sdk.AccAddress that have attribute name assigned
func (k Keeper) AccountsByAttribute(ctx sdk.Context, name string) (addresses []sdk.AccAddress, err error) {
	revName := types.ReverseName(name)
	err = k.nameAddrCounts.Walk(ctx, napNameRange(revName)(nil), func(key types.NameAddrPair, _ uint64) (stop bool, walkErr error) {
		if key.RevName == revName {
			addrCopy := make([]byte, len(key.AddrBytes))
			copy(addrCopy, key.AddrBytes)
			addresses = append(addresses, addrCopy)
		}
		return false, nil
	})
	return
}

// DeleteAttribute removes attributes under the given account from the state store.
func (k Keeper) DeleteAttribute(ctx sdk.Context, addr string, name string, value *[]byte, owner sdk.AccAddress) error {
	defer telemetry.MeasureSince(telemetry.Now(), types.ModuleName, "keeper_method", "delete")

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

	addrz := types.GetAttributeAddressBytes(addr)
	rng := attrAddrNameRange(addrz, types.ReverseName(name))(nil)
	attrToDelete := make([]types.Attribute, 0)

	walkErr := k.attributes.Walk(ctx, rng, func(_ types.AttrTriple, attr types.Attribute) (stop bool, err error) {
		// NOTE: this name check is now load-bearing (see below), not just defensive.
		if attr.Address != addr || attr.Name != name {
			return false, nil
		}
		if deleteDistinct && !bytes.Equal(*value, attr.Value) {
			return false, nil
		}
		attrToDelete = append(attrToDelete, attr)
		return false, nil
	})
	if walkErr != nil {
		return walkErr
	}

	for _, attr := range attrToDelete {
		addrBz := attr.GetAddressBytes()
		if err := k.attributes.Remove(ctx, types.BuildAttrTriple(attr)); err != nil {
			return err
		}
		k.DecAttrNameAddressLookup(ctx, attr.Name, addrBz)
		if err := k.removeExpireEntry(ctx, attr); err != nil {
			return err
		}
		if !deleteDistinct {
			if err := ctx.EventManager().EmitTypedEvent(types.NewEventAttributeDelete(name, addr, owner.String())); err != nil {
				return err
			}
		} else {
			if err := ctx.EventManager().EmitTypedEvent(types.NewEventDistinctAttributeDelete(name, string(*value), addr, owner.String())); err != nil {
				return err
			}
		}
	}

	if len(attrToDelete) == 0 {
		if deleteDistinct {
			return fmt.Errorf("no keys deleted with name %q and value %q", name, string(*value))
		}
		return fmt.Errorf("no keys deleted with name %q", name)
	}

	return nil
}

// attrsForAddr gets all the attributes on the provided address.
func (k Keeper) attrsForAddr(ctx sdk.Context, addrBz []byte) (attrs []types.Attribute, err error) {
	err = k.attributes.Walk(ctx, attrAddrRange(addrBz)(nil), func(_ types.AttrTriple, attr types.Attribute) (bool, error) {
		if !isExpired(ctx, attr) {
			attrs = append(attrs, attr)
		}
		return false, nil
	})
	return attrs, err
}

// attrsForAddrName gets all the attributes with the provided name on the provided address.
func (k Keeper) attrsForAddrName(ctx sdk.Context, addrBz []byte, name string) (attrs []types.Attribute, err error) {
	rng := attrAddrNameRange(addrBz, types.ReverseName(name))(nil)
	err = k.attributes.Walk(ctx, rng, func(_ types.AttrTriple, attr types.Attribute) (bool, error) {
		// The range is on a name prefix, so make sure it is the name being asked for.
		if strings.EqualFold(attr.Name, name) && !isExpired(ctx, attr) {
			attrs = append(attrs, attr)
		}
		return false, nil
	})
	return attrs, err
}

// AttrsUnderName gets all the attributes on an address with a name under the provided name.
// E.g. a name of "kyc.pb" will return the attributes named "x.kyc.pb" and "y.x.kyc.pb", but not "kyc.pb".
// This is the wildcard ("*.kyc.pb") lookup, and is a single prefix scan.
func (k Keeper) AttrsUnderName(ctx sdk.Context, addr string, name string) (attrs []types.Attribute, err error) {
	addrBz := types.GetAttributeAddressBytes(addr)
	rng := attrAddrNameRange(addrBz, types.ReverseName(name)+".")(nil)
	err = k.attributes.Walk(ctx, rng, func(_ types.AttrTriple, attr types.Attribute) (bool, error) {
		attrs = append(attrs, attr)
		return false, nil
	})
	return attrs, err
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
	revName := types.ReverseName(name)

	// The attribute is needed alongside its key to clear the expiration entry and to build
	// the delete event, so collect both.
	type keyedAttr struct {
		key  types.AttrTriple
		attr types.Attribute
	}

	for _, acct := range accts {
		rng := attrAddrNameRange(acct, revName)(nil)
		var toRemove []keyedAttr
		if walkErr := k.attributes.Walk(ctx, rng, func(key types.AttrTriple, attr types.Attribute) (stop bool, err error) {
			// The range is a name prefix, so attr.Name == name is load-bearing.
			if bytes.Equal(key.AddrBytes, []byte(acct)) && attr.Name == name {
				toRemove = append(toRemove, keyedAttr{key: key, attr: attr})
			}
			return false, nil
		}); walkErr != nil {
			return walkErr
		}
		for _, entry := range toRemove {
			if err = k.attributes.Remove(ctx, entry.key); err != nil {
				return err
			}
			k.DecAttrNameAddressLookup(ctx, name, acct)
			if err = k.removeExpireEntry(ctx, entry.attr); err != nil {
				return err
			}
			deleteEvent := types.NewEventAttributeDelete(name, entry.attr.Address, owner.String())
			if err = ctx.EventManager().EmitTypedEvent(deleteEvent); err != nil {
				return err
			}
		}
	}
	return nil
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
	key := types.BuildAttrTriple(attr)

	has, hasErr := k.attributes.Has(ctx, key)
	if hasErr != nil {
		return hasErr
	}
	isNew := !has

	if err = k.attributes.Set(ctx, key, attr); err != nil {
		return err
	}
	if isNew {
		if err = k.incNameAddrCount(ctx, attr.Name, attr.GetAddressBytes()); err != nil {
			return err
		}
	}
	if err := k.addExpireEntry(ctx, attr); err != nil {
		return err
	}
	return nil
}

// DeleteExpiredAttributes find and delete expired attributes returns the total deleted
// limit sets the max amount to delete in a call, 0 for not limit
func (k Keeper) DeleteExpiredAttributes(ctx sdk.Context, limit int) int {
	// This sweep reads the expiration index through the raw store instead of
	// k.expirationIndex.Walk because a Walk aborts on the first key it cannot decode,
	// which would stop expired attributes from being deleted at all. Entries that can't
	// be decoded are deleted and skipped. This also runs in the begin blocker (outside a
	// tx), where there's no panic recovery, so it must not panic on malformed input.
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.AttributeExpirationKeyPrefix
	// Inclusive of the current block time: an attribute expiring exactly now is expired.
	end := storetypes.PrefixEndBytes(types.GetAttributeExpireTimePrefix(ctx.BlockTime()))

	type expiredEntry struct {
		rawKey    []byte
		expKey    types.ExpireTriple
		decodable bool
	}
	var entries []expiredEntry

	iter, iterErr := store.Iterator(prefix, end)
	if iterErr != nil {
		ctx.Logger().Error("attribute: unable to iterate expiration index", "error", iterErr)
		return 0
	}
	for ; iter.Valid(); iter.Next() {
		rawKey := bytes.Clone(iter.Key())
		entry := expiredEntry{rawKey: rawKey}
		if _, expKey, decErr := types.ExpireTripleKey.Decode(rawKey[len(prefix):]); decErr == nil {
			entry.expKey = expKey
			entry.decodable = true
		} else {
			ctx.Logger().Error(fmt.Sprintf("unable to decode expiration key: %v error: %v", rawKey, decErr))
		}
		entries = append(entries, entry)
		if limit != 0 && len(entries) >= limit {
			break
		}
	}
	if closeErr := iter.Close(); closeErr != nil {
		ctx.Logger().Error("attribute: unable to close expiration iterator", "error", closeErr)
	}

	for _, entry := range entries {
		if entry.decodable && !k.expireAttribute(ctx, entry.expKey) {
			continue
		}
		if delErr := store.Delete(entry.rawKey); delErr != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to remove expiration entry: %v error: %v", entry.rawKey, delErr))
		}
	}

	return len(entries)
}

// isExpired returns true if the provided attribute's expiration is before the current blocktime.
func isExpired(ctx sdk.Context, attr types.Attribute) bool {
	return attr.ExpirationDate != nil && !attr.ExpirationDate.After(ctx.BlockTime())
}

// ValidateExpirationDate returns error if attribute has an expiration date that is in the past of current block time
func (k Keeper) ValidateExpirationDate(ctx sdk.Context, attr types.Attribute) error {
	if isExpired(ctx, attr) {
		return fmt.Errorf("attribute expiration date %v is not after block time of %v", attr.ExpirationDate.UTC(), ctx.BlockTime().UTC())
	}
	return nil
}

// GetAccountData gets the value of the special accountdata attribute for the given address.
func (k Keeper) GetAccountData(ctx sdk.Context, addr string) (string, error) {
	attrs, err := k.GetAttributes(ctx, addr, types.AccountDataName)
	if err != nil {
		return "", fmt.Errorf("error finding %s for %q: %w", types.AccountDataName, addr, err)
	}
	// There should only ever be 0 or 1 entries. If there's more, just ignore the rest.
	if len(attrs) == 0 {
		return "", nil
	}
	return string(attrs[0].Value), nil
}

// SetAccountData sets/updates/deletes the value of the special accountdata attribute for a given address.
// An error is only returned if the account data cannot be set as requested.
func (k Keeper) SetAccountData(ctx sdk.Context, addr string, value string) error {
	existings, err := k.GetAttributes(ctx, addr, types.AccountDataName)
	if err != nil {
		return fmt.Errorf("could not look up existing %s for %q: %w", types.AccountDataName, addr, err)
	}
	if len(existings) > 0 {
		err = k.DeleteAttribute(ctx, addr, types.AccountDataName, nil, k.modAddr)
		if err != nil {
			return fmt.Errorf("could not delete existing %s for %q: %w", types.AccountDataName, addr, err)
		}
	}

	if len(value) > 0 {
		attr := types.Attribute{
			Name:          types.AccountDataName,
			Value:         []byte(value),
			AttributeType: types.AttributeType_String,
			Address:       addr,
		}
		err = k.SetAttribute(ctx, attr, k.modAddr)
		if err != nil {
			return fmt.Errorf("could not set %s for %q: %w", types.AccountDataName, addr, err)
		}
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventAccountDataUpdated{Account: addr})
}
func (k Keeper) addExpireEntry(ctx sdk.Context, attr types.Attribute) error {
	key, ok := types.BuildExpireTriple(attr)
	if !ok {
		return nil
	}
	return k.expirationIndex.Set(ctx, key, true)
}

func (k Keeper) removeExpireEntry(ctx sdk.Context, attr types.Attribute) error {
	key, ok := types.BuildExpireTriple(attr)
	if !ok {
		return nil
	}
	return k.expirationIndex.Remove(ctx, key)
}

// attrRange builds a Range over the attributes map. The start key is the one provided, or the
// beginning of the range when nil (which is how pagination resumes from a NextKey). It returns nil
// when the provided start key is outside the range, so a client-supplied pagination key can't
// widen a query to other accounts or names.
type attrRange func(start *types.AttrTriple) *collections.Range[types.AttrTriple]

// napRange is the attrRange equivalent for the nameAddrCounts map.
type napRange func(start *types.NameAddrPair) *collections.Range[types.NameAddrPair]

// nextBytes returns the smallest byte slice greater than every slice starting with the one provided.
// The bool is false when there isn't one, i.e. the provided bytes are all 0xFF.
func nextBytes(bz []byte) ([]byte, bool) {
	rv := bytes.Clone(bz)
	for i := len(rv) - 1; i >= 0; i-- {
		if rv[i] != 0xFF {
			rv[i]++
			return rv[:i+1], true
		}
	}
	return nil, false
}

// attrAddrEnd returns an AttrTriple that sorts after every attribute key for addrBz. Names are
// normalized ASCII and never contain 0xFF, so a RevName of "\xff" is past all of them. Bounding on
// the next address instead doesn't work: the key starts with the address length, so incrementing
// an address that ends in 0xFF yields a shorter key that sorts before the start of the range.
func attrAddrEnd(addrBz []byte) types.AttrTriple {
	return types.AttrTriple{AddrBytes: addrBz, RevName: "\xff"}
}

// attrAddrRange returns an attrRange covering all the attributes on addrBz.
func attrAddrRange(addrBz []byte) attrRange {
	return func(start *types.AttrTriple) *collections.Range[types.AttrTriple] {
		from := types.AttrTriple{AddrBytes: addrBz}
		if start != nil {
			if !bytes.Equal(start.AddrBytes, addrBz) {
				return nil
			}
			from = *start
		}
		return new(collections.Range[types.AttrTriple]).StartInclusive(from).EndExclusive(attrAddrEnd(addrBz))
	}
}

// attrAddrNameRange returns an attrRange covering the attributes on addrBz whose reversed name
// starts with revName. Since names are stored reversed, this is also how all the names under a
// name are found, e.g. a revName of "pb.kyc." covers everything under "kyc.pb".
func attrAddrNameRange(addrBz []byte, revName string) attrRange {
	return func(start *types.AttrTriple) *collections.Range[types.AttrTriple] {
		from := types.AttrTriple{AddrBytes: addrBz, RevName: revName}
		if start != nil {
			if !bytes.Equal(start.AddrBytes, addrBz) || !strings.HasPrefix(start.RevName, revName) {
				return nil
			}
			from = *start
		}
		rng := new(collections.Range[types.AttrTriple]).StartInclusive(from)
		if end, ok := nextBytes([]byte(revName)); ok {
			return rng.EndExclusive(types.AttrTriple{AddrBytes: addrBz, RevName: string(end)})
		}
		return rng.EndExclusive(attrAddrEnd(addrBz))
	}
}

// napNameRange returns a napRange covering the nameAddrCounts entries whose reversed name starts
// with revName.
func napNameRange(revName string) napRange {
	return func(start *types.NameAddrPair) *collections.Range[types.NameAddrPair] {
		from := types.NameAddrPair{RevName: revName}
		if start != nil {
			if !strings.HasPrefix(start.RevName, revName) {
				return nil
			}
			from = *start
		}
		rng := new(collections.Range[types.NameAddrPair]).StartInclusive(from)
		if end, ok := nextBytes([]byte(revName)); ok {
			rng = rng.EndExclusive(types.NameAddrPair{RevName: string(end)})
		}
		return rng
	}
}

// getAttr gets the attribute stored under key. The bool is false (with a nil error) when there
// isn't one; a non-nil error means the lookup itself failed.
func (k Keeper) getAttr(ctx sdk.Context, key types.AttrTriple) (types.Attribute, bool, error) {
	attr, err := k.attributes.Get(ctx, key)
	switch {
	case err == nil:
		return attr, true, nil
	case errors.Is(err, collections.ErrNotFound):
		return attr, false, nil
	default:
		return attr, false, err
	}
}

// expireAttribute deletes the attribute that an expiration entry points to, if it's actually
// expired. It returns false only when the attribute is expired but could not be removed, i.e.
// when the expiration entry should be kept so a later sweep can retry. Every other outcome (the
// attribute is already gone, can't be read, or isn't actually expired) returns true, since
// retrying can't change any of those.
func (k Keeper) expireAttribute(ctx sdk.Context, expKey types.ExpireTriple) bool {
	attrKey := types.AttrTriple{
		AddrBytes: expKey.AddrBytes,
		RevName:   expKey.RevName,
		ValueHash: expKey.ValueHash,
	}
	attr, exists, err := k.getAttr(ctx, attrKey)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("unable to read attribute to expire: %v error: %v", attrKey, err))
		return true
	}
	if !exists {
		return true // Already gone.
	}
	if !isExpired(ctx, attr) {
		// Expirations are indexed by the second, so an attribute that expires later in the current
		// second is in this sweep's range without being expired yet.
		current, hasExpiration := types.BuildExpireTriple(attr)
		return !hasExpiration || current.EpochSecs != expKey.EpochSecs
	}

	if err = k.attributes.Remove(ctx, attrKey); err != nil {
		ctx.Logger().Error(fmt.Sprintf("unable to remove expired attribute: %v error: %v", attrKey, err))
		return false
	}
	k.DecAttrNameAddressLookup(ctx, attr.Name, attr.GetAddressBytes())
	if err = ctx.EventManager().EmitTypedEvent(types.NewEventAttributeExpired(attr)); err != nil {
		ctx.Logger().Error(fmt.Sprintf("failed to emit attribute expired typed event %v", err))
	}
	return true
}
