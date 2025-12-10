package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/x/sanction"
	"github.com/provenance-io/provenance/x/sanction/errors"
)

type Keeper struct {
	cdc codec.BinaryCodec

	govKeeper sanction.GovKeeper

	authority string

	unsanctionableAddrs map[string]bool

	msgSanctionTypeURL          string
	msgUnsanctionTypeURL        string
	msgExecLegacyContentTypeURL string

	StoreService store.KVStoreService

	// Collections Schema
	Schema collections.Schema

	// Params collection: name -> value (as []byte)
	ParameterStore collections.Map[string, []byte]

	// Sanctioned addresses: address -> []byte{0x01}
	SanctionedAddressesStore collections.Map[sdk.AccAddress, []byte]

	// Temporary entries with backward-compatible key encoding
	// Key: (address, govPropID) -> Value: []byte{0x01 or 0x00}
	TemporaryEntriesStore collections.Map[collections.Pair[sdk.AccAddress, uint64], []byte]

	// Proposal index with backward-compatible key encoding
	// Key: (govPropID, address) -> Value: []byte{0x01 or 0x00}
	ProposalIndex collections.Map[collections.Pair[uint64, sdk.AccAddress], []byte]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	bankKeeper sanction.BankKeeper,
	govKeeper *govkeeper.Keeper,
	authority string,
	unsanctionableAddrs []sdk.AccAddress,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:                         cdc,
		authority:                   authority,
		govKeeper:                   WrapGovKeeper(govKeeper),
		unsanctionableAddrs:         make(map[string]bool),
		msgSanctionTypeURL:          sdk.MsgTypeURL(&sanction.MsgSanction{}),
		msgUnsanctionTypeURL:        sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
		msgExecLegacyContentTypeURL: sdk.MsgTypeURL(&govv1.MsgExecLegacyContent{}),

		// Params: simple string -> []byte map
		ParameterStore: collections.NewMap(
			sb,
			ParamsPrefix,
			"params",
			collections.StringKey,
			collections.BytesValue,
		),

		// Sanctioned addresses: use custom AccAddress codec
		SanctionedAddressesStore: collections.NewMap(
			sb,
			SanctionedPrefix,
			"sanctioned_addresses",
			sanction.AccAddressKey{},
			collections.BytesValue,
		),

		// Temporary entries: (AccAddress, uint64) -> []byte with custom codec
		// Collections adds 0x02 prefix, codec handles [len][addr][propID]
		TemporaryEntriesStore: collections.NewMap(
			sb,
			TemporaryPrefix,
			"temporary_entries",
			sanction.TemporaryKeyCodec{},
			collections.BytesValue,
		),

		// Proposa
		// l index: (uint64, AccAddress) -> []byte with custom codec
		// Collections adds 0x03 prefix, codec handles [propID][len][addr]
		ProposalIndex: collections.NewMap(
			sb,
			ProposalIndexPrefix,
			"proposal_index",
			sanction.ProposalIndexKeyCodec{},
			collections.BytesValue,
		),
		StoreService: storeService,
	}

	for _, addr := range unsanctionableAddrs {
		k.unsanctionableAddrs[string(addr)] = true
	}

	schema, err := sb.Build()
	if err != nil {
		panic(fmt.Errorf("building sanction keeper schema: %w", err))
	}
	k.Schema = schema

	bankKeeper.AppendSendRestriction(k.SendRestrictionFn)
	return k
}

