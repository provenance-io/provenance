package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/google/uuid"

	"github.com/provenance-io/provenance/x/metadata/types"
)

var _ types.QueryServer = Keeper{}

// NameRecords within the GenesisState
type ObjectStoreLocators []types.ObjectStoreLocator

// Params queries params of metadata module
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Scope returns a specific scope by id
func (k Keeper) Scope(c context.Context, req *types.ScopeRequest) (*types.ScopeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ScopeUuid == "" {
		return nil, status.Error(codes.InvalidArgument, "scope uuid cannot be empty")
	}

	scopeUUID, err := uuid.Parse(req.ScopeUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope uuid: %s", err.Error())
	}
	scopeAddress := types.ScopeMetadataAddress(scopeUUID)

	retval := types.ScopeResponse{ScopeUuid: scopeUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	scope, found := k.GetScope(ctx, scopeAddress)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "scope uuid %s not found", scopeUUID)
	}
	retval.Scope = &scope

	err = k.IterateRecords(ctx, scopeAddress, func(record types.Record) (stop bool) {
		retval.Records = append(retval.Records, &record)
		return false
	})
	if err != nil {
		return &retval, status.Errorf(codes.Internal, "error iterating scope records: %v", err)
	}

	err = k.IterateSessions(ctx, scopeAddress, func(session types.Session) (stop bool) {
		retval.Sessions = append(retval.Sessions, &session)
		return false
	})
	if err != nil {
		return &retval, status.Errorf(codes.Internal, "error iterating scope sessions: %v", err)
	}

	return &retval, nil
}

// SessionContextByUUID returns a specific group context within a scope (or all groups)
func (k Keeper) SessionContextByUUID(c context.Context, req *types.SessionContextByUUIDRequest) (*types.SessionContextByUUIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.GetScopeUuid() == "" {
		return nil, status.Error(codes.InvalidArgument, "scope uuid cannot be empty")
	}

	scopeUUID, err := uuid.Parse(req.GetScopeUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope uuid: %s", err.Error())
	}

	scopeID := types.ScopeMetadataAddress(scopeUUID)

	ctx := sdk.UnwrapSDKContext(c)
	sessions := []*types.Session{}
	if req.GetSessionUuid() == "" {
		err = k.IterateSessions(ctx, scopeID, func(s types.Session) (stop bool) {
			sessions = append(sessions, &s)
			return false
		})
		if err != nil {
			return nil, err
		}
		return &types.SessionContextByUUIDResponse{ScopeId: scopeID.String(), Sessions: sessions}, nil
	}

	sessionUUID, err := uuid.Parse(req.GetSessionUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session uuid: %s", err.Error())
	}

	sessionID := types.SessionMetadataAddress(scopeUUID, sessionUUID)

	session, found := k.GetSession(ctx, sessionID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "session id %s not found", sessionID)
	}
	sessions = append(sessions, &session)
	return &types.SessionContextByUUIDResponse{ScopeId: scopeID.String(), SessionId: sessionID.String(), Sessions: sessions}, nil
}

// SessionContextByID returns a specific session context within a scope (or all groups)
func (k Keeper) SessionContextByID(c context.Context, req *types.SessionContextByIDRequest) (*types.SessionContextByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.GetScopeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "scope id cannot be empty")
	}

	scopeID, err := types.MetadataAddressFromBech32(req.GetScopeId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "incorrect scope id: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	sessions := []*types.Session{}
	if req.GetSessionId() == "" {
		err = k.IterateSessions(ctx, scopeID, func(s types.Session) (stop bool) {
			sessions = append(sessions, &s)
			return false
		})
		if err != nil {
			return nil, err
		}
		return &types.SessionContextByIDResponse{ScopeId: scopeID.String(), Sessions: sessions}, nil
	}

	sessionID, err := types.MetadataAddressFromBech32(req.GetSessionId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "incorrect scope id: %s", err.Error())
	}

	session, found := k.GetSession(ctx, sessionID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "session id %s not found", session.SessionId)
	}
	sessions = append(sessions, &session)
	return &types.SessionContextByIDResponse{ScopeId: scopeID.String(), SessionId: sessionID.String(), Sessions: sessions}, nil
}

// RecordsByScopeUUID returns a collection of the records in a scope or a specific one by name
func (k Keeper) RecordsByScopeUUID(c context.Context, req *types.RecordsByScopeUUIDRequest) (*types.RecordsByScopeUUIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.GetScopeUuid() == "" {
		return nil, status.Error(codes.InvalidArgument, "scope uuid cannot be empty")
	}

	scopeUUID, err := uuid.Parse(req.GetScopeUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope uuid: %s", err.Error())
	}

	scopeAddr := types.ScopeMetadataAddress(scopeUUID)
	ctx := sdk.UnwrapSDKContext(c)
	records, err := k.GetRecords(ctx, scopeAddr, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get records: %s", err.Error())
	}

	return &types.RecordsByScopeUUIDResponse{ScopeUuid: scopeUUID.String(), ScopeId: scopeAddr.String(), Records: records}, nil
}

