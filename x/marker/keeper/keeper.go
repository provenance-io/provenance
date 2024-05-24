package keeper

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// Handler is a handler function for use with IterateRecords.
type Handler func(record types.MarkerAccountI) error

// MarkerKeeperI provides a read/write iterate interface to marker acccounts in the auth account keeper store
type MarkerKeeperI interface {
	// Returns a new marker instance with the address and baseaccount assigned.  Does not save to auth store
	NewMarker(sdk.Context, types.MarkerAccountI) types.MarkerAccountI

	// GetMarker looks up a marker by a given address
	GetMarker(sdk.Context, sdk.AccAddress) (types.MarkerAccountI, error)
	// Set a marker in the auth account store
	SetMarker(sdk.Context, types.MarkerAccountI)
	// Remove a marker from the auth account store
	RemoveMarker(sdk.Context, types.MarkerAccountI)

	GetEscrow(sdk.Context, types.MarkerAccountI) sdk.Coins

	// IterateMarker processes all markers with the given handler function.
	IterateMarkers(sdk.Context, func(types.MarkerAccountI) bool)

	// GetAuthority returns the signing authority
	GetAuthority() string
}

// Keeper defines the name module Keeper
type Keeper struct {
	// To check whether accounts exist for addresses.
	authKeeper types.AccountKeeper

	// To check whether accounts exist for addresses.
	authzKeeper types.AuthzKeeper

	// To handle movement of coin between accounts and check total supply
	bankKeeper types.BankKeeper

	// To pass through grant creation for callers with admin access on a marker.
	feegrantKeeper types.FeeGrantKeeper

	// To access attributes for addresses
	attrKeeper types.AttrKeeper
	// To access names and normalize required attributes
	nameKeeper types.NameKeeper

	// Key to access the key-value store from sdk.Context.
	storeKey storetypes.StoreKey

	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// the signing authority for the gov proposals
	authority string

	markerModuleAddr sdk.AccAddress

	ibcTransferModuleAddr sdk.AccAddress

	feeCollectorAddr sdk.AccAddress

	// Used to transfer the ibc marker
	ibcTransferServer types.IbcTransferMsgServer

	// reqAttrBypassAddrs is a set of addresses that are allowed to bypass the required attribute check.
	// When sending to one of these, if there are required attributes, it behaves as if the addr has them;
	// if there aren't required attributes, the sender still needs transfer permission.
	// When sending from one of these, if there are required attributes, the destination must have them;
	// if there aren't required attributes, it behaves as if the sender has transfer permission.
	reqAttrBypassAddrs types.ImmutableAccAddresses

	// groupChecker provides a way to check if an account is in a group.
	groupChecker types.GroupChecker
}

// NewKeeper returns a marker keeper. It handles:
// - managing MarkerAccounts
// - enforcing permissions for marker creation/deletion/management
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	authzKeeper types.AuthzKeeper,
	feegrantKeeper types.FeeGrantKeeper,
	attrKeeper types.AttrKeeper,
	nameKeeper types.NameKeeper,
	ibcTransferServer types.IbcTransferMsgServer,
	reqAttrBypassAddrs []sdk.AccAddress,
	checker types.GroupChecker,
) Keeper {
	rv := Keeper{
		authKeeper:            authKeeper,
		authzKeeper:           authzKeeper,
		bankKeeper:            bankKeeper,
		feegrantKeeper:        feegrantKeeper,
		attrKeeper:            attrKeeper,
		nameKeeper:            nameKeeper,
		storeKey:              key,
		cdc:                   cdc,
		authority:             authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		markerModuleAddr:      authtypes.NewModuleAddress(types.CoinPoolName),
		ibcTransferModuleAddr: authtypes.NewModuleAddress(ibctypes.ModuleName),
		feeCollectorAddr:      authtypes.NewModuleAddress(authtypes.FeeCollectorName),
		ibcTransferServer:     ibcTransferServer,
		reqAttrBypassAddrs:    types.NewImmutableAccAddresses(reqAttrBypassAddrs),
		groupChecker:          checker,
	}
	bankKeeper.AppendSendRestriction(rv.SendRestrictionFn)
	return rv
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

var _ MarkerKeeperI = &Keeper{}

// NewMarker returns a new marker instance with the address and baseaccount assigned.  Does not save to auth store
func (k Keeper) NewMarker(ctx sdk.Context, marker types.MarkerAccountI) types.MarkerAccountI {
	return k.authKeeper.NewAccount(ctx, marker).(types.MarkerAccountI)
}

// GetMarker looks up a marker by a given address
func (k Keeper) GetMarker(ctx sdk.Context, address sdk.AccAddress) (types.MarkerAccountI, error) {
	mac := k.authKeeper.GetAccount(ctx, address)
	if mac != nil {
		macc, ok := mac.(types.MarkerAccountI)
		if !ok {
			return nil, fmt.Errorf("account at %s is not a marker account", address.String())
		}
		return macc, nil
	}
	return nil, nil
}

// SetMarker sets a marker in the auth account store will panic if the marker account is not valid or
// if the auth module account keeper fails to marshall the account.
func (k Keeper) SetMarker(ctx sdk.Context, marker types.MarkerAccountI) {
	store := ctx.KVStore(k.storeKey)

	if err := marker.Validate(); err != nil {
		panic(err)
	}
	k.authKeeper.SetAccount(ctx, marker)
	store.Set(types.MarkerStoreKey(marker.GetAddress()), marker.GetAddress())
}

// RemoveMarker removes a marker from the auth account store. Note: if the account holds coins this will
// likely cause an invariant constraint violation for the coin supply
func (k Keeper) RemoveMarker(ctx sdk.Context, marker types.MarkerAccountI) {
	store := ctx.KVStore(k.storeKey)
	k.authKeeper.RemoveAccount(ctx, marker)

	k.RemoveNetAssetValues(ctx, marker.GetAddress())
	k.ClearSendDeny(ctx, marker.GetAddress())
	store.Delete(types.MarkerStoreKey(marker.GetAddress()))
}

// IterateMarkers iterates all markers with the given handler function.
func (k Keeper) IterateMarkers(ctx sdk.Context, cb func(marker types.MarkerAccountI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.MarkerStoreKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		account := k.authKeeper.GetAccount(ctx, iterator.Value())
		ma, ok := account.(types.MarkerAccountI)
		if !ok {
			panic(fmt.Errorf("invalid account type in marker account registry"))
		}
		if cb(ma) {
			break
		}
	}
}

// GetEscrow returns the balances of all coins held in escrow in the marker
func (k Keeper) GetEscrow(ctx sdk.Context, marker types.MarkerAccountI) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, marker.GetAddress())
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