// GetAuthority returns this module's authority string.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// IsSanctionedAddr returns true if the provided address is currently sanctioned (either permanently or temporarily).
func (k Keeper) IsSanctionedAddr(goCtx context.Context, addr sdk.AccAddress) bool {
	if len(addr) == 0 || k.IsAddrThatCannotBeSanctioned(addr) {
		return false
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	tempEntry := k.getLatestTempEntry(ctx, addr)
	if IsSanctionBz(tempEntry) {
		return true
	}
	if IsUnsanctionBz(tempEntry) {
		return false
	}
	// Check permanent sanction
	has, err := k.SanctionedAddressesStore.Has(ctx, addr)
	if err != nil {
		return false
	}
	return has
}

// SanctionAddresses creates permanent sanctioned address entries for each of the provided addresses.
// Also deletes any temporary entries for each address.
func (k Keeper) SanctionAddresses(ctx sdk.Context, addrs ...sdk.AccAddress) error {
	val := []byte{SanctionB}
	for _, addr := range addrs {
		if k.IsAddrThatCannotBeSanctioned(addr) {
			return errors.ErrUnsanctionableAddr.Wrap(addr.String())
		}

		if err := k.SanctionedAddressesStore.Set(ctx, addr, val); err != nil {
			return fmt.Errorf("setting sanctioned address %s: %w", addr.String(), err)
		}

		if err := ctx.EventManager().EmitTypedEvent(sanction.NewEventAddressSanctioned(addr)); err != nil {
			return fmt.Errorf("emitting sanction event for %s: %w", addr.String(), err)
		}
	}

	if err := k.DeleteAddrTempEntries(ctx, addrs...); err != nil {
		return fmt.Errorf("deleting temp entries: %w", err)
	}

	return nil
}

// UnsanctionAddresses deletes any sanctioned address entries for each provided address.
// Also deletes any temporary entries for each address.
func (k Keeper) UnsanctionAddresses(ctx sdk.Context, addrs ...sdk.AccAddress) error {
	for _, addr := range addrs {
		if err := k.SanctionedAddressesStore.Remove(ctx, addr); err != nil {
			return fmt.Errorf("removing sanctioned address %s: %w", addr.String(), err)
		}

		if err := ctx.EventManager().EmitTypedEvent(sanction.NewEventAddressUnsanctioned(addr)); err != nil {
			return fmt.Errorf("emitting unsanction event for %s: %w", addr.String(), err)
		}
	}

	if err := k.DeleteAddrTempEntries(ctx, addrs...); err != nil {
		return fmt.Errorf("deleting temp entries: %w", err)
	}

	return nil
}

// AddTemporarySanction adds a temporary sanction with the given gov prop id for each of the provided addresses.
func (k Keeper) AddTemporarySanction(ctx sdk.Context, govPropID uint64, addrs ...sdk.AccAddress) error {
	return k.addTempEntries(ctx, SanctionB, govPropID, addrs)
}

// AddTemporaryUnsanction adds a temporary unsanction with the given gov prop id for each of the provided addresses.
func (k Keeper) AddTemporaryUnsanction(ctx sdk.Context, govPropID uint64, addrs ...sdk.AccAddress) error {
	return k.addTempEntries(ctx, UnsanctionB, govPropID, addrs)
}

// addTempEntries adds a temporary entry with the given value and gov prop id for each address given.
func (k Keeper) addTempEntries(ctx sdk.Context, value byte, govPropID uint64, addrs []sdk.AccAddress) error {
	val := []byte{value}
	for _, addr := range addrs {
		if value == SanctionB && k.IsAddrThatCannotBeSanctioned(addr) {
			return errors.ErrUnsanctionableAddr.Wrap(addr.String())
		}

		key := collections.Join(addr, govPropID)
		if err := k.TemporaryEntriesStore.Set(ctx, key, val); err != nil {
			return fmt.Errorf("setting temporary entry for %s/%d: %w", addr.String(), govPropID, err)
		}

		indexKey := collections.Join(govPropID, addr)
		if err := k.ProposalIndex.Set(ctx, indexKey, val); err != nil {
			return fmt.Errorf("setting proposal index for %d/%s: %w", govPropID, addr.String(), err)
		}
		if err := ctx.EventManager().EmitTypedEvent(NewTempEvent(value, addr)); err != nil {
			return err
		}
	}
	return nil
}

// getLatestTempEntry gets the most recent temporary entry for the given address.
func (k Keeper) getLatestTempEntry(ctx sdk.Context, addr sdk.AccAddress) []byte {
	if len(addr) == 0 {
		return nil
	}

	var latestValue []byte
	var latestPropID uint64
	found := false

	// Walk all temporary entries and find the one for this address with highest propID
	err := k.TemporaryEntriesStore.Walk(ctx, nil,
		func(key collections.Pair[sdk.AccAddress, uint64], value []byte) (stop bool, err error) {
			// Only consider entries for this address
			if !key.K1().Equals(addr) {
				return false, nil
			}

			propID := key.K2()
			if !found || propID > latestPropID {
				latestPropID = propID
				latestValue = value
				found = true
			}
			return false, nil
		})

	if err != nil {
		return nil
	}

	return latestValue
}

// DeleteGovPropTempEntries deletes the temporary entries for the given proposal id.
func (k Keeper) DeleteGovPropTempEntries(ctx sdk.Context, govPropID uint64) error {
	var toRemove []collections.Pair[sdk.AccAddress, uint64]
	var indexToRemove []collections.Pair[uint64, sdk.AccAddress]

	err := k.ProposalIndex.Walk(ctx, nil,
		func(key collections.Pair[uint64, sdk.AccAddress], value []byte) (stop bool, err error) {
			if key.K1() == govPropID {
				addr := key.K2()
				toRemove = append(toRemove, collections.Join(addr, govPropID))
				indexToRemove = append(indexToRemove, key)
			}
			return false, nil
		})

	if err != nil {
		return fmt.Errorf("walking proposal index for %d: %w", govPropID, err)
	}

	for _, key := range toRemove {
		if err := k.TemporaryEntriesStore.Remove(ctx, key); err != nil {
			return fmt.Errorf("removing temporary entry %v: %w", key, err)
		}
	}

	for _, key := range indexToRemove {
		if err := k.ProposalIndex.Remove(ctx, key); err != nil {
			return fmt.Errorf("removing proposal index entry %v: %w", key, err)
		}
	}

	return nil
}

// DeleteAddrTempEntries deletes all temporary entries for each given address.
func (k Keeper) DeleteAddrTempEntries(ctx sdk.Context, addrs ...sdk.AccAddress) error {
	if len(addrs) == 0 {
		return nil
	}
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}

		var tempKeys []collections.Pair[sdk.AccAddress, uint64]
		var indexKeys []collections.Pair[uint64, sdk.AccAddress]

		// Collect all temporary entries for this address
		err := k.TemporaryEntriesStore.Walk(ctx, nil,
			func(key collections.Pair[sdk.AccAddress, uint64], value []byte) (stop bool, err error) {
				if key.K1().Equals(addr) {
					tempKeys = append(tempKeys, key)
					indexKeys = append(indexKeys, collections.Join(key.K2(), addr))
				}
				return false, nil
			})

		if err != nil {
			return fmt.Errorf("walking temp entries for %s: %w", addr.String(), err)
		}

		// Delete temporary entries
		for _, key := range tempKeys {
			if err := k.TemporaryEntriesStore.Remove(ctx, key); err != nil {
				return fmt.Errorf("removing temp entry %v: %w", key, err)
			}
		}

		// Delete index entries
		for _, key := range indexKeys {
			if err := k.ProposalIndex.Remove(ctx, key); err != nil {
				return fmt.Errorf("removing index entry %v: %w", key, err)
			}
		}
	}

	return nil
}

