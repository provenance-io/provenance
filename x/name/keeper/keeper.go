package keeper

import (
	"bytes"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// Keeper defines the name module Keeper
type Keeper struct {
	// Key to access the key-value store from sdk.Context.
	storeKey storetypes.StoreKey
	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// the signing authority for the gov proposals
	authority string

	// Attribute keeper
	attrKeeper types.AttributeKeeper

	// Schema definition
	Schema collections.Schema

	addrIndex   collections.Map[[]byte, types.NameRecord] // key: 0x05 + addr + name key
	nameRecords collections.Map[[]byte, types.NameRecord] // key: 0x03 + hash(name)
	paramsStore collections.Item[types.Params]
}

// NewKeeper returns a name keeper. It handles:
// - managing a hierarchy of names
// - enforcing permissions for name creation/deletion
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	storeService store.KVStoreService,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:         cdc,
		storeKey:    key,
		authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		nameRecords: collections.NewMap(sb, collections.NewPrefix(types.NameKeyPrefix), "names", types.RawBytesKey, codec.CollValue[types.NameRecord](cdc)),
		addrIndex:   collections.NewMap(sb, collections.NewPrefix(types.AddressKeyPrefix), "addr_index", types.RawBytesKey, codec.CollValue[types.NameRecord](cdc)),
		paramsStore: collections.NewItem(sb, types.NameParamStoreKey, "params", codec.CollValue[types.Params](cdc)),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(fmt.Sprintf("name module schema build failed: %v", err))
	}
	k.Schema = schema
	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
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

// SetAttributeKeeper sets the attribute keeper
func (k *Keeper) SetAttributeKeeper(ak types.AttributeKeeper) {
	if k.attrKeeper != nil && ak != nil && k.attrKeeper != ak {
		panic("the attribute keeper has already been set")
	}
	k.attrKeeper = ak
}

// ResolvesTo to determines whether a name resolves to a given address.
func (k Keeper) ResolvesTo(ctx sdk.Context, name string, addr sdk.AccAddress) bool {
	stored, err := k.GetRecordByName(ctx, name)
	if err != nil {
		return false
	}
	return addr.String() == stored.Address
}

// SetNameRecord binds a name to an address.
func (k Keeper) SetNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	var err error
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return err
	}
	if err = types.ValidateAddress(addr); err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	if err = k.addRecord(ctx, name, addr, restrict, false); err != nil {
		return err
	}

	nameBoundEvent := types.NewEventNameBound(addr.String(), normalizedName, restrict)

	return ctx.EventManager().EmitTypedEvent(nameBoundEvent)
}

// UpdateNameRecord updates the owner address and restricted flag on a name.
func (k Keeper) UpdateNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	var err error
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return err
	}
	if err := types.ValidateAddress(addr); err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	nameKey, err := types.GetNameKeyPrefix(normalizedName)
	if err != nil {
		return types.ErrNameNotBound.Wrapf("failed to get name key prefix for name %q: %v", normalizedName, err)
	}

	// Get existing record
	existing, err := k.nameRecords.Get(ctx, nameKey)
	if err != nil {
		return err
	}
	// If address is changing, remove old address index
	if existing.Address != addr.String() {
		oldAddr, err := sdk.AccAddressFromBech32(existing.Address)
		if err != nil {
			return err
		}
		oldAddrPrefix, err := types.GetAddressKeyPrefix(oldAddr)
		if err != nil {
			return err
		}
		oldAddrIndexKey := append(oldAddrPrefix, nameKey...)
		if err := k.addrIndex.Remove(ctx, oldAddrIndexKey); err != nil {
			return err
		}
	}
	// Create new record
	record := types.NewNameRecord(normalizedName, addr, restrict)
	if err := record.Validate(); err != nil {
		return err
	}

	// Update name record
	if err := k.nameRecords.Set(ctx, nameKey, record); err != nil {
		return err
	}

	// Update address index
	addrPrefix, err := types.GetAddressKeyPrefix(addr)
	if err != nil {
		return err
	}
	addrIndexKey := append(addrPrefix, nameKey...)
	if err := k.addrIndex.Set(ctx, addrIndexKey, record); err != nil {
		return err
	}

	nameUpdateEvent := types.NewEventNameUpdate(addr.String(), name, restrict)

	return ctx.EventManager().EmitTypedEvent(nameUpdateEvent)
}

// GetRecordByName resolves a record by name.
func (k Keeper) GetRecordByName(ctx sdk.Context, name string) (record *types.NameRecord, err error) {
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return nil, err
	}
	nameKey, err := types.GetNameKeyPrefix(normalizedName)
	if err != nil {
		return nil, types.ErrNameNotBound.Wrapf("failed to get name key prefix for name %q: %v", normalizedName, err)
	}
	namerecord, err := k.nameRecords.Get(ctx, nameKey)
	if err != nil {
		return nil, types.ErrNameNotBound
	}
	return &namerecord, nil
}

