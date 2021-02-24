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

	if req.ScopeId == "" {
		return nil, status.Error(codes.InvalidArgument, "scope id cannot be empty")
	}

	id, err := uuid.Parse(req.ScopeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope id: %s", err.Error())
	}
	scopeAddress := types.ScopeMetadataAddress(id)
	ctx := sdk.UnwrapSDKContext(c)

	s, found := k.GetScope(ctx, scopeAddress)
	if !found {
		return nil, status.Errorf(codes.NotFound, "scope %s not found", req.ScopeId)
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

// Record returns a collection of the records in a scope or a specific one by name
func (k Keeper) Record(c context.Context, req *types.RecordRequest) (*types.RecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ScopeId == "" {
		return nil, status.Error(codes.InvalidArgument, "scope id cannot be empty")
	}

	id, err := uuid.Parse(req.ScopeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope id: %s", err.Error())
	}

	records := []*types.Record{}
	k.IterateRecords(ctx, types.ScopeMetadataAddress(id), func(r types.Record) (stop bool) {
		if req.Name == "" {
			records = append(records, &r)
		} else if req.Name == r.Name {
			records = append(records, &r)
		}
		return false
	})

	return &types.RecordResponse{ScopeId: req.ScopeId, Records: records}, nil
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

	scopes := make([]string, req.Pagination.Size())
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
	return &types.OwnershipResponse{ScopeIds: scopes, Pagination: pageRes}, nil
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
	return &types.ValueOwnershipResponse{ScopeIds: scopes, Pagination: pageRes}, nil
}

// ScopeSpecification returns a specific scope by id
func (k Keeper) ScopeSpecification(c context.Context, req *types.ScopeSpecificationRequest) (*types.ScopeSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.SpecificationId == "" {
		return nil, status.Error(codes.InvalidArgument, "specification id cannot be empty")
	}

	id, err := uuid.Parse(req.SpecificationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification id: %s", err.Error())
	}
	addr := types.ScopeSpecMetadataAddress(id)
	ctx := sdk.UnwrapSDKContext(c)

	spec, found := k.GetScopeSpecification(ctx, addr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "scope specification %s not found", req.SpecificationId)
	}

	return &types.ScopeSpecificationResponse{ScopeSpecification: &spec}, nil
}