// getSanctionedAddressPrefixStore returns a kv store prefixed for sanctioned addresses, and the prefix bytes.
// func (k Keeper) getSanctionedAddressPrefixStore(ctx sdk.Context) storetypes.KVStore {
// 	return prefix.NewStore(ctx.KVStore(k.storeKey), SanctionedPrefix)
// }

// IterateSanctionedAddresses iterates over all of the permanently sanctioned addresses.
// The callback takes in the sanctioned address and should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateSanctionedAddresses(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool)) error {
	return k.SanctionedAddressesStore.Walk(ctx, nil, func(addr sdk.AccAddress, value []byte) (stop bool, err error) {
		return cb(addr), nil
	})
}

// getTemporaryEntryPrefixStore returns a kv store prefixed for temporary sanction/unsanction entries, and the prefix bytes used.
// If an addr is provided, the store is prefixed for just the given address.
// If addr is empty, it will be prefixed for all temporary entries.
// func (k Keeper) getTemporaryEntryPrefixStore(ctx sdk.Context, addr sdk.AccAddress) (storetypes.KVStore, []byte) {
// 	pre := CreateTemporaryAddrPrefix(addr)
// 	return prefix.NewStore(ctx.KVStore(k.storeKey), pre), pre
// }

// IterateTemporaryEntries iterates over each of the temporary entries.
// If an address is provided, only the temporary entries for that address are iterated,
// otherwise all entries are iterated.
// The callback takes in the address in question, the governance proposal associated with it, and whether it's a sanction (true) or unsanction (false).
// The callback should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateTemporaryEntries(ctx sdk.Context, addr sdk.AccAddress, cb func(addr sdk.AccAddress, govPropID uint64, isSanction bool) (stop bool)) error {
	return k.TemporaryEntriesStore.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, uint64], value []byte) (stop bool, err error) {
		if len(addr) > 0 && !key.K1().Equals(addr) {
			return false, nil
		}

		isSanction := IsSanctionBz(value)
		return cb(key.K1(), key.K2(), isSanction), nil
	})
}

