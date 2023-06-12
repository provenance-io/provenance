package keeper

import (
	"net/url"

	"github.com/gogo/protobuf/proto"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// MetadataKeeperI is the internal state api for the metadata module.
type MetadataKeeperI interface {
	// GetScope returns the scope with the given id.
	GetScope(sdk.Context, types.MetadataAddress) (types.Scope, bool)
	// SetScope stores a scope in the module kv store.
	SetScope(sdk.Context, types.Scope)
	// RemoveScope removes a scope from the module kv store along with all its records and sessions.
	RemoveScope(sdk.Context, types.MetadataAddress)

	// IterateScopes processes all stored scopes with the given handler.
	IterateScopes(sdk.Context, func(types.Scope) bool) error
	// IterateScopesForAddress processes scopes associated with the provided address with the given handler.
	IterateScopesForAddress(sdk.Context, sdk.AccAddress, func(types.MetadataAddress) bool) error
	// IterateScopesForScopeSpec processes scopes associated with the provided scope specification id with the given handler.
	IterateScopesForScopeSpec(sdk.Context, types.MetadataAddress, func(types.MetadataAddress) bool) error

	// GetSession returns the session with the given id.
	GetSession(sdk.Context, types.MetadataAddress) (types.Session, bool)
	// SetSession stores a session in the module kv store.
	SetSession(sdk.Context, types.Session)
	// RemoveSession removes a session from the module kv store if there are no records associated with it.
	RemoveSession(sdk.Context, types.MetadataAddress)

	// IterateSessions processes stored sessions with the given handler.
	IterateSessions(sdk.Context, types.MetadataAddress, func(types.Session) bool) error

	// GetRecord returns the record with the given id.
	GetRecord(sdk.Context, types.MetadataAddress) (types.Record, bool)
	// GetRecords returns records for a scope optionally limited to a name.
	GetRecords(sdk.Context, types.MetadataAddress, string) ([]*types.Record, error)
	// SetRecord stores a record in the module kv store.
	SetRecord(sdk.Context, types.Record)
	// RemoveRecord removes a record from the module kv store.
	RemoveRecord(sdk.Context, types.MetadataAddress)

	// IterateRecords processes stored records with the given handler.
	IterateRecords(sdk.Context, types.MetadataAddress, func(types.Record) bool) error

	// GetScopeSpecification returns the scope specification with the given id.
	GetScopeSpecification(sdk.Context, types.MetadataAddress) (types.ScopeSpecification, bool)
	// SetScopeSpecification stores a scope specification in the module kv store.
	SetScopeSpecification(sdk.Context, types.ScopeSpecification)
	// RemoveScopeSpecification removes a scope specification from the module kv store.
	RemoveScopeSpecification(sdk.Context, types.MetadataAddress) error

	// IterateScopeSpecs processes all scope specs using a given handler.
	IterateScopeSpecs(ctx sdk.Context, handler func(specification types.ScopeSpecification) (stop bool)) error
	// IterateScopeSpecsForOwner processes all scope specs owned by an address using a given handler.
	IterateScopeSpecsForOwner(ctx sdk.Context, ownerAddress sdk.AccAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error
	// IterateScopeSpecsForContractSpec processes all scope specs associated with a contract spec id using a given handler.
	IterateScopeSpecsForContractSpec(ctx sdk.Context, contractSpecID types.MetadataAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error

	// GetContractSpecification returns the contract specification with the given id.
	GetContractSpecification(sdk.Context, types.MetadataAddress) (types.ContractSpecification, bool)
	// SetContractSpecification stores a contract specification in the module kv store.
	SetContractSpecification(sdk.Context, types.ContractSpecification)
	// RemoveContractSpecification removes a contract specification from the module kv store.
	RemoveContractSpecification(sdk.Context, types.MetadataAddress) error

	// IterateContractSpecs processes all contract specs using a given handler.
	IterateContractSpecs(ctx sdk.Context, handler func(specification types.ContractSpecification) (stop bool)) error
	// IterateContractSpecsForOwner processes all contract specs owned by an address using a given handler.
	IterateContractSpecsForOwner(ctx sdk.Context, ownerAddress sdk.AccAddress, handler func(contractSpecID types.MetadataAddress) (stop bool)) error

	// GetRecordSpecification returns the record specification with the given id.
	GetRecordSpecification(sdk.Context, types.MetadataAddress) (types.RecordSpecification, bool)
	// SetRecordSpecification stores a record specification in the module kv store.
	SetRecordSpecification(sdk.Context, types.RecordSpecification)
	// RemoveRecordSpecification removes a record specification from the module kv store.
	RemoveRecordSpecification(sdk.Context, types.MetadataAddress) error

	// IterateRecordSpecs processes all record specs using a given handler.
	IterateRecordSpecs(ctx sdk.Context, handler func(specification types.RecordSpecification) (stop bool)) error
	// IterateRecordSpecsForOwner processes all record specs owned by an address using a given handler.
	IterateRecordSpecsForOwner(ctx sdk.Context, ownerAddress sdk.AccAddress, handler func(recordSpecID types.MetadataAddress) (stop bool)) error
	// IterateRecordSpecsForContractSpec processes all record specs for a contract spec using a given handler.
	IterateRecordSpecsForContractSpec(ctx sdk.Context, contractSpecID types.MetadataAddress, handler func(recordSpecID types.MetadataAddress) (stop bool)) error
	// GetRecordSpecificationsForContractSpecificationID returns all the record specifications associated with given contractSpecID
	GetRecordSpecificationsForContractSpecificationID(ctx sdk.Context, contractSpecID types.MetadataAddress) ([]*types.RecordSpecification, error)

	// GetOsLocatorRecord returns the OS locator records for a given name record.
	GetOsLocatorRecord(ctx sdk.Context, ownerAddr sdk.AccAddress) (types.ObjectStoreLocator, bool)
	// return if OSLocator exists for a given owner addr
	OSLocatorExists(ctx sdk.Context, ownerAddr sdk.AccAddress) bool
	// add OSLocator instance
	SetOSLocator(ctx sdk.Context, ownerAddr, encryptionKey sdk.AccAddress, uri string) error
	// get OS locator by scope UUID.
	GetOSLocatorByScope(ctx sdk.Context, scopeID string) ([]types.ObjectStoreLocator, error)
}

// Keeper is the concrete state-based API for the metadata module.
type Keeper struct {
	// Key to access the key-value store from sdk.Context
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	// To check if accounts exist and set public keys.
	authKeeper AuthKeeper

	// To check granter grantee authorization of messages.
	authzKeeper AuthzKeeper

	// For getting/setting account data.
	attrKeeper AttrKeeper
}

// NewKeeper creates new instances of the metadata Keeper.
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace,
	authKeeper AuthKeeper, authzKeeper AuthzKeeper, attrKeeper AttrKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.OSParamKeyTable())
	}
	return Keeper{
		storeKey:    key,
		cdc:         cdc,
		paramSpace:  paramSpace,
		authKeeper:  authKeeper,
		authzKeeper: authzKeeper,
		attrKeeper:  attrKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

var _ MetadataKeeperI = &Keeper{}

// VerifyCorrectOwner to determines whether the signer resolves to the owner of the OSLocator record.
func (k Keeper) VerifyCorrectOwner(ctx sdk.Context, ownerAddr sdk.AccAddress) bool {
	stored, found := k.GetOsLocatorRecord(ctx, ownerAddr)
	if !found {
		return false
	}
	return ownerAddr.String() == stored.Owner
}

func (k Keeper) EmitEvent(ctx sdk.Context, event proto.Message) {
	err := ctx.EventManager().EmitTypedEvent(event)
	if err != nil {
		ctx.Logger().Error("unable to emit event", "error", err, "event", event)
	}
}

// unionUnique gets a union of the provided sets of strings without any duplicates.
func (k Keeper) UnionDistinct(sets ...[]string) []string {
	retval := []string{}
	for _, s := range sets {
		for _, v := range s {
			f := false
			for _, r := range retval {
				if r == v {
					f = true
					break
				}
			}
			if !f {
				retval = append(retval, v)
			}
		}
	}
	return retval
}

func (k Keeper) checkValidURI(uri string, ctx sdk.Context) (*url.URL, error) {
	urlToPersist, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if urlToPersist.Scheme == "" || urlToPersist.Host == "" {
		return nil, types.ErrOSLocatorURIInvalid
	}

	if int(k.GetOSLocatorParams(ctx).MaxUriLength) < len(uri) {
		return nil, types.ErrOSLocatorURIToolong
	}
	return urlToPersist, nil
}