func getNameRecord(ctx sdk.Context, keeper Keeper, key []byte) (record *types.NameRecord, err error) {
	store := ctx.KVStore(keeper.storeKey)
	if !store.Has(key) {
		return nil, types.ErrNameNotBound
	}
	bz := store.Get(key)
	record = &types.NameRecord{}
	err = keeper.cdc.Unmarshal(bz, record)
	return record, err
}

// NameExists returns true if store contains a record for the given name.
func (k Keeper) NameExists(ctx sdk.Context, name string) bool {
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return false
	}
	nameKey, err := types.GetNameKeyPrefix(normalizedName)
	if err != nil {
		return false
	}
	exists, _ := k.nameRecords.Has(ctx, nameKey)
	return exists
}

// GetRecordsByAddress looks up all names bound to an address.
func (k Keeper) GetRecordsByAddress(ctx sdk.Context, address sdk.AccAddress) (types.NameRecords, error) {
	addrPrefix, err := types.GetAddressKeyPrefix(address)
	if err != nil {
		return nil, err
	}
	var records []types.NameRecord
	err = k.addrIndex.Walk(ctx, nil, func(key []byte, record types.NameRecord) (bool, error) {
		if bytes.HasPrefix(key, addrPrefix) {
			records = append(records, record)
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

// DeleteRecord removes a name record from the kvstore.
func (k Keeper) DeleteRecord(ctx sdk.Context, name string) error {
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return err
	}
	nameKey, err := types.GetNameKeyPrefix(normalizedName)
	if err != nil {
		return types.ErrNameNotBound.Wrapf("failed to get name key prefix for name %q: %v", normalizedName, err)
	}

	record, err := k.nameRecords.Get(ctx, nameKey)
	if err != nil {
		return types.ErrNameNotBound
	}

	addr, err := sdk.AccAddressFromBech32(record.Address)
	if err != nil {
		return err
	}

	// Delete address index
	addrPrefix, err := types.GetAddressKeyPrefix(addr)
	if err != nil {
		return err
	}
	addrIndexKey := append(addrPrefix, nameKey...)
	if err := k.addrIndex.Remove(ctx, addrIndexKey); err != nil {
		return err
	}

	// Delete name record
	if err := k.nameRecords.Remove(ctx, nameKey); err != nil {
		return err
	}

	nameUnboundEvent := types.NewEventNameUnbound(record.Address, name, record.Restricted)

	return ctx.EventManager().EmitTypedEvent(nameUnboundEvent)
}

// IterateRecords iterates over all the stored name records and passes them to a callback function.
func (k Keeper) IterateRecords(ctx sdk.Context, prefix []byte, handle func(record types.NameRecord) error) error {
	rng := (&collections.Range[[]byte]{}).
		StartInclusive(prefix).
		EndExclusive(storetypes.PrefixEndBytes(prefix))

	return k.nameRecords.Walk(ctx, rng, func(_ []byte, record types.NameRecord) (bool, error) {
		if err := handle(record); err != nil {
			return true, err
		}
		return false, nil
	})
}

// Normalize returns a name is storage format.
func (k Keeper) Normalize(ctx sdk.Context, name string) (string, error) {
	normalized := types.NormalizeName(name)
	if !types.IsValidName(normalized) {
		return "", types.ErrNameInvalid
	}
	segments := strings.Split(normalized, ".")
	if len(segments) > int(k.GetMaxNameLevels(ctx)) {
		return "", types.ErrNameHasTooManySegments
	}
	for _, segment := range segments {
		segLen := len(segment)
		isUUID := types.IsValidUUID(segment)

		if segLen < int(k.GetMinSegmentLength(ctx)) {
			return "", types.ErrNameSegmentTooShort
		}
		if segLen > int(k.GetMaxSegmentLength(ctx)) && !isUUID {
			return "", types.ErrNameSegmentTooLong
		}
	}
	return normalized, nil
}

func (k Keeper) addRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict, isModifiable bool) error {
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return err
	}
	if err := types.ValidateAddress(addr); err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	key, err := types.GetNameKeyPrefix(normalizedName)
	if err != nil {
		return types.ErrNameNotBound.Wrapf("failed to get name key prefix for name %q: %v", normalizedName, err)
	}

	if !isModifiable {
		exists, err := k.nameRecords.Has(ctx, key)
		if err != nil {
			return err
		}
		if exists {
			return types.ErrNameAlreadyBound
		}
	}
	if isModifiable {
		existing, err := k.nameRecords.Get(ctx, key)
		if err == nil && existing.Address != addr.String() {
			// Remove old address index
			oldAddr, err := sdk.AccAddressFromBech32(existing.Address)
			if err != nil {
				return err
			}
			oldAddrPrefix, err := types.GetAddressKeyPrefix(oldAddr)
			if err != nil {
				return err
			}
			oldAddrKey := append(oldAddrPrefix, key...)
			if err := k.addrIndex.Remove(ctx, oldAddrKey); err != nil {
				return err
			}
		}
	}
	record := types.NewNameRecord(normalizedName, addr, restrict)
	if err := record.Validate(); err != nil {
		return err
	}
	// Set name record
	if err := k.nameRecords.Set(ctx, key, record); err != nil {
		return err
	}
	// Set address index
	addrPrefix, err := types.GetAddressKeyPrefix(addr)
	if err != nil {
		return err
	}
	addrIndexKey := append(addrPrefix, key...) // [0x04] :: [addr-bytes] :: [name-key-bytes]
	if err := k.addrIndex.Set(ctx, addrIndexKey, record); err != nil {
		return err
	}

	return nil
}

// DeleteInvalidAddressIndexEntries is only for the rust upgrade. It goes over all the address -> name entries and
// deletes any that are no longer accurate.
func (k Keeper) DeleteInvalidAddressIndexEntries(ctx sdk.Context) {
	logger := k.Logger(ctx)
	logger.Info("Checking address -> name index entries.")

	keepCount := 0
	var toDelete [][]byte

	extractNameKey := func(key []byte) []byte {
		// byte 1 is the type byte (0x05), it's ignored here.
		// The 2nd byte is the length of the address that immediately follows it.
		// The name key starts directly after the address, and is the rest of the key.
		addrLen := int(key[1])
		nameKeyStart := addrLen + 2
		return key[nameKeyStart:]
	}

	err := k.addrIndex.Walk(ctx, nil, func(key []byte, record types.NameRecord) (stop bool, err error) {
		nameKey := extractNameKey(key)
		if nameKey == nil {
			toDelete = append(toDelete, key)
			return false, nil
		}

		// Check if name record exists
		exists, err := k.nameRecords.Has(ctx, nameKey)
		if err != nil {
			return true, err
		}
		if !exists {
			toDelete = append(toDelete, key)
			return false, nil
		}

		// Check if index value matches name record
		mainRecord, err := k.nameRecords.Get(ctx, nameKey)
		if err != nil {
			return true, err
		}
		if !bytes.Equal(k.cdc.MustMarshal(&record), k.cdc.MustMarshal(&mainRecord)) {
			toDelete = append(toDelete, key)
			return false, nil
		}

		keepCount++
		return false, nil
	})

	if err != nil {
		logger.Error("Error during index validation", "error", err)
		return
	}

	if len(toDelete) == 0 {
		logger.Info(fmt.Sprintf("All %d index entries are valid", keepCount))
		return
	}

	logger.Info(fmt.Sprintf("Found %d invalid entries, deleting", len(toDelete)))
	for _, key := range toDelete {
		if err := k.addrIndex.Remove(ctx, key); err != nil {
			logger.Error("Failed to delete index entry", "key", key, "error", err)
		}
	}

	logger.Info(fmt.Sprintf("Done checking address -> name index entries. Deleted %d invalid entries and kept %d valid entries.", len(toDelete), keepCount))
}

func (k *Keeper) GetNameRecord(ctx sdk.Context, key []byte) (types.NameRecord, error) {
	return k.nameRecords.Get(ctx, key)
}

func (k *Keeper) GetAddrIndexRecord(ctx sdk.Context, key []byte) (types.NameRecord, error) {
	return k.addrIndex.Get(ctx, key)
}

func (k Keeper) CreateRootName(ctx sdk.Context, name, owner string, restricted bool) error {
	// err is suppressed because it returns an error on not found.  TODO - Remove use of error for not found
	// Check root name
	if k.NameExists(ctx, name) {
		return types.ErrNameAlreadyBound
	}
	addr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return err
	}
	logger := k.Logger(ctx)

	// Create all intermediate domains
	n := ""
	segments := strings.Split(name, ".")
	for i := len(segments) - 1; i >= 0; i-- {
		n = strings.Join([]string{segments[i], n}, ".")
		n = strings.TrimRight(n, ".")

		if !k.NameExists(ctx, n) {
			if err := k.SetNameRecord(ctx, n, addr, restricted); err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("Created %s with owner %s", n, owner))
		} else {
			logger.Info(fmt.Sprintf("Domain %s already exists, skipping", n))
		}
	}

	return nil
}
