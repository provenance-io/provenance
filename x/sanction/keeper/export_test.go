package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/sanction"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

var (
	// OnlyTestsConcatBzPlusCap, for unit tests, exposes the concatBzPlusCap function.
	OnlyTestsConcatBzPlusCap = concatBzPlusCap

	// OnlyTestsToCoinsOrDefault, for unit tests, exposes the toCoinsOrDefault function.
	OnlyTestsToCoinsOrDefault = toCoinsOrDefault

	// OnlyTestsToAccAddrs, for unit tests, exposes the toAccAddrs function.
	OnlyTestsToAccAddrs = toAccAddrs
)

// WithGovKeeper, for unit tests, creates a copy of this, setting the govKeeper to the provided one.
func (k Keeper) WithGovKeeper(govKeeper sanction.GovKeeper) Keeper {
	k.govKeeper = govKeeper
	return k
}

// WithAuthority, for unit tests, creates a copy of this, setting the authority to the provided one.
func (k Keeper) WithAuthority(authority string) Keeper {
	k.authority = authority
	return k
}

// WithUnsanctionableAddrs, for unit tests, creates a copy of this, setting the unsanctionableAddrs to the provided one.
// This does not add the provided ones to the unsanctionableAddrs, it overwrites the
// existing ones with the ones provided.
func (k Keeper) WithUnsanctionableAddrs(unsanctionableAddrs map[string]bool) Keeper {
	k.unsanctionableAddrs = unsanctionableAddrs
	return k
}

// StoreKey, for unit tests, exposes this keeper's storekey.
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.storeKey
}

// MsgSanctionTypeURL, for unit tests, exposes this keeper's msgSanctionTypeURL.
func (k Keeper) MsgSanctionTypeURL() string {
	return k.msgSanctionTypeURL
}

// MsgUnsanctionTypeURL, for unit tests, exposes this keeper's msgUnsanctionTypeURL.
func (k Keeper) MsgUnsanctionTypeURL() string {
	return k.msgUnsanctionTypeURL
}

// MsgExecLegacyContentTypeURL, for unit tests, exposes this keeper's msgExecLegacyContentTypeURL.
func (k Keeper) MsgExecLegacyContentTypeURL() string {
	return k.msgExecLegacyContentTypeURL
}

// GetParamAsCoinsOrDefault, for unit tests, exposes this keeper's getParamAsCoinsOrDefault function.
func (k Keeper) GetParamAsCoinsOrDefault(ctx sdk.Context, name string, dflt sdk.Coins) sdk.Coins {
	return k.getParamAsCoinsOrDefault(ctx, name, dflt)
}

// GetLatestTempEntry, for unit tests, exposes this keeper's getLatestTempEntry function.
func (k Keeper) GetLatestTempEntry(store storetypes.KVStore, addr sdk.AccAddress) []byte {
	return k.getLatestTempEntry(store, addr)
}

// GetParam, for unit tests, exposes this keeper's getParam function.
func (k Keeper) GetParam(store storetypes.KVStore, name string) (string, bool) {
	return k.getParam(store, name)
}

// SetParam, for unit tests, exposes this keeper's setParam function.
func (k Keeper) SetParam(store storetypes.KVStore, name, value string) {
	k.setParam(store, name, value)
}

// DeleteParam, for unit tests, exposes this keeper's deleteParam function.
func (k Keeper) DeleteParam(store storetypes.KVStore, name string) {
	k.deleteParam(store, name)
}

// ProposalGovHook, for unit tests, exposes this keeper's proposalGovHook function.
func (k Keeper) ProposalGovHook(ctx context.Context, proposalID uint64) error {
	return k.proposalGovHook(ctx, proposalID)
}

// IsModuleGovHooksMsgURL, for unit tests, exposes this keeper's isModuleGovHooksMsgURL function.
func (k Keeper) IsModuleGovHooksMsgURL(url string) bool {
	return k.isModuleGovHooksMsgURL(url)
}

// GetMsgAddresses, for unit tests, exposes this keeper's getMsgAddresses function.
func (k Keeper) GetMsgAddresses(msg *codectypes.Any) []sdk.AccAddress {
	return k.getMsgAddresses(msg)
}

// ImmediateMinDeposit, for unit tests, exposes this keeper's getImmediateMinDeposit function.
func (k Keeper) ImmediateMinDeposit(ctx sdk.Context, msg *codectypes.Any) sdk.Coins {
	return k.getImmediateMinDeposit(ctx, msg)
}
