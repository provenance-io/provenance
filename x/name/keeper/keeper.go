package keeper

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// NameRecordIndexes defines indexes for name records
type NameRecordIndexes struct {
	// AddrIndex maps address -> list of names
	AddrIndex *indexes.Multi[[]byte, string, types.NameRecord]
}

// List implements collections.Indexes
func (i NameRecordIndexes) IndexesList() []collections.Index[string, types.NameRecord] {
	return []collections.Index[string, types.NameRecord]{i.AddrIndex}
}

// Keeper defines the name module Keeper.
type Keeper struct {
	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// the signing authority for the gov proposals
	authority string

	// Attribute keeper
	attrKeeper types.AttributeKeeper

	storeService store.KVStoreService

	// Schema definition
	schema collections.Schema

	// Primary: name (hashed) -> NameRecord, indexed by addr
	nameRecords *collections.IndexedMap[string, types.NameRecord, NameRecordIndexes]

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
		collections.NewPrefix(types.AddressKeyPrefix),
		"addr_index",
		collections.BytesKey, // Index key is []byte
		types.HashedStringKeyCodec{},
		func(name string, record types.NameRecord) ([]byte, error) {
			addr, err := sdk.AccAddressFromBech32(record.Address)
			if err != nil {
				return nil, err
			}
			return addr, nil // addr is sdk.AccAddress, which is a []byte type
		},
	)

	indexes := NameRecordIndexes{
		AddrIndex: addrIndex,
	}

	// Create name records collection
	nameRecords := collections.NewIndexedMap(
		sb,
		collections.NewPrefix(types.NameKeyPrefix),
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
	if err := types.ValidateAddress(addr); err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}
	_, err = k.nameRecords.Get(ctx, normalizedName)
	if err != nil {
		return types.ErrNameNotBound
	}
	record := types.NewNameRecord(normalizedName, addr, restrict)
	if err := record.Validate(); err != nil {
		return err
	}
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
		return nil, types.ErrNameNotBound
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
	iter, err := k.nameRecords.Indexes.AddrIndex.MatchExact(ctx, address)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var records []types.NameRecord
	for ; iter.Valid(); iter.Next() {
		pk, err := iter.PrimaryKey()
		if err != nil {
			return nil, err
		}
		record, err := k.nameRecords.Get(ctx, pk)
		if err != nil {
			return nil, err
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
		return types.ErrNameNotBound
	}
	if err := k.nameRecords.Remove(ctx, normalizedName); err != nil {
		return err
	}
	nameUnboundEvent := types.NewEventNameUnbound(record.Address, name, record.Restricted)
	return ctx.EventManager().EmitTypedEvent(nameUnboundEvent)
}

func (k Keeper) IterateRecords(ctx sdk.Context, prefix []byte, handle func(record types.NameRecord) error) error {
	return k.nameRecords.Walk(ctx, nil, func(name string, record types.NameRecord) (bool, error) {
		if prefix != nil && !bytes.HasPrefix([]byte(name), prefix) {
			return false, nil
		}
		if err := handle(record); err != nil {
			return true, err
		}
		return false, nil
	})
}

// Normalize returns a name in storage format.
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

// DeleteInvalidAddressIndexEntries is only for the rust upgrade. It goes over all the address -> name entries and
// deletes any that are no longer accurate.
func (k Keeper) DeleteInvalidAddressIndexEntries(ctx sdk.Context) {
	logger := k.Logger(ctx)
	logger.Info("Checking address -> name index entries.")
	keepCount := 0
	var toDelete [][]byte

	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.AddressKeyPrefix, storetypes.PrefixEndBytes(types.AddressKeyPrefix))
	if err != nil {
		logger.Error("failed to create iterator for address index", "error", err)
		return
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		if len(key) < 23 { // 1 (prefix) + 1 (len) + 20 (min addr) + 20 (hash)
			toDelete = append(toDelete, key)
			continue
		}
		addrLen := int(key[1])
		if len(key) != 2+addrLen+20 {
			toDelete = append(toDelete, key)
			continue
		}
		addrBz := key[2 : 2+addrLen]
		nameHash := key[2+addrLen:]

		nameKey := append(types.NameKeyPrefix, nameHash...)
		exists, err := store.Has(nameKey)
		if err != nil || !exists {
			toDelete = append(toDelete, key)
			continue
		}

		bz, err := store.Get(nameKey)
		if err != nil {
			toDelete = append(toDelete, key)
			continue
		}

		var record types.NameRecord
		if err := k.cdc.Unmarshal(bz, &record); err != nil {
			toDelete = append(toDelete, key)
			continue
		}

		fullHash := sha256.Sum256([]byte(record.Name))
		computedHash := fullHash[:20]
		if !bytes.Equal(computedHash, nameHash) {
			toDelete = append(toDelete, key)
			continue
		}

		addr, err := sdk.AccAddressFromBech32(record.Address)
		if err != nil || !bytes.Equal(addr.Bytes(), addrBz) {
			toDelete = append(toDelete, key)
			continue
		}
		keepCount++
	}

	if len(toDelete) == 0 {
		logger.Info(fmt.Sprintf("All %d index entries are valid", keepCount))
		return
	}

	logger.Info(fmt.Sprintf("Found %d invalid entries, deleting", len(toDelete)))
	for _, key := range toDelete {
		store.Delete(key)
	}
	logger.Info(fmt.Sprintf("Done checking address -> name index entries. Deleted %d invalid entries and kept %d valid entries.", len(toDelete), keepCount))
}

func (k *Keeper) GetNameRecord(ctx sdk.Context, key string) (types.NameRecord, error) {
	return k.nameRecords.Get(ctx, key)
}
func (k Keeper) GetAddrIndex() *indexes.Multi[[]byte, string, types.NameRecord] {
	return k.nameRecords.Indexes.AddrIndex
}

func (k Keeper) CreateRootName(ctx sdk.Context, name, owner string, restricted bool) error {
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
