package keeper

import (
	"context"

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

// GetRegistry returns the registry for a given key
func (qs QueryServer) GetRegistry(ctx context.Context, req *types.QueryGetRegistryRequest) (*types.QueryGetRegistryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	reg, err := qs.keeper.GetRegistry(sdkCtx, req.Key)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetRegistryResponse{Registry: *reg}, nil
}

// HasRole returns true if the address has the role for the given key
func (qs QueryServer) HasRole(ctx context.Context, req *types.QueryHasRoleRequest) (*types.QueryHasRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	keyStr := req.Key.String()

	// ensure the registry exists
	has, err := qs.keeper.Registry.Has(sdkCtx, keyStr)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(keyStr)
	}

	hasRole, err := qs.keeper.HasRole(sdkCtx, req.Key, req.Role, req.Address)
	if err != nil {
		return nil, err
	}

	return &types.QueryHasRoleResponse{HasRole: hasRole}, nil
}