// getProposalIndexPrefixStore returns a kv store prefixed for the gov prop -> temporary sanction/unsanction index entries,
// and the prefix bytes used.
// If a gov prop id is provided, the store is prefixed for just that proposal.
// If not provided, it will be prefixed for all temp index entries.
// func (k Keeper) getProposalIndexPrefixStore(ctx sdk.Context, govPropID *uint64) (storetypes.KVStore, []byte) {
// 	pre := CreateProposalTempIndexPrefix(govPropID)
// 	return prefix.NewStore(ctx.KVStore(k.storeKey), pre), pre
// }

// IterateProposalIndexEntries iterates over all of the index entries for temp entries.
// The callback takes in the gov prop id and address.
// The callback should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateProposalIndexEntries(ctx sdk.Context, govPropID *uint64, cb func(govPropID uint64, addr sdk.AccAddress) (stop bool)) error {
	return k.ProposalIndex.Walk(ctx, nil, func(key collections.Pair[uint64, sdk.AccAddress], value []byte) (stop bool, err error) {
		if govPropID != nil && key.K1() != *govPropID {
			return false, nil
		}

		return cb(key.K1(), key.K2()), nil
	})
}

// IsAddrThatCannotBeSanctioned returns true if the provided address is one of the ones that cannot be sanctioned.
// Returns false if the addr can be sanctioned.
func (k Keeper) IsAddrThatCannotBeSanctioned(addr sdk.AccAddress) bool {
	// Okay. I know this is a clunky name for this function.
	// IsUnsanctionableAddr would be a better name if it weren't WAY too close to IsSanctionedAddr.
	// The latter is the key function of this module, and I wanted to help prevent
	// confusion between this one and that one since they have vastly different purposes.
	return k.unsanctionableAddrs[string(addr)]
}

// GetParams gets the sanction module's params.
// If there isn't anything set in state, the defaults are returned.
func (k Keeper) GetParams(ctx sdk.Context) *sanction.Params {
	rv := sanction.DefaultParams()

	// Get immediate sanction min deposit
	sanctionDeposit, err := k.ParameterStore.Get(ctx, ParamNameImmediateSanctionMinDeposit)
	if err == nil {
		coins, parseErr := sdk.ParseCoinsNormalized(string(sanctionDeposit))
		if parseErr == nil {
			rv.ImmediateSanctionMinDeposit = coins
		}
	}

	// Get immediate unsanction min deposit
	unsanctionDeposit, err := k.ParameterStore.Get(ctx, ParamNameImmediateUnsanctionMinDeposit)
	if err == nil {
		coins, parseErr := sdk.ParseCoinsNormalized(string(unsanctionDeposit))
		if parseErr == nil {
			rv.ImmediateUnsanctionMinDeposit = coins
		}
	}

	return rv
}

