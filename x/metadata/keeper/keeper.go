package keeper

import (
	"fmt"

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

	//GetOSLocator returns the OS locator records for a given name record.
	GetOsLocatorRecord(sdk.Context, types.MetadataAddress)
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

// ValidateAllOwnersAreSigners makes sure that all entries in the existingOwners list are contained in the signers list.
func (k Keeper) ValidateAllOwnersAreSigners(existingOwners []string, signers []string) error {
	for _, owner := range existingOwners {
		found := false
		for _, signer := range signers {
			if owner == signer {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing signature from existing owner %s; required for update", owner)
		}
	}
	return nil
}

// ValidateRequiredSignatures validate all owners are signers
func (k Keeper) ValidateRequiredSignatures(owners []types.Party, signers []string) error {
	// Validate any changes to the ValueOwner property.
	requiredSignatures := []string{}
	for _, p := range owners {
		requiredSignatures = append(requiredSignatures, p.Address)
	}

	if err := k.ValidateAllOwnersAreSigners(requiredSignatures, signers); err != nil {
		return err
	}
	return nil
}

// ValidatePartiesInvolved validate that all required parties are involved
func (k Keeper) ValidatePartiesInvolved(parties []types.Party, requiredParties []types.PartyType) error {
	for _, pi := range requiredParties {
		found := false
		for _, p := range parties {
			if p.Role.String() == pi.String() {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing party type from required parties %s", pi.String())
		}
	}
	return nil
}
