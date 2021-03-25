package keeper

import (
	"fmt"
	"net/url"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/provenance-io/provenance/x/metadata/types"
	"github.com/tendermint/tendermint/libs/log"
)

// MetadataKeeperI is the internal state api for the metadata module.
type MetadataKeeperI interface {
	// GetScope returns the scope with the given address.
	GetScope(sdk.Context, types.MetadataAddress) (types.Scope, bool)
	// SetScope persists the provided scope
	SetScope(sdk.Context, types.Scope)
	// RemoveScope removes the provided scope
	RemoveScope(sdk.Context, types.MetadataAddress)

	// IterateScopes processes all stored scopes with the given handler.
	IterateScopes(sdk.Context, func(types.Scope) bool) error

	// GetSession returns the scope with the given address.
	GetSession(sdk.Context, types.MetadataAddress) (types.Session, bool)
	// SetSession persists the provided scope
	SetSession(sdk.Context, types.Session)
	// RemoveSession persists the provided scope
	RemoveSession(sdk.Context, types.MetadataAddress)

	// GetRecord returns the record with the given address.
	GetRecord(sdk.Context, types.MetadataAddress) (types.Record, bool)
	// GetRecords returns records with giving scope and/or name
	GetRecords(sdk.Context, types.MetadataAddress, string) ([]*types.Record, error)
	// SetRecord persists the provided record
	SetRecord(sdk.Context, types.Record)
	// RemoveRecord persists the provided scope
	RemoveRecord(sdk.Context, types.MetadataAddress)

	// IterateRecords processes all stored record for a scope with the given handler.
	IterateRecords(sdk.Context, types.MetadataAddress, func(types.Record) bool) error

	// GetScopeSpecification returns the record with the given address.
	GetScopeSpecification(sdk.Context, types.MetadataAddress) (types.ScopeSpecification, bool)
	// SetScopeSpecification persists the provided scope specification
	SetScopeSpecification(sdk.Context, types.ScopeSpecification)
	// RemoveScopeSpecification removes a scope specification from the module kv store.
	RemoveScopeSpecification(sdk.Context, types.MetadataAddress) error

	// IterateScopeSpecs processes all scope specs using a given handler.
	IterateScopeSpecs(ctx sdk.Context, handler func(specification types.ScopeSpecification) (stop bool)) error
	// IterateScopeSpecsForOwner processes all scope specs owned by an address using a given handler.
	IterateScopeSpecsForOwner(ctx sdk.Context, ownerAddress sdk.AccAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error
	// IterateScopeSpecsForContractSpec processes all scope specs associated with a contract spec id using a given handler.
	IterateScopeSpecsForContractSpec(ctx sdk.Context, contractSpecID types.MetadataAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error

	// GetContractSpecification returns the contract specification with the given address.
	GetContractSpecification(sdk.Context, types.MetadataAddress) (types.ContractSpecification, bool)
	// SetContractSpecification persists the provided contract specification
	SetContractSpecification(sdk.Context, types.ContractSpecification)
	// RemoveContractSpecification removes a contract specification from the module kv store.
	RemoveContractSpecification(sdk.Context, types.MetadataAddress) error

	// IterateContractSpecs processes all contract specs using the given handler.
	IterateContractSpecs(ctx sdk.Context, handler func(specification types.ContractSpecification) (stop bool)) error
	// IterateContractSpecsForOwner processes all contract specs owned by an address using a given handler.
	IterateContractSpecsForOwner(ctx sdk.Context, ownerAddress sdk.AccAddress, handler func(contractSpecID types.MetadataAddress) (stop bool)) error

	// GetRecordSpecification returns the record specification with the given address.
	GetRecordSpecification(sdk.Context, types.MetadataAddress) (types.RecordSpecification, bool)
	// SetRecordSpecification persists the provided record specification
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
	GetOsLocatorRecord(ctx sdk.Context, ownerAddress sdk.AccAddress) (types.ObjectStoreLocator, bool)
	// return if OSLocator exists for a given owner addr
	OSLocatorExists(ctx sdk.Context, ownerAddr string) bool
	// add OSLocator instance
	SetOSLocatorRecord(ctx sdk.Context, ownerAddr sdk.AccAddress, uri string) error
	// get OS locator by scope UUID.
	GetOSLocatorByScopeUUID(ctx sdk.Context, scopeID string) (*types.OSLocatorByScopeUUIDResponse, error)
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
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.OSParamKeyTable())
	}
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

// ValidateAllOwnersAreSigners makes sure that all entries in the existingOwners list are contained in the signers list.
func (k Keeper) ValidateAllOwnersAreSigners(existingOwners []string, signers []string) error {
	missing := FindMissing(existingOwners, signers)
	switch len(missing) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("missing signature from existing owner %s; required for update", missing[0])
	default:
		return fmt.Errorf("missing signatures from existing owners %v; required for update", missing)
	}
}

// ValidateAllPartiesAreSigners validate all parties are signers
func (k Keeper) ValidateAllPartiesAreSigners(parties []types.Party, signers []string) error {
	addresses := make([]string, len(parties))
	for i, party := range parties {
		addresses[i] = party.Address
	}
	missing := FindMissing(addresses, signers)
	switch len(missing) {
	case 0:
		return nil
	case 1:
		for _, party := range parties {
			if party.Address == missing[0] {
				return fmt.Errorf("missing signature from %s (%s)", party.Address, party.Role.String())
			}
		}
		// Should never get here, but the compiler can't tell that.
		return fmt.Errorf("missing signature from %s", missing[0])
	default:
		missingWithRoles := make([]string, len(missing))
		for i, addr := range missing {
			for _, party := range parties {
				if addr == party.Address {
					missingWithRoles[i] = fmt.Sprintf("%s (%s)", addr, party.Role.String())
					break
				}
			}
		}
		return fmt.Errorf("missing signatures from %v", missingWithRoles)
	}
}

// ValidatePartiesInvolved validate that all required parties are involved
func (k Keeper) ValidatePartiesInvolved(parties []types.Party, requiredParties []types.PartyType) error {
	partyRoles := make([]string, len(parties))
	reqRoles := make([]string, len(requiredParties))
	for i, party := range parties {
		partyRoles[i] = party.Role.String()
	}
	for i, req := range requiredParties {
		reqRoles[i] = req.String()
	}
	missing := FindMissing(reqRoles, partyRoles)
	switch len(missing) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("missing required party type %s from parties", missing[0])
	default:
		return fmt.Errorf("missing required party types %v from parties", missing)
	}
}

// FindMissing returns all elements of the required list that are not found in the entries list
// It is exported only so that it can be unit tested.
func FindMissing(required []string, entries []string) []string {
	retval := []string{}
	for _, req := range required {
		found := false
		for _, entry := range entries {
			if req == entry {
				found = true
				break
			}
		}
		if !found {
			retval = append(retval, req)
		}
	}
	return retval
}

// VerifyCorrectOwner to determines whether the signer resolves to the owner of the OSLocator record.
func (k Keeper) VerifyCorrectOwner(ctx sdk.Context, ownerAddr sdk.AccAddress) bool { // nolint:interfacer
	stored, found := k.GetOsLocatorRecord(ctx, ownerAddr)
	if !found {
		return false
	}
	return ownerAddr.String() == stored.Owner
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
