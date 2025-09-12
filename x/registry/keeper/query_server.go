package keeper

import (
	"context"

	"cosmossdk.io/collections"

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
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	registries, pageRes, err := qs.keeper.GetRegistries(sdkCtx, req.Pagination, req.AssetClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetRegistriesResponse{Registries: registries, Pagination: pageRes}, nil
}

// HasRole returns true if the address has the specified role for the given key.
func (qs QueryServer) HasRole(ctx context.Context, req *types.QueryHasRoleRequest) (*types.QueryHasRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// ensure the registry exists
	has, err := qs.keeper.Registry.Has(sdkCtx, collections.Join(req.Key.AssetClassId, req.Key.NftId))
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