// SetParams sets the sanction module's params.
// Providing a nil params will cause all params to be deleted (so that defaults are used).
func (k Keeper) SetParams(ctx sdk.Context, params *sanction.Params) error {
	if params == nil {
		params = &sanction.Params{
			ImmediateSanctionMinDeposit:   sanction.DefaultImmediateSanctionMinDeposit,
			ImmediateUnsanctionMinDeposit: sanction.DefaultImmediateUnsanctionMinDeposit,
		}
	}

	if err := k.ParameterStore.Set(ctx, ParamNameImmediateSanctionMinDeposit,
		[]byte(params.ImmediateSanctionMinDeposit.String())); err != nil {
		return fmt.Errorf("setting immediate sanction min deposit: %w", err)
	}

	if err := k.ParameterStore.Set(ctx, ParamNameImmediateUnsanctionMinDeposit,
		[]byte(params.ImmediateUnsanctionMinDeposit.String())); err != nil {
		return fmt.Errorf("setting immediate unsanction min deposit: %w", err)
	}

	return ctx.EventManager().EmitTypedEvent(&sanction.EventParamsUpdated{})
}

// IterateParams iterates over all params entries.
// The callback takes in the name and value, and should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateParams(ctx sdk.Context, cb func(name, value string) (stop bool)) error {
	return k.ParameterStore.Walk(ctx, nil, func(name string, value []byte) (stop bool, err error) {
		return cb(name, string(value)), nil
	})
}

// GetImmediateSanctionMinDeposit gets the minimum deposit for a sanction to happen immediately.
func (k Keeper) GetImmediateSanctionMinDeposit(ctx sdk.Context) sdk.Coins {
	params := k.GetParams(ctx)
	return params.ImmediateSanctionMinDeposit
}

// GetImmediateUnsanctionMinDeposit gets the minimum deposit for an unsanction to happen immediately.
func (k Keeper) GetImmediateUnsanctionMinDeposit(ctx sdk.Context) sdk.Coins {
	params := k.GetParams(ctx)
	return params.ImmediateUnsanctionMinDeposit
}

// getParam returns a param value and whether it existed.
func (k Keeper) getParam(ctx sdk.Context, name string) (string, bool) {
	key, err := k.ParameterStore.Get(ctx, name)
	if err != nil {
		return "", false
	}

	return string(key), true
}

// setParam sets a param value.
func (k Keeper) setParam(ctx sdk.Context, name, value string) {
	k.ParameterStore.Set(ctx, name, []byte(value))
}

// deleteParam deletes a param value.
func (k Keeper) deleteParam(ctx sdk.Context, name string) {
	k.ParameterStore.Remove(ctx, name)
}

// getParamAsCoinsOrDefault gets a param value and converts it to a coins if possible.
// If the param doesn't exist, the default is returned.
// If the param's value cannot be converted to a Coins, the default is returned.
func (k Keeper) getParamAsCoinsOrDefault(ctx sdk.Context, name string, dflt sdk.Coins) sdk.Coins {
	coins, has := k.getParam(ctx, name)
	if !has {
		return dflt
	}
	return toCoinsOrDefault(string(coins), dflt)
}

// toCoinsOrDefault converts a string to coins if possible or else returns the provided default.
func toCoinsOrDefault(coins string, dflt sdk.Coins) sdk.Coins {
	rv, err := sdk.ParseCoinsNormalized(coins)
	if err != nil {
		return dflt
	}
	return rv
}

// toAccAddrs converts the provided strings into a slice of sdk.AccAddress.
// If any fail to convert, an error is returned.
func toAccAddrs(addrs []string) ([]sdk.AccAddress, error) {
	var err error
	rv := make([]sdk.AccAddress, len(addrs))
	for i, addr := range addrs {
		rv[i], err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid address[%d]: %w", i, err)
		}
	}
	return rv, nil
}
