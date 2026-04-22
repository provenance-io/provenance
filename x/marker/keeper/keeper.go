package keeper

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"

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
	feegrantKeeper feegrantkeeper.Keeper

	// To access attributes for addresses
	attrKeeper types.AttrKeeper
	// To access names and normalize required attributes
	nameKeeper types.NameKeeper

	// Key to access the key-value store from sdk.Context.
	storeService store.KVStoreService
	schema       collections.Schema
	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// markers stores marker address index: key = AccAddress, value = raw address bytes.
	// Key layout: [0x02][len(addr)][addr] → [addr_bytes]
	markers collections.Map[sdk.AccAddress, []byte]

	// denySend stores the send deny list: key = (markerAddr, denyAddr), value = sentinel.
	// Key layout: [0x03][len(marker)][marker][len(deny)][deny] → []byte{}
	denySend collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], bool]

	// navs stores net asset values: key = (markerAddr, priceDenom), value = NetAssetValue.
	// Key layout: [0x04][len(marker)][marker][denom] → proto(NetAssetValue)
	navs collections.Map[collections.Pair[sdk.AccAddress, string], types.NetAssetValue]

	// params stores module parameters as a singleton.
	// Key layout: [0x05] → proto(Params)
	params collections.Item[types.Params]

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

	// exchangeKeeper is an optional keeper for committing funds to exchange markets.
	exchangeKeeper types.ExchangeKeeper
}

// NewKeeper returns a marker keeper. It handles:
// - managing MarkerAccounts
// - enforcing permissions for marker creation/deletion/management
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	authzKeeper types.AuthzKeeper,
	feegrantKeeper feegrantkeeper.Keeper,
	attrKeeper types.AttrKeeper,
	nameKeeper types.NameKeeper,
	ibcTransferServer types.IbcTransferMsgServer,
	reqAttrBypassAddrs []sdk.AccAddress,
	checker types.GroupChecker,
) Keeper {
	addrCodec := types.MarkerAddrKeyCodec
	pairAddrCodec := collections.PairKeyCodec(addrCodec, addrCodec)
	navPairCodec := collections.PairKeyCodec(addrCodec, types.DenomStringKeyCodec)

	sb := collections.NewSchemaBuilder(storeService)
	rv := Keeper{
		authKeeper:            authKeeper,
		authzKeeper:           authzKeeper,
		bankKeeper:            bankKeeper,
		feegrantKeeper:        feegrantKeeper,
		attrKeeper:            attrKeeper,
		nameKeeper:            nameKeeper,
		storeService:          storeService,
		cdc:                   cdc,
		authority:             authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		markerModuleAddr:      authtypes.NewModuleAddress(types.CoinPoolName),
		ibcTransferModuleAddr: authtypes.NewModuleAddress(ibctypes.ModuleName),
		feeCollectorAddr:      authtypes.NewModuleAddress(authtypes.FeeCollectorName),
		ibcTransferServer:     ibcTransferServer,
		reqAttrBypassAddrs:    types.NewImmutableAccAddresses(reqAttrBypassAddrs),
		groupChecker:          checker,

		markers: collections.NewMap(
			sb,
			collections.NewPrefix(types.MarkerStoreKeyPrefix), // [0x02]
			"markers",
			addrCodec,
			types.RawBytesValue,
		),
		denySend: collections.NewMap(
			sb,
			collections.NewPrefix(types.DenySendKeyPrefix), // [0x03]
			"deny_send",
			pairAddrCodec,
			types.SentinelValue,
		),
		navs: collections.NewMap(
			sb,
			collections.NewPrefix(types.NetAssetValuePrefix), // [0x04]
			"net_asset_values",
			navPairCodec,
			codec.CollValue[types.NetAssetValue](cdc),
		),
		params: collections.NewItem(
			sb,
			collections.NewPrefix(types.MarkerParamStoreKey), // [0x05]
			"params",
			codec.CollValue[types.Params](cdc),
		),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(fmt.Errorf("marker: failed to build collections schema: %w", err))
	}
	rv.schema = schema
	bankKeeper.AppendSendRestriction(rv.SendRestrictionFn)
	return rv
}

// SetExchangeKeeper sets the exchange keeper and returns the updated Keeper.
// This must be called after both keepers are constructed to resolve the circular dependency
// (exchange depends on marker via MarkerKeeper interface, marker depends on exchange via ExchangeKeeper interface).
func (k *Keeper) SetExchangeKeeper(ek types.ExchangeKeeper) {
	k.exchangeKeeper = ek
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
	if err := marker.Validate(); err != nil {
		panic(err)
	}
	k.authKeeper.SetAccount(ctx, marker)
	if err := k.markers.Set(ctx, marker.GetAddress(), marker.GetAddress()); err != nil {
		panic(fmt.Errorf("failed to set marker index: %w", err))
	}
}

// RemoveMarker removes a marker from the auth account store. Note: if the account holds coins this will
// likely cause an invariant constraint violation for the coin supply
func (k Keeper) RemoveMarker(ctx sdk.Context, marker types.MarkerAccountI) {
	k.authKeeper.RemoveAccount(ctx, marker)
	k.RemoveNetAssetValues(ctx, marker.GetAddress())
	k.ClearSendDeny(ctx, marker.GetAddress())
	if err := k.markers.Remove(ctx, marker.GetAddress()); err != nil {
		panic(fmt.Errorf("failed to remove marker index: %w", err))
	}
}

