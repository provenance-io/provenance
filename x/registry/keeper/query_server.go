package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/registry"
)

// QueryServer implements the gRPC querier service for the registry module
type QueryServer struct {
	keeper RegistryKeeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(keeper RegistryKeeper) *QueryServer {
	return &QueryServer{keeper: keeper}
}

// GetRegistry returns the registry for a given key
func (qs QueryServer) GetRegistry(ctx context.Context, req *registry.QueryGetRegistryRequest) (*registry.QueryGetRegistryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	reg, err := qs.keeper.GetRegistry(sdkCtx, req.Key)
	if err != nil {
		return nil, err
	}

	return &registry.QueryGetRegistryResponse{Registry: *reg}, nil
}

// HasRole returns true if the address has the role for the given key
func (qs QueryServer) HasRole(ctx context.Context, req *registry.QueryHasRoleRequest) (*registry.QueryHasRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	hasRole, err := qs.keeper.HasRole(sdkCtx, req.Key, req.Role, req.Address)
	if err != nil {
		return nil, err
	}

	return &registry.QueryHasRoleResponse{HasRole: hasRole}, nil
}
