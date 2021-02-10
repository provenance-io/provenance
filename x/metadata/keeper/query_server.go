package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	// TODO
	return &types.ScopeResponse{}, nil
}

// GroupContext returns a specific group context within a scope (or all groups)
func (k Keeper) GroupContext(c context.Context, req *types.GroupContextRequest) (*types.GroupContextResponse, error) {
	// TODO
	return &types.GroupContextResponse{}, nil
}

// Record returns a collection of the records in a scope or a specific one by name
func (k Keeper) Record(c context.Context, req *types.RecordRequest) (*types.RecordResponse, error) {
	// TODO
	return &types.RecordResponse{}, nil
}

// Ownership returns a list of scope identifiers that list the given address as a party/owner
func (k Keeper) Ownership(c context.Context, req *types.OwnershipRequest) (*types.OwnershipResponse, error) {
	// TODO
	return &types.OwnershipResponse{}, nil
}
