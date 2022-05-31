package keeper

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	uuid "github.com/google/uuid"

	"github.com/provenance-io/provenance/x/name/types"
)

// Handler is a name record handler function for use with IterateRecords.
type Handler func(record types.NameRecord) error

// Keeper defines the name module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set account specific params
	paramSpace paramtypes.Subspace

	// Key to access the key-value store from sdk.Context.
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc codec.BinaryCodec
}

// NewKeeper returns a name keeper. It handles:
// - managing a hierarchy of names
// - enforcing permissions for name creation/deletion
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
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ResolvesTo to determines whether a name resolves to a given address.
func (keeper Keeper) ResolvesTo(ctx sdk.Context, name string, addr sdk.AccAddress) bool { // nolint:interfacer
	stored, err := keeper.GetRecordByName(ctx, name)
	if err != nil {
		return false
	}
	return addr.String() == stored.Address
}

// SetNameRecord binds a name to an address.
func (keeper Keeper) SetNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	var err error
	if name, err = keeper.Normalize(ctx, name); err != nil {
		return err
	}
	if err = types.ValidateAddress(addr); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidAddress, err.Error())
	}
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return err
	}
	store := ctx.KVStore(keeper.storeKey)
	if store.Has(key) {
		return types.ErrNameAlreadyBound
	}
	record := types.NewNameRecord(name, addr, restrict)
	if err = record.ValidateBasic(); err != nil {
		return err
	}
	bz, err := keeper.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	// Now index by address
	addrPrefix, err := types.GetAddressKeyPrefix(addr)
	if err != nil {
		return err
	}
	addrPrefix = append(addrPrefix, key...) // [0x04] :: [addr-bytes] :: [name-key-bytes]
	store.Set(addrPrefix, bz)

	nameBoundEvent := types.NewEventNameBound(record.Address, name, record.Restricted)

	if err := ctx.EventManager().EmitTypedEvent(nameBoundEvent); err != nil {
		return err
	}

	return nil
}

// GetRecordByName resolves a record by name.
func (keeper Keeper) GetRecordByName(ctx sdk.Context, name string) (record *types.NameRecord, err error) {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return nil, err
	}
	return getNameRecord(ctx, keeper, key)
}

// GetRecordByName resolves a record by name.
func (keeper Keeper) GetRecordByNameLegacy(ctx sdk.Context, name string) (record *types.NameRecord, err error) {
	key, err := types.GetNameKeyPrefixLegacyAmino(name)
	if err != nil {
		return nil, err
	}
	return getNameRecord(ctx, keeper, key)
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
func (keeper Keeper) NameExists(ctx sdk.Context, name string) bool {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return false
	}
	store := ctx.KVStore(keeper.storeKey)
	return store.Has(key)
}

// GetRecordsByAddress looks up all names bound to an address.
func (keeper Keeper) GetRecordsByAddress(ctx sdk.Context, address sdk.AccAddress) (types.NameRecords, error) {
	// Return value data structure.
	records := types.NameRecords{}
	// Handler that adds records if account address matches.
	appendToRecords := func(record types.NameRecord) error {
		if record.Address == address.String() {
			records = append(records, record)
		}
		return nil
	}
	// Calculate address prefix
	addrPrefix, err := types.GetAddressKeyPrefix(address)
	if err != nil {
		return nil, err
	}
	// Collect and return all names that match.
	if err := keeper.IterateRecords(ctx, addrPrefix, appendToRecords); err != nil {
		return records, err
	}
	return records, nil
}

// DeleteRecord removes a name record from the kvstore.
func (keeper Keeper) DeleteRecord(ctx sdk.Context, name string) error {
	// Need the record to clear the address index
	record, err := keeper.GetRecordByName(ctx, name)
	if err != nil {
		return err
	}
	address, err := sdk.AccAddressFromBech32(record.Address)
	if err != nil {
		return err
	}
	// Delete the main name record
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return err
	}
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(key)
	// Delete the address index record
	addrPrefix, err := types.GetAddressKeyPrefix(address)
	if err != nil {
		return err
	}
	addrPrefix = append(addrPrefix, key...) // [0x02] :: [addr-bytes] :: [name-key-bytes]
	if store.Has(addrPrefix) {
		store.Delete(addrPrefix)
	}

	nameUnboundEvent := types.NewEventNameUnbound(record.Address, name, record.Restricted)

	if err := ctx.EventManager().EmitTypedEvent(nameUnboundEvent); err != nil {
		return err
	}

	return nil
}

// IterateRecords iterates over all the stored name records and passes them to a callback function.
func (keeper Keeper) IterateRecords(ctx sdk.Context, prefix []byte, handle Handler) error {
	// Init a name record iterator
	store := ctx.KVStore(keeper.storeKey)
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
			if err := keeper.cdc.Unmarshal(iterator.Value(), &record); err != nil {
				return err
			}
		}
		if err := handle(record); err != nil {
			return err
		}
	}
	return nil
}

// Normalize returns a name is storage format.
func (keeper Keeper) Normalize(ctx sdk.Context, name string) (string, error) {
	comps := make([]string, 0)
	for _, comp := range strings.Split(name, ".") {
		comp = strings.ToLower(strings.TrimSpace(comp))
		lenComp := uint32(len(comp))
		isUUID := isValidUUID(comp)
		if lenComp < keeper.GetMinSegmentLength(ctx) {
			return "", types.ErrNameSegmentTooShort
		}
		if lenComp > keeper.GetMaxSegmentLength(ctx) && !isUUID {
			return "", types.ErrNameSegmentTooLong
		}
		if !isValid(comp) {
			return "", types.ErrNameInvalid
		}
		comps = append(comps, comp)
	}
	if uint32(len(comps)) > keeper.GetMaxNameLevels(ctx) {
		return "", types.ErrNameHasTooManySegments
	}
	return strings.Join(comps, "."), nil
}

// Check whether a name component is valid
func isValid(s string) bool {
	// Allow valid UUID
	if isValidUUID(s) {
		return true
	}
	// Only allow a single dash if not a UUID
	if strings.Count(s, "-") > 1 {
		return false
	}
	for _, c := range s {
		if c == '-' {
			continue
		}
		if !unicode.IsGraphic(c) {
			return false
		}
		if !unicode.IsLower(c) && !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// Ensure a string can be parsed into a UUID.
func isValidUUID(s string) bool {
	if _, err := uuid.Parse(s); err != nil {
		return false
	}
	return true
}
