package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/registry/types"
)

// QueryServer implements the gRPC querier service for the registry module
type QueryServer struct {
	keeper Keeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(keeper Keeper) *QueryServer {
	return &QueryServer{keeper: keeper}
}

// GetRegistry returns the registry entry for a given key.
// This method retrieves the complete registry entry including all roles and addresses.
func (qs QueryServer) GetRegistry(ctx context.Context, req *types.QueryGetRegistryRequest) (*types.QueryGetRegistryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	reg, err := qs.keeper.GetRegistry(sdkCtx, req.Key)
	if err != nil {
		return nil, err
	}

	if reg == nil {
		return nil, nil
	}

	return &types.QueryGetRegistryResponse{Registry: *reg}, nil
}

// GetRegistries returns the registries paginated
func (qs QueryServer) GetRegistries(ctx context.Context, req *types.QueryGetRegistriesRequest) (*types.QueryGetRegistriesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	registries, pageRes, err := qs.keeper.GetRegistries(sdkCtx, req.Pagination, req.AssetClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetRegistriesResponse{Registries: registries, Pagination: pageRes}, nil
}

// HasRole returns true if the address has the specified role for the given key.
func (qs QueryServer) HasRole(ctx context.Context, req *types.QueryHasRoleRequest) (*types.QueryHasRoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// ensure the registry exists
	has, err := qs.keeper.Registry.Has(sdkCtx, req.Key.CollKey())
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(req.Key.String())
	}

	hasRole, err := qs.keeper.HasRole(sdkCtx, req.Key, req.Role, req.Address)
	if err != nil {
		return nil, err
	}

	return &types.QueryHasRoleResponse{HasRole: hasRole}, nil
}

// PendingRoleChange returns a single pending role change by its id.
func (qs QueryServer) PendingRoleChange(ctx context.Context, req *types.QueryPendingRoleChangeRequest) (*types.QueryPendingRoleChangeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	change, err := qs.keeper.GetPendingRoleChange(sdkCtx, req.Id)
	if err != nil {
		return nil, err
	}
	if change == nil {
		return nil, types.NewErrCodePendingChangeNotFound(req.Id)
	}

	return &types.QueryPendingRoleChangeResponse{PendingRoleChange: *change}, nil
}

// PendingRoleChanges returns the pending role changes, optionally filtered by registry key.
func (qs QueryServer) PendingRoleChanges(ctx context.Context, req *types.QueryPendingRoleChangesRequest) (*types.QueryPendingRoleChangesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	changes, pageRes, err := qs.keeper.GetPendingRoleChanges(sdkCtx, req.Pagination, req.Key)
	if err != nil {
		return nil, err
	}

	return &types.QueryPendingRoleChangesResponse{PendingRoleChanges: changes, Pagination: pageRes}, nil
}

// RegistryClass returns a single registry class (including its authorization policy) by id.
func (qs QueryServer) RegistryClass(ctx context.Context, req *types.QueryRegistryClassRequest) (*types.QueryRegistryClassResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	class, err := qs.keeper.GetRegistryClass(sdkCtx, req.RegistryClassId)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, types.NewErrCodeRegistryClassNotFound(req.RegistryClassId)
	}

	return &types.QueryRegistryClassResponse{RegistryClass: *class}, nil
}

// RegistryClasses returns all registry classes, paginated.
func (qs QueryServer) RegistryClasses(ctx context.Context, req *types.QueryRegistryClassesRequest) (*types.QueryRegistryClassesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	classes, pageRes, err := qs.keeper.GetRegistryClasses(sdkCtx, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryRegistryClassesResponse{RegistryClasses: classes, Pagination: pageRes}, nil
}

// Params returns the registry module parameters.
func (qs QueryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryParamsResponse{Params: qs.keeper.GetParams(sdkCtx)}, nil
}
