package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/google/uuid"

	"github.com/provenance-io/provenance/x/metadata/types"
)

var _ types.QueryServer = Keeper{}

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

	id, err := uuid.Parse(req.ScopeUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope uuid: %s", err.Error())
	}
	scopeAddress := types.ScopeMetadataAddress(id)
	ctx := sdk.UnwrapSDKContext(c)

	s, found := k.GetScope(ctx, scopeAddress)
	if !found {
		return nil, status.Errorf(codes.NotFound, "scope uuid %s not found", req.ScopeUuid)
	}

	records := []*types.Record{}
	err = k.IterateRecords(ctx, scopeAddress, func(r types.Record) (stop bool) {
		records = append(records, &r)
		return false
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't iterate scope records %v", err)
	}

	sessions := []*types.Session{}
	err = k.IterateSessions(ctx, scopeAddress, func(rg types.Session) (stop bool) {
		sessions = append(sessions, &rg)
		return false
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't iterate scope sessions %v", err)
	}
	return &types.ScopeResponse{Scope: &s, Records: records, Sessions: sessions}, nil
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
		return nil, status.Errorf(codes.NotFound, "session id %s not found", session.SessionId)
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

	if req.SpecificationUuid == "" {
		return nil, status.Error(codes.InvalidArgument, "specification uuid cannot be empty")
	}

	id, err := uuid.Parse(req.SpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification uuid: %s", err.Error())
	}
	addr := types.ScopeSpecMetadataAddress(id)
	ctx := sdk.UnwrapSDKContext(c)

	spec, found := k.GetScopeSpecification(ctx, addr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "scope specification uuid %s not found", req.SpecificationUuid)
	}

	return &types.ScopeSpecificationResponse{ScopeSpecification: &spec}, nil
}

// ContractSpecification returns a specific scope specification by id
func (k Keeper) ContractSpecification(c context.Context, req *types.ContractSpecificationRequest) (*types.ContractSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.SpecificationUuid == "" {
		return nil, status.Error(codes.InvalidArgument, "specification id cannot be empty")
	}

	id, err := uuid.Parse(req.SpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification id: %s", err.Error())
	}
	addr := types.ContractSpecMetadataAddress(id)
	ctx := sdk.UnwrapSDKContext(c)

	spec, found := k.GetContractSpecification(ctx, addr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "contract specification %s not found", req.SpecificationUuid)
	}

	return &types.ContractSpecificationResponse{ContractSpecification: &spec}, nil
}
