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
	stored, err := keeper.Resolve(ctx.Context(), &types.QueryResolveRequest{Name: name})
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

func (keeper Keeper) getRecordByName(ctx sdk.Context, name string) (record *types.NameRecord, err error) {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return nil, err
	}
	store := ctx.KVStore(keeper.storeKey)
	if !store.Has(key) {
		return nil, nil
	}
	bz := store.Get(key)
	record = &types.NameRecord{}
	err = types.ModuleCdc.UnmarshalBinaryBare(bz, record)
	return record, err
}

func (keeper Keeper) getRecordsByAddress(ctx sdk.Context, address sdk.AccAddress) (records types.NameRecords, err error) {
	return
}

func (keeper Keeper) deleteRecord(ctx sdk.Context, name string) (err error) {
	return
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

/*
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		keeper.Logger(ctx).Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Fetch the name record from the name keeper.
	record, err := keeper.Resolve(ctx, msg.RootName)
	if err != nil {
		keeper.Logger(ctx).Error("unable to find root name record", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Ensure that if the root name is restricted, it resolves to the given root address (message signer).
	if record.Restricted && !keeper.ResolvesTo(ctx, msg.RootName, msg.RootAddress) {
		errm := "root name is restricted and does not resolve to the provided root address"
		keeper.Logger(ctx).Error(errm)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, errm)
	}
	// Combine names and check for existing record
	name := fmt.Sprintf("%s.%s", msg.Name, msg.RootName)
	existing, err := keeper.Resolve(ctx, name)
	// Handle failures
	if err != nil && err != ErrNameNotBound {
		keeper.Logger(ctx).Error("resolve failure", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Bind name to address
	if err == ErrNameNotBound {
		if berr := keeper.BindName(ctx, name, msg.Address, msg.Restricted); berr != nil {
			keeper.Logger(ctx).Error("unable to bind name", "err", berr)
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, berr.Error())
		}
		// Emit event and return
		events := []sdk.Event{sdk.NewEvent(
			eventTypeNameBound,
			sdk.NewAttribute(eventAttributeNameKeyAttribute, name),
			sdk.NewAttribute(eventAddressKeyAttribute, msg.Address.String()),
		)}
		return &sdk.Result{Events: events}, nil
	}
	// Name is already bound.
	keeper.Logger(ctx).Error("name already bound", "name", name, "address", existing)
	return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name already bound")

*/
