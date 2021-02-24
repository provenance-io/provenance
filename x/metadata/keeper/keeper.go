package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/provenance-io/provenance/x/metadata/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// MetadataKeeperI is the internal state api for the metadata module.
type MetadataKeeperI interface {
	// GetScope returns the scope with the given address.
	GetScope(sdk.Context, types.MetadataAddress) (types.Scope, bool)
	// SetScope persists the provided scope
	SetScope(sdk.Context, types.Scope)
	// RemoveScope persists the provided scope
	RemoveScope(sdk.Context, types.MetadataAddress)

	// IterateScopes processes all stored scopes with the given handler.
	IterateScopes(sdk.Context, func(types.Scope) bool) error

	// GetRecordGroup returns the scope with the given address.
	GetRecordGroup(sdk.Context, types.MetadataAddress) (types.RecordGroup, bool)
	// SetRecordGroup persists the provided scope
	SetRecordGroup(sdk.Context, types.RecordGroup)
	// RemoveRecordGroup persists the provided scope
	RemoveRecordGroup(sdk.Context, types.MetadataAddress)

	// GetRecord returns the record with the given address.
	GetRecord(sdk.Context, types.MetadataAddress) (types.Record, bool)
	// SetRecord persists the provided record
	SetRecord(sdk.Context, types.Record)
	// RemoveRecord persists the provided scope
	RemoveRecord(sdk.Context, types.MetadataAddress)

	// IterateRecords processes all stored record for a scope with the given handler.
	IterateRecords(sdk.Context, types.MetadataAddress, func(types.Record) bool) error

	// GetGroupSpecification returns the record with the given address.
	GetGroupSpecification(sdk.Context, types.MetadataAddress) (types.GroupSpecification, bool)
	// SetGroupSpecification persists the provided group specification
	SetGroupSpecification(sdk.Context, types.GroupSpecification)

	// GetScopeSpecification returns the record with the given address.
	GetScopeSpecification(sdk.Context, types.MetadataAddress) (types.ScopeSpecification, bool)
	// SetScopeSpecification persists the provided scope specification
	SetScopeSpecification(sdk.Context, types.ScopeSpecification)
	// DeleteScopeSpecification deletes a scope specification from the module kv store.
	DeleteScopeSpecification(ctx sdk.Context, id types.MetadataAddress)

	// IterateScopeSpecs processes all scope specs using a given handler.
	IterateScopeSpecs(ctx sdk.Context, handler func(specification types.ScopeSpecification) (stop bool)) error
	// IterateScopeSpecsForAddress processes all scope specs associated with an address using a given handler.
	IterateScopeSpecsForAddress(ctx sdk.Context, address sdk.AccAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error
	// IterateScopeSpecsForContractSpec processes all scope specs associated with a contract spec id using a given handler.
	IterateScopeSpecsForContractSpec(ctx sdk.Context, contractSpecID types.MetadataAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error
}

// Keeper is the concrete state-based API for the metadata module.
type Keeper struct {
	// Key to access the key-value store from sdk.Context
	storeKey   sdk.StoreKey
	cdc        codec.BinaryMarshaler
	paramSpace paramtypes.Subspace

	// To check if accounts exist and set public keys.
	authKeeper authkeeper.AccountKeeper
}

// NewKeeper creates new instances of the metadata Keeper.
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	authKeeper authkeeper.AccountKeeper,
) Keeper {
	return Keeper{
		storeKey:   key,
		cdc:        cdc,
		paramSpace: paramSpace,
		authKeeper: authKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

var _ MetadataKeeperI = &Keeper{}

// GetAccount looks up an account by address
func (k Keeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return k.authKeeper.GetAccount(ctx, addr)
}

// CreateAccountForKey creates a new account for the given address with a public key set if given.
func (k Keeper) CreateAccountForKey(ctx sdk.Context, addr sdk.AccAddress, pubKey cryptotypes.PubKey) error {
	account := k.authKeeper.NewAccountWithAddress(ctx, addr)
	if pubKey != nil {
		if err := account.SetPubKey(pubKey); err != nil {
			return err
		}
	}
	k.authKeeper.SetAccount(ctx, account)
	return nil
}