// RecordsByScopeID returns a collection of the records in a scope or a specific one by name
func (k Keeper) RecordsByScopeID(c context.Context, req *types.RecordsByScopeIDRequest) (*types.RecordsByScopeIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.GetScopeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "scope id cannot be empty")
	}

	scopeAddr, err := types.MetadataAddressFromBech32(req.GetScopeId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope id %s : %s", req.GetScopeId(), err)
	}

	scopeUUID, err := scopeAddr.ScopeUUID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract uuid from scope metaaddress %s", err)
	}

	ctx := sdk.UnwrapSDKContext(c)
	records, err := k.GetRecords(ctx, scopeAddr, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get records: %s", err.Error())
	}

	return &types.RecordsByScopeIDResponse{ScopeUuid: scopeUUID.String(), ScopeId: req.GetScopeId(), Records: records}, nil
}

// Ownership returns a list of scope identifiers that list the given address as a data or value owner
func (k Keeper) Ownership(c context.Context, req *types.OwnershipRequest) (*types.OwnershipResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	scopeStore := prefix.NewStore(store, types.GetAddressScopeCacheIteratorPrefix(addr))

	scopeUUIDs := make([]string, req.Pagination.Size())
	pageRes, err := query.Paginate(scopeStore, req.Pagination, func(key, _ []byte) error {
		var ma types.MetadataAddress
		if mErr := ma.Unmarshal(key); mErr != nil {
			return mErr
		}
		scopeUUID, sErr := ma.ScopeUUID()
		if sErr != nil {
			return sErr
		}
		scopeUUIDs = append(scopeUUIDs, scopeUUID.String())
		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}
	return &types.OwnershipResponse{ScopeUuids: scopeUUIDs, Pagination: pageRes}, nil
}

// ValueOwnership returns a list of scope identifiers that list the given address as a value owner
func (k Keeper) ValueOwnership(c context.Context, req *types.ValueOwnershipRequest) (*types.ValueOwnershipResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	scopeStore := prefix.NewStore(store, types.GetValueOwnerScopeCacheIteratorPrefix(addr))

	scopes := []string{}
	pageRes, err := query.Paginate(scopeStore, req.Pagination, func(key, _ []byte) error {
		var ma types.MetadataAddress
		if mErr := ma.Unmarshal(key); mErr != nil {
			return mErr
		}
		scopeID, sErr := ma.ScopeUUID()
		if sErr != nil {
			return sErr
		}
		scopes = append(scopes, scopeID.String())
		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}
	return &types.ValueOwnershipResponse{ScopeUuids: scopes, Pagination: pageRes}, nil
}

// ScopeSpecification returns a specific scope specification by id
func (k Keeper) ScopeSpecification(c context.Context, req *types.ScopeSpecificationRequest) (*types.ScopeSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.SpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "specification uuid cannot be empty")
	}

	specUUID, err := uuid.Parse(req.SpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification uuid: %s", err.Error())
	}
	specID := types.ScopeSpecMetadataAddress(specUUID)

	retval := types.ScopeSpecificationResponse{SpecificationUuid: specUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetScopeSpecification(ctx, specID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "scope specification uuid %s not found", req.SpecificationUuid)
	}
	retval.ScopeSpecification = &spec

	return &retval, nil
}

// ContractSpecification returns a specific contract specification by id
func (k Keeper) ContractSpecification(c context.Context, req *types.ContractSpecificationRequest) (*types.ContractSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.SpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "specification uuid cannot be empty")
	}

	specUUID, err := uuid.Parse(req.SpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification uuid: %s", err.Error())
	}
	specID := types.ContractSpecMetadataAddress(specUUID)

	retval := types.ContractSpecificationResponse{ContractSpecificationUuid: specUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetContractSpecification(ctx, specID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "contract specification with id %s (uuid %s) not found",
			specID, req.SpecificationUuid)
	}
	retval.ContractSpecification = &spec

	return &retval, nil
}

// ContractSpecificationExtended returns a specific contract specification and record specifications by contract specification id
func (k Keeper) ContractSpecificationExtended(c context.Context, req *types.ContractSpecificationExtendedRequest) (*types.ContractSpecificationExtendedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.SpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "specification uuid cannot be empty")
	}

	contractSpecUUID, err := uuid.Parse(req.SpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification uuid: %s", err.Error())
	}
	contractSpecID := types.ContractSpecMetadataAddress(contractSpecUUID)

	retval := types.ContractSpecificationExtendedResponse{ContractSpecificationUuid: contractSpecUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	contractSpec, found := k.GetContractSpecification(ctx, contractSpecID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "contract specification with id %s (uuid %s) not found",
			contractSpecID, req.SpecificationUuid)
	}
	retval.ContractSpecification = &contractSpec

	recSpecs, err := k.GetRecordSpecificationsForContractSpecificationID(ctx, contractSpecID)
	if err != nil {
		return &retval, status.Errorf(codes.Aborted, "error getting record specifications for contract spec uuid %s: %s",
			contractSpecUUID, err.Error())
	}
	retval.RecordSpecifications = recSpecs

	return &retval, nil
}