// IsSendDeny returns true if sender address is denied for marker
func (k Keeper) IsSendDeny(ctx sdk.Context, markerAddr, senderAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.DenySendKey(markerAddr, senderAddr))
}

// AddSendDeny set sender address to denied for marker
func (k Keeper) AddSendDeny(ctx sdk.Context, markerAddr, senderAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.DenySendKey(markerAddr, senderAddr), []byte{})
}

// RemoveSendDeny removes sender address from marker deny list
func (k Keeper) RemoveSendDeny(ctx sdk.Context, markerAddr, senderAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.DenySendKey(markerAddr, senderAddr))
}

// ClearSendDeny removes all entries of a marker from a send deny list
func (k Keeper) ClearSendDeny(ctx sdk.Context, markerAddr sdk.AccAddress) {
	list := k.GetSendDenyList(ctx, markerAddr)
	for _, sender := range list {
		k.RemoveSendDeny(ctx, markerAddr, sender)
	}
}

// IterateMarkers  iterates all markers with the given handler function.
func (k Keeper) IterateSendDeny(ctx sdk.Context, handler func(key []byte) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.DenySendKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		if handler(iterator.Key()) {
			break
		}
	}
}

// GetSendDenyList gets the list of sender addresses from the marker's deny list
func (k Keeper) GetSendDenyList(ctx sdk.Context, markerAddr sdk.AccAddress) []sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.DenySendMarkerPrefix(markerAddr))
	list := []sdk.AccAddress{}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		_, denied := types.GetDenySendAddresses(iterator.Key())
		list = append(list, denied)
	}

	return list
}

// AddSetNetAssetValues adds a set of net asset values to a marker
func (k Keeper) AddSetNetAssetValues(ctx sdk.Context, marker types.MarkerAccountI, netAssetValues []types.NetAssetValue, source string) error {
	var errs []error
	for _, nav := range netAssetValues {
		if nav.Price.Denom == marker.GetDenom() {
			errs = append(errs, fmt.Errorf("net asset value denom cannot match marker denom %q", marker.GetDenom()))
			continue
		}

		if nav.Price.Denom != types.UsdDenom {
			_, err := k.GetMarkerByDenom(ctx, nav.Price.Denom)
			if err != nil {
				if err2 := nav.Validate(); err2 == nil {
					navEvent := types.NewEventSetNetAssetValue(marker.GetDenom(), nav.Price, nav.Volume, source)
					_ = ctx.EventManager().EmitTypedEvent(navEvent)
				}
				errs = append(errs, fmt.Errorf("net asset value denom does not exist: %w", err))
				continue
			}
		}

		if err := k.SetNetAssetValue(ctx, marker, nav, source); err != nil {
			errs = append(errs, fmt.Errorf("cannot set net asset value: %w", err))
		}
	}
	return errors.Join(errs...)
}

