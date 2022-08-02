package keeper

import (
	"fmt"
	"net/url"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzKeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

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
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	// To check if accounts exist and set public keys.
	authKeeper authkeeper.AccountKeeper

	// To check granter grantee authorization of messages.
	authzKeeper authzKeeper.Keeper
}

// NewKeeper creates new instances of the metadata Keeper.
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	authKeeper authkeeper.AccountKeeper,
	authzKeeper authzKeeper.Keeper,
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

// GetMessageTypeURLs return a hierarchical list of message type urls.
// For example passing in `/provenance.metadata.v1.MsgAddScopeDataAccessRequest` would return a list containing
// ("/provenance.metadata.v1.MsgAddScopeDataAccessRequest", "/provenance.metadata.v1.MsgWriteScopeRequest")
func (k Keeper) GetMessageTypeURLs(msgTypeURL string) []string {
	urls := []string{}
	if len(msgTypeURL) > 0 {
		urls = append(urls, msgTypeURL)
	}
	switch msgTypeURL {
	case types.TypeURLMsgAddScopeDataAccessRequest, types.TypeURLMsgDeleteScopeDataAccessRequest,
		types.TypeURLMsgAddScopeOwnerRequest, types.TypeURLMsgDeleteScopeOwnerRequest:
		urls = append(urls, types.TypeURLMsgWriteScopeRequest)
	case types.TypeURLMsgWriteRecordRequest:
		urls = append(urls, types.TypeURLMsgWriteSessionRequest)
	case types.TypeURLMsgAddContractSpecToScopeSpecRequest, types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest:
		urls = append(urls, types.TypeURLMsgWriteScopeSpecificationRequest)
	case types.TypeURLMsgWriteRecordSpecificationRequest:
		urls = append(urls, types.TypeURLMsgWriteContractSpecificationRequest)
	case types.TypeURLMsgDeleteRecordSpecificationRequest:
		urls = append(urls, types.TypeURLMsgDeleteContractSpecificationRequest)
	}
	return urls
}

// checkAuthZForMissing checks to see if the missing types.Party have an assigned grantee that can sing on their behalf
func (k Keeper) checkAuthzForMissing(ctx sdk.Context, addrs []string, signers []string, msgTypeURL string) []string {
	stillMissing := []string{}
	// return as a list this message type and its parent
	// type if it is a message belonging to a hierarchy
	msgTypeURLs := k.GetMessageTypeURLs(msgTypeURL)

	for _, addr := range addrs {
		found := false
		granter := types.MustAccAddressFromBech32(addr)

		// loop through all the signers
		for _, signer := range signers {
			grantee := types.MustAccAddressFromBech32(signer)

			for _, msgType := range msgTypeURLs {
				authz, _ := k.authzKeeper.GetCleanAuthorization(ctx, grantee, granter, msgType)
				if authz != nil {
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			stillMissing = append(stillMissing, addr)
		}
	}

	return stillMissing
}

// ValidateAllOwnersAreSignersWithAuthz makes sure that all entries in the existingOwners list
// are contained in the signers list and checks to see if any missing entries have an assigned grantee
func (k Keeper) ValidateAllOwnersAreSignersWithAuthz(
	ctx sdk.Context,
	existingOwners []string,
	signers []string,
	msgTypeURL string,
) error {
	missing := FindMissing(existingOwners, signers)
	stillMissing := missing
	// Authz grants rights to address on specific message types.
	// If no message type URL is provided, skip the Authz check.
	if len(msgTypeURL) > 0 {
		stillMissing = k.checkAuthzForMissing(ctx, missing, signers, msgTypeURL)
	}

	switch len(stillMissing) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("missing signature from existing owner %s; required for update", stillMissing[0])
	default:
		return fmt.Errorf("missing signatures from existing owners %v; required for update", stillMissing)
	}
}

// ValidateAllPartiesAreSignersWithAuthz validate all parties are signers with authz module
func (k Keeper) ValidateAllPartiesAreSignersWithAuthz(ctx sdk.Context, parties []types.Party, signers []string, msgTypeURL string) error {
	addresses := make([]string, len(parties))
	for i, party := range parties {
		addresses[i] = party.Address
	}

	missing := FindMissing(addresses, signers)
	stillMissing := missing
	// Authz grants rights to address on specific message types.
	// If no message type URL is provided, skip the Authz check.
	if len(msgTypeURL) > 0 {
		stillMissing = k.checkAuthzForMissing(ctx, missing, signers, msgTypeURL)
	}

	if len(stillMissing) > 0 {
		missingWithRoles := make([]string, len(missing))
		for i, addr := range stillMissing {
			for _, party := range parties {
				if addr == party.Address {
					missingWithRoles[i] = fmt.Sprintf("%s (%s)", addr, party.Role.String())
					break
				}
			}
		}
		return fmt.Errorf("missing signature%s from %v", pluralEnding(len(missing)), missingWithRoles)
	}

	return nil
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
	if len(missing) > 0 {
		return fmt.Errorf("missing required party type%s %v from parties", pluralEnding(len(missing)), missing)
	}
	return nil
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

// FindMissingMdAddr returns all elements of the required list that are not found in the entries list
// It is exported only so that it can be unit tested.
func FindMissingMdAddr(required, entries []types.MetadataAddress) []types.MetadataAddress {
	retval := []types.MetadataAddress{}
	for _, req := range required {
		found := false
		for _, entry := range entries {
			if req.Equals(entry) {
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

// pluralEnding returns "" if i == 1, or "s" otherwise.
func pluralEnding(i int) string {
	if i == 1 {
		return ""
	}
	return "s"
}
