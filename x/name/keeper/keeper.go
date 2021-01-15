package keeper

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// To check whether accounts exist for addresses.
	authKeeper types.AccountKeeper

	// Key to access the key-value store from sdk.Context.
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc codec.BinaryMarshaler
}

// NewKeeper returns a name keeper. It handles:
// - managing a heirarchy of names
// - enforcing permissions for name creation/deletion
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, paramSpace paramtypes.Subspace, authKeeper types.AccountKeeper) Keeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   key,
		paramSpace: paramSpace,
		authKeeper: authKeeper,
		cdc:        cdc,
	}
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return keeper.Logger(ctx).With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ResolvesTo to determines whether a name resolves to a given address.
func (keeper Keeper) ResolvesTo(ctx sdk.Context, name string, addr sdk.AccAddress) bool {
	stored, err := keeper.getRecordByName(ctx, name)
	if err != nil {
		return false
	}
	return addr.String() == stored.Address
}

// SetNameRecord binds a name to an address. An error is returned if no account exists for the address.
func (keeper Keeper) setNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	var err error
	if name, err = keeper.Normalize(ctx, name); err != nil {
		return err
	}
	if account := keeper.authKeeper.GetAccount(ctx, addr); account == nil {
		return types.ErrInvalidAddress
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
	bz, err := types.ModuleCdc.MarshalBinaryBare(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// setGenesisRecord will allow a record to be created for an address that does not exist if in proper format
func (keeper Keeper) setGenesisRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	var err error
	if addr.Empty() {
		return types.ErrNameInvalid
	}
	if err = sdk.VerifyAddressFormat(addr); err != nil {
		return err
	}
	if name, err = keeper.Normalize(ctx, name); err != nil {
		return err
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
	bz, err := types.ModuleCdc.MarshalBinaryBare(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

func (keeper Keeper) getRecordByName(ctx sdk.Context, name string) (record *types.NameRecord, err error) {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return nil, err
	}
	store := ctx.KVStore(keeper.storeKey)
	if !store.Has(key) {
		return nil, types.ErrNameNotBound
	}
	bz := store.Get(key)
	record = &types.NameRecord{}
	err = types.ModuleCdc.UnmarshalBinaryBare(bz, record)
	return record, err
}

// Logger returns a module-specific logger.
func (keeper Keeper) nameExists(ctx sdk.Context, name string) bool {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return false
	}
	store := ctx.KVStore(keeper.storeKey)
	return store.Has(key)
}

func (keeper Keeper) getRecordsByAddress(ctx sdk.Context, address sdk.AccAddress) (types.NameRecords, error) {
	//
	// TODO: Refactor this once name records are indexed by address...
	//
	// Return value data structure.
	records := types.NameRecords{}
	// Handler that adds records if account address matches.
	appendToRecords := func(record types.NameRecord) error {
		if record.Address == address.String() {
			records = append(records, record)
		}
		return nil
	}
	// Collect and return all names that match.
	if err := keeper.IterateRecords(ctx, appendToRecords); err != nil {
		return records, err
	}
	return records, nil
}

// Delete a name record from the kvstore.
func (keeper Keeper) deleteRecord(ctx sdk.Context, name string) error {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return err
	}
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(key)
	return nil
}

// IterateRecords iterates over all the stored name records and passes them to a callback function.
func (keeper Keeper) IterateRecords(ctx sdk.Context, handle Handler) error {
	// Init a name record iterator
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.NameKeyPrefix)
	defer iterator.Close()
	// Iterate over records, processing callbacks.
	for ; iterator.Valid(); iterator.Next() {
		record := types.NameRecord{}
		if err := types.ModuleCdc.UnmarshalBinaryBare(iterator.Value(), &record); err != nil {
			return err
		}
		if err := handle(record); err != nil {
			return err
		}
	}
	return nil
}

// Normalize returns a name is storage format.
func (keeper Keeper) Normalize(ctx sdk.Context, name string) (string, error) {
	var comps []string
	for _, comp := range strings.Split(name, ".") {
		comp = strings.ToLower(strings.TrimSpace(comp))
		lenComp := uint32(len(comp))
		if lenComp < keeper.GetMinSegmentLength(ctx) {
			return "", types.ErrNameSegmentTooShort
		}
		if lenComp > keeper.GetMaxSegmentLength(ctx) {
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