// IterateMarkers iterates all markers with the given handler function.
func (k Keeper) IterateMarkers(ctx sdk.Context, cb func(marker types.MarkerAccountI) (stop bool)) {
	err := k.markers.Walk(ctx, nil, func(key sdk.AccAddress, value []byte) (bool, error) {
		account := k.authKeeper.GetAccount(ctx, sdk.AccAddress(value))
		ma, ok := account.(types.MarkerAccountI)
		if !ok {
			return true, fmt.Errorf("invalid account type in marker account registry")
		}
		return cb(ma), nil
	})
	if err != nil {
		panic(err)
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
	has, err := k.denySend.Has(ctx, collections.Join(markerAddr, senderAddr))
	if err != nil {
		return false
	}
	return has
}

// AddSendDeny set sender address to denied for marker
func (k Keeper) AddSendDeny(ctx sdk.Context, markerAddr, senderAddr sdk.AccAddress) {
	if err := k.denySend.Set(ctx, collections.Join(markerAddr, senderAddr), true); err != nil {
		panic(fmt.Errorf("failed to add send deny: %w", err))
	}
}

// RemoveSendDeny removes sender address from marker deny list
func (k Keeper) RemoveSendDeny(ctx sdk.Context, markerAddr, senderAddr sdk.AccAddress) {
	if err := k.denySend.Remove(ctx, collections.Join(markerAddr, senderAddr)); err != nil {
		panic(fmt.Errorf("failed to remove send deny: %w", err))
	}
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
	err := k.denySend.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, sdk.AccAddress], _ bool) (bool, error) {
		fullKey := types.DenySendKey(key.K1(), key.K2())
		return handler(fullKey), nil
	})
	if err != nil {
		panic(err)
	}
}

// GetSendDenyList gets the list of sender addresses from the marker's deny list
func (k Keeper) GetSendDenyList(ctx sdk.Context, markerAddr sdk.AccAddress) []sdk.AccAddress {
	list := []sdk.AccAddress{}
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.AccAddress](markerAddr)
	err := k.denySend.Walk(ctx, rng, func(key collections.Pair[sdk.AccAddress, sdk.AccAddress], _ bool) (bool, error) {
		list = append(list, key.K2())
		return false, nil
	})
	if err != nil {
		panic(err)
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
	netAssetValue.UpdatedBlockHeight = uint64(ctx.BlockHeight()) //nolint:gosec // G115: ctx.BlockHeight() is always non-negative.
	if err := netAssetValue.Validate(); err != nil {
		return err
	}

	setNetAssetValueEvent := types.NewEventSetNetAssetValue(marker.GetDenom(), netAssetValue.Price, netAssetValue.Volume, source)
	if err := ctx.EventManager().EmitTypedEvent(setNetAssetValueEvent); err != nil {
		return err
	}

	return k.navs.Set(ctx, collections.Join(marker.GetAddress(), netAssetValue.Price.Denom), netAssetValue)
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

	return k.navs.Set(ctx, collections.Join(marker.GetAddress(), netAssetValue.Price.Denom), netAssetValue)
}

// GetNetAssetValue gets the NetAssetValue for a marker denom with a specific price denom.
func (k Keeper) GetNetAssetValue(ctx sdk.Context, markerDenom, priceDenom string) (*types.NetAssetValue, error) {
	markerAddr, err := types.MarkerAddress(markerDenom)
	if err != nil {
		return nil, fmt.Errorf("could not get marker %q address: %w", markerDenom, err)
	}

	markerNav, err := k.navs.Get(ctx, collections.Join(markerAddr, priceDenom))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not read nav for marker %q with price denom %q: %w", markerDenom, priceDenom, err)
	}

	return &markerNav, nil
}

// IterateNetAssetValues iterates net asset values for marker
func (k Keeper) IterateNetAssetValues(ctx sdk.Context, markerAddr sdk.AccAddress, handler func(state types.NetAssetValue) (stop bool)) error {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, string](markerAddr)
	return k.navs.Walk(ctx, rng, func(_ collections.Pair[sdk.AccAddress, string], nav types.NetAssetValue) (bool, error) {
		return handler(nav), nil
	})
}

// IterateAllNetAssetValues iterates all net asset values
func (k Keeper) IterateAllNetAssetValues(ctx sdk.Context, handler func(sdk.AccAddress, types.NetAssetValue) (stop bool)) error {
	return k.navs.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, string], nav types.NetAssetValue) (bool, error) {
		return handler(key.K1(), nav), nil
	})
}

// RemoveNetAssetValues removes all net asset values for a marker
func (k Keeper) RemoveNetAssetValues(ctx sdk.Context, markerAddr sdk.AccAddress) {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, string](markerAddr)
	var keys []collections.Pair[sdk.AccAddress, string]
	err := k.navs.Walk(ctx, rng, func(key collections.Pair[sdk.AccAddress, string], _ types.NetAssetValue) (bool, error) {
		keys = append(keys, key)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	for _, key := range keys {
		if err := k.navs.Remove(ctx, key); err != nil {
			panic(err)
		}
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

// IsMarkerAccount returns true if the provided address is one for a marker account.
func (k Keeper) IsMarkerAccount(ctx sdk.Context, addr sdk.AccAddress) bool {
	if len(addr) == 0 {
		return false
	}
	has, err := k.markers.Has(ctx, addr)
	if err != nil {
		return false
	}
	return has
}