// SetNetAssetValue adds/updates a net asset value to marker
func (k Keeper) SetNetAssetValue(ctx sdk.Context, marker types.MarkerAccountI, netAssetValue types.NetAssetValue, source string) error {
	netAssetValue.UpdatedBlockHeight = uint64(ctx.BlockHeight())
	if err := netAssetValue.Validate(); err != nil {
		return err
	}

	setNetAssetValueEvent := types.NewEventSetNetAssetValue(marker.GetDenom(), netAssetValue.Price, netAssetValue.Volume, source)
	if err := ctx.EventManager().EmitTypedEvent(setNetAssetValueEvent); err != nil {
		return err
	}

	key := types.NetAssetValueKey(marker.GetAddress(), netAssetValue.Price.Denom)
	bz, err := k.cdc.Marshal(&netAssetValue)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(key, bz)

	return nil
}

// SetNetAssetValueWithBlockHeight adds/updates a net asset value to marker with a specific block height
func (k Keeper) SetNetAssetValueWithBlockHeight(ctx sdk.Context, marker types.MarkerAccountI, netAssetValue types.NetAssetValue, source string, blockHeight uint64) error {
	netAssetValue.UpdatedBlockHeight = blockHeight
	if err := netAssetValue.Validate(); err != nil {
		return err
	}

	setNetAssetValueEvent := types.NewEventSetNetAssetValue(marker.GetDenom(), netAssetValue.Price, netAssetValue.Volume, source)
	if err := ctx.EventManager().EmitTypedEvent(setNetAssetValueEvent); err != nil {
		return err
	}

	key := types.NetAssetValueKey(marker.GetAddress(), netAssetValue.Price.Denom)
	bz, err := k.cdc.Marshal(&netAssetValue)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(key, bz)

	return nil
}

// GetNetAssetValue gets the NetAssetValue for a marker denom with a specific price denom.
func (k Keeper) GetNetAssetValue(ctx sdk.Context, markerDenom, priceDenom string) (*types.NetAssetValue, error) {
	store := ctx.KVStore(k.storeKey)
	markerAddr, err := types.MarkerAddress(markerDenom)
	if err != nil {
		return nil, fmt.Errorf("could not get marker %q address: %w", markerDenom, err)
	}

	key := types.NetAssetValueKey(markerAddr, priceDenom)
	value := store.Get(key)
	if len(value) == 0 {
		return nil, nil
	}

	var markerNav types.NetAssetValue
	err = k.cdc.Unmarshal(value, &markerNav)
	if err != nil {
		return nil, fmt.Errorf("could not read nav for marker %q with price denom %q: %w", markerDenom, priceDenom, err)
	}

	return &markerNav, nil
}

// IterateNetAssetValues iterates net asset values for marker
func (k Keeper) IterateNetAssetValues(ctx sdk.Context, markerAddr sdk.AccAddress, handler func(state types.NetAssetValue) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := storetypes.KVStorePrefixIterator(store, types.NetAssetValueKeyPrefix(markerAddr))
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var markerNav types.NetAssetValue
		err := k.cdc.Unmarshal(it.Value(), &markerNav)
		if err != nil {
			return err
		} else if handler(markerNav) {
			break
		}
	}
	return nil
}

// IterateAllNetAssetValues iterates all net asset values
func (k Keeper) IterateAllNetAssetValues(ctx sdk.Context, handler func(sdk.AccAddress, types.NetAssetValue) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := storetypes.KVStorePrefixIterator(store, types.NetAssetValuePrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		markerAddr := types.GetMarkerFromNetAssetValueKey(it.Key())
		var markerNav types.NetAssetValue
		err := k.cdc.Unmarshal(it.Value(), &markerNav)
		if err != nil {
			return err
		} else if handler(markerAddr, markerNav) {
			break
		}
	}
	return nil
}

// RemoveNetAssetValues removes all net asset values for a marker
func (k Keeper) RemoveNetAssetValues(ctx sdk.Context, markerAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	it := storetypes.KVStorePrefixIterator(store, types.NetAssetValueKeyPrefix(markerAddr))
	var keys [][]byte
	for ; it.Valid(); it.Next() {
		keys = append(keys, it.Key())
	}
	it.Close()

	for _, key := range keys {
		store.Delete(key)
	}
}

// GetReqAttrBypassAddrs returns a deep copy of the addresses that bypass the required attributes checking.
func (k Keeper) GetReqAttrBypassAddrs() []sdk.AccAddress {
	return k.reqAttrBypassAddrs.GetSlice()
}

// IsReqAttrBypassAddr returns true if the provided addr can bypass the required attributes checking.
func (k Keeper) IsReqAttrBypassAddr(addr sdk.AccAddress) bool {
	return k.reqAttrBypassAddrs.Has(addr)
}
