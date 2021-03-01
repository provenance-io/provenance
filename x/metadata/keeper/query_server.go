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

	groups := []*types.RecordGroup{}
	err = k.IterateGroups(ctx, scopeAddress, func(rg types.RecordGroup) (stop bool) {
		groups = append(groups, &rg)
		return false
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't iterate scope groups %v", err)
	}
	return &types.ScopeResponse{Scope: &s, Records: records, RecordGroups: groups}, nil
}

// GroupContext returns a specific group context within a scope (or all groups)
func (k Keeper) GroupContext(c context.Context, req *types.GroupContextRequest) (*types.GroupContextResponse, error) {
	// TODO
	return &types.GroupContextResponse{}, nil
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
	records, err := k.record(c, &scopeAddr, req.Name)
	if err != nil {
		return nil, err
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

	records, err := k.record(c, &scopeAddr, req.Name)
	if err != nil {
		return nil, err
	}

	return &types.RecordsByScopeIDResponse{ScopeUuid: scopeUUID.String(), ScopeId: req.GetScopeId(), Records: records}, nil
}

// record returns a collection of the records in a scope or a specific one by name
func (k Keeper) record(c context.Context, scopeAddress *types.MetadataAddress, name string) ([]*types.Record, error) {
	ctx := sdk.UnwrapSDKContext(c)
	records := []*types.Record{}
	err := k.IterateRecords(ctx, *scopeAddress, func(r types.Record) (stop bool) {
		if name == "" {
			records = append(records, &r)
		} else if name == r.Name {
			records = append(records, &r)
		}
		return false
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to iterate records: %s", err.Error())
	}

	return records, nil
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

// ScopeSpecification returns a specific scope by id
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
