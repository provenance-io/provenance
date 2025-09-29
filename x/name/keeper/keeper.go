package keeper

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// Keeper defines the name module Keeper.
type Keeper struct {
	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// the signing authority for the gov proposals
	authority string

	// Attribute keeper
	attrKeeper types.AttributeKeeper

	// storeService abstracts access to the module's KVStore.
	storeService store.KVStoreService
	// Schema definition
	schema collections.Schema
	// Primary: name (hashed) -> NameRecord, indexed by addr
	nameRecords *collections.IndexedMap[string, types.NameRecord, types.NameRecordIndexes]
	// paramsStore manages the module's configurable parameters.
	paramsStore collections.Item[types.Params]
}

// NewKeeper returns a name keeper. It handles:
// - managing a hierarchy of names
// - enforcing permissions for name creation/deletion
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	// Create address index
	addrIndex := indexes.NewMulti(
		sb,
		types.AddressKeyPrefix,
		"addr_index",
		collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey),
		collections.StringKey,
		func(name string, record types.NameRecord) (collections.Pair[sdk.AccAddress, string], error) {
			addr, err := sdk.AccAddressFromBech32(record.Address)
			if err != nil {
				return collections.Pair[sdk.AccAddress, string]{}, err
			}
			return collections.Join(addr, name), nil
		},
	)

	indexes := types.NameRecordIndexes{
		AddrIndex: addrIndex,
	}

	// Create name records collection
	nameRecords := collections.NewIndexedMap(
		sb,
		types.NameKeyPrefix,
		"name_records",
		types.HashedStringKeyCodec{},
		codec.CollValue[types.NameRecord](cdc),
		indexes,
	)
	// Create params collection
	params := collections.NewItem(
		sb,
		types.NameParamStoreKey,
		"params",
		codec.CollValue[types.Params](cdc),
	)

	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		nameRecords:  nameRecords,
		paramsStore:  params,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(fmt.Sprintf("name module schema build failed: %v", err))
	}
	k.schema = schema

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

// ResolvesTo determines whether a name resolves to a given address.
func (k Keeper) ResolvesTo(ctx sdk.Context, name string, addr sdk.AccAddress) bool {
	stored, err := k.GetRecordByName(ctx, name)
	if err != nil {
		return false
	}
	return addr.String() == stored.Address
}

// SetNameRecord binds a name to an address.
func (k Keeper) SetNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return err
	}
	if err = types.ValidateAddress(addr); err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}
	exists, err := k.nameRecords.Has(ctx, normalizedName)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrNameAlreadyBound
	}
	record := types.NewNameRecord(normalizedName, addr, restrict)
	if err := record.Validate(); err != nil {
		return err
	}

	if err := k.nameRecords.Set(ctx, normalizedName, record); err != nil {
		return err
	}

	nameBoundEvent := types.NewEventNameBound(addr.String(), normalizedName, restrict)
	return ctx.EventManager().EmitTypedEvent(nameBoundEvent)
}

// UpdateNameRecord updates the owner address and restricted flag on a name.
func (k Keeper) UpdateNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return err
	}
	if err = types.ValidateAddress(addr); err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}
	exists, err := k.nameRecords.Has(ctx, normalizedName)
	if err != nil {
		return err
	}
	if !exists {
		return types.ErrNameNotBound
	}

	// Create and validate updated record
	record := types.NewNameRecord(normalizedName, addr, restrict)
	if err := record.Validate(); err != nil {
		return err
	}
	// Update the record.
	if err := k.nameRecords.Set(ctx, normalizedName, record); err != nil {
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
	namerecord, err := k.nameRecords.Get(ctx, normalizedName)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, types.ErrNameNotBound
		}
		return nil, err
	}
	return &namerecord, nil
}

// NameExists returns true if store contains a record for the given name.
func (k Keeper) NameExists(ctx sdk.Context, name string) bool {
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return false
	}
	exists, _ := k.nameRecords.Has(ctx, normalizedName)
	return exists
}

// GetRecordsByAddress looks up all names bound to an address.
func (k Keeper) GetRecordsByAddress(ctx sdk.Context, address sdk.AccAddress) (types.NameRecords, error) {
	var records types.NameRecords

	// We create a prefix of the composite index key (address + name) by fixing the address part
	// with PairPrefix, then build a PrefixedPairRange to efficiently query all entries matching
	// that address in the multi-index (one-to-many relationship)
	refKeyPrefix := collections.PairPrefix[sdk.AccAddress, string](address)
	prefixRange := collections.NewPrefixedPairRange[
		collections.Pair[sdk.AccAddress, string],
		string,
	](refKeyPrefix)

	iter, err := k.nameRecords.Indexes.AddrIndex.Iterate(ctx, prefixRange)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		primaryKey, err := iter.PrimaryKey()
		if err != nil {
			continue
		}
		record, err := k.nameRecords.Get(ctx, primaryKey)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

// DeleteRecord removes a name record from the kvstore.
func (k Keeper) DeleteRecord(ctx sdk.Context, name string) error {
	normalizedName, err := k.Normalize(ctx, name)
	if err != nil {
		return err
	}
	record, err := k.nameRecords.Get(ctx, normalizedName)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrNameNotBound
		}
		return err
	}
	if err := k.nameRecords.Remove(ctx, normalizedName); err != nil {
		return err
	}
	nameUnboundEvent := types.NewEventNameUnbound(record.Address, name, record.Restricted)
	return ctx.EventManager().EmitTypedEvent(nameUnboundEvent)
}

func (k Keeper) IterateRecords(ctx sdk.Context, handle func(record types.NameRecord) error) error {
	err := k.nameRecords.Walk(ctx, nil, func(_ string, record types.NameRecord) (bool, error) {
		if err := handle(record); err != nil {
			return true, err
		}
		return false, nil
	})

	return err
}

// Normalize returns a name in storage format.
func (k Keeper) Normalize(ctx sdk.Context, name string) (string, error) {
	normalized := types.NormalizeName(name)
	if !types.IsValidName(normalized) {
		return "", types.ErrNameInvalid
	}
	segCount := uint32(0)
	for _, segment := range strings.Split(normalized, ".") {
		segCount++
		segLen := len(segment)
		isUUID := types.IsValidUUID(segment)
		if segLen < int(k.GetMinSegmentLength(ctx)) {
			return "", types.ErrNameSegmentTooShort
		}
		if segLen > int(k.GetMaxSegmentLength(ctx)) && !isUUID {
			return "", types.ErrNameSegmentTooLong
		}
	}
	if segCount > k.GetMaxNameLevels(ctx) {
		return "", types.ErrNameHasTooManySegments
	}
	return normalized, nil
}

func (k Keeper) GetAddrIndex() *indexes.Multi[collections.Pair[sdk.AccAddress, string], string, types.NameRecord] {
	return k.nameRecords.Indexes.AddrIndex
}

func (k Keeper) CreateRootName(ctx sdk.Context, name, owner string, restricted bool) error {
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