// RecordSpecificationsForContractSpecification returns the record specifications associated with a contract specification
func (k Keeper) RecordSpecificationsForContractSpecification(
	c context.Context,
	req *types.RecordSpecificationsForContractSpecificationRequest,
) (*types.RecordSpecificationsForContractSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.ContractSpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "contract specification uuid cannot be empty")
	}
	contractSpecUUID, err := uuid.Parse(req.ContractSpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid contract specification uuid: %s", err.Error())
	}
	contractSpecID := types.ContractSpecMetadataAddress(contractSpecUUID)

	retval := types.RecordSpecificationsForContractSpecificationResponse{ContractSpecificationUuid: contractSpecUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	recSpecs, err := k.GetRecordSpecificationsForContractSpecificationID(ctx, contractSpecID)
	if err != nil {
		return &retval, status.Errorf(codes.Aborted, "error getting record specifications for contract spec uuid %s: %s",
			contractSpecUUID, err.Error())
	}
	if len(recSpecs) == 0 {
		return &retval, status.Errorf(codes.NotFound, "no record specifications found for contract spec uuid %s", contractSpecUUID)
	}
	retval.RecordSpecifications = recSpecs

	return &retval, err
}

// RecordSpecification returns a specific record specification by contract spec uuid and record name
func (k Keeper) RecordSpecification(c context.Context, req *types.RecordSpecificationRequest) (*types.RecordSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.ContractSpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "contract specification uuid cannot be empty")
	}
	contractSpecUUID, err := uuid.Parse(req.ContractSpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid contract specification uuid: %s", err.Error())
	}

	if len(strings.TrimSpace(req.Name)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	recSpecID := types.RecordSpecMetadataAddress(contractSpecUUID, req.Name)

	retval := types.RecordSpecificationResponse{
		ContractSpecificationUuid: contractSpecUUID.String(),
		Name:                      req.Name,
	}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetRecordSpecification(ctx, recSpecID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "record specification not found for id %s (contract spec uuid %s and name %s)",
			recSpecID, req.ContractSpecificationUuid, req.Name)
	}
	retval.RecordSpecification = &spec

	return &retval, nil
}

// RecordSpecification returns a specific record specification by contract spec uuid and record name
func (k Keeper) RecordSpecificationByID(c context.Context, req *types.RecordSpecificationByIDRequest) (*types.RecordSpecificationByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.RecordSpecificationId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "record specification id cannot be empty")
	}

	recSpecID, err := types.MetadataAddressFromBech32(req.RecordSpecificationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid record specification id: %s", err.Error())
	}
	if !recSpecID.IsRecordSpecificationAddress() {
		return nil, status.Errorf(codes.InvalidArgument, "metadata address %s is not a record specification id", recSpecID.String())
	}

	retval := types.RecordSpecificationByIDResponse{RecordSpecificationId: recSpecID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetRecordSpecification(ctx, recSpecID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "record specification not found for id %s", recSpecID)
	}
	retval.RecordSpecification = &spec

	return &retval, nil
}

func (k Keeper) OSLocatorParams(c context.Context, request *types.OSLocatorQueryParamsRequest) (*types.OSLocatorQueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.OSLocatorParams
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.OSLocatorQueryParamsResponse{Params: params}, nil
}

func (k Keeper) OSLocator(c context.Context, request *types.OSLocatorRequest) (*types.OSLocatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	accAddr, err := sdk.AccAddressFromBech32(request.Owner)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	record, exists := k.GetOsLocatorRecord(ctx, accAddr)

	if !exists {
		return nil, types.ErrAddressNotBound
	}
	return &types.OSLocatorResponse{Locator: &record}, nil
}

func (k Keeper) OSLocatorByURI(ctx context.Context, request *types.OSLocatorByURIRequest) (*types.OSLocatorResponses, error) {
	ctxSDK := sdk.UnwrapSDKContext(ctx)
	// Return value data structure.
	var records []types.ObjectStoreLocator
	// Handler that adds records if account address matches.
	appendToRecords := func(record types.ObjectStoreLocator) bool {
		if record.LocatorUri == request.Uri {
			records = append(records, record)
			// have to get all the uri associated with an address..imo..check
		}
		return false
	}

	if err := k.IterateLocators(ctxSDK, appendToRecords); err != nil {
		return nil, err
	}
	if records == nil {
		return nil, types.ErrNoRecordsFound
	}
	uniqueRecords := uniqueRecords(records)
	return &types.OSLocatorResponses{Locator: uniqueRecords}, nil
}

func (k Keeper) OSLocatorByScopeUUID(ctx context.Context, request *types.ScopeRequest) (*types.OSLocatorScopeResponse, error) {
	ctxSDK := sdk.UnwrapSDKContext(ctx)
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if request.ScopeUuid == "" {
		return nil, status.Error(codes.InvalidArgument, "scope uuid cannot be empty")
	}

	return k.GetOSLocatorByScopeUUID(ctxSDK, request.ScopeUuid)
}
