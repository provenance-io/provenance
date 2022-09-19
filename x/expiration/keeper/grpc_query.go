package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/expiration/types"
)

var _ types.QueryServer = Keeper{}

// Params queries params
func (k Keeper) Params(
	goCtx context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Expiration queries for an expiration based on request parameters
func (k Keeper) Expiration(
	goCtx context.Context,
	req *types.QueryExpirationRequest,
) (*types.QueryExpirationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	expiration, err := k.GetExpiration(ctx, req.ModuleAssetId)
	if err != nil {
		return nil, err
	}

	return &types.QueryExpirationResponse{Expiration: expiration}, nil
}

// AllExpirations queries for all expirations
func (k Keeper) AllExpirations(
	goCtx context.Context,
	req *types.QueryAllExpirationsRequest,
) (*types.QueryAllExpirationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pageRes, expirations, err := k.filteredPaginate(ctx, req.Pagination, func(expiration types.Expiration) bool {
		return true // do not filter
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllExpirationsResponse{Expirations: expirations, Pagination: pageRes}, nil
}

// AllExpirationsByOwner queries all expirations for a particular owner
func (k Keeper) AllExpirationsByOwner(
	goCtx context.Context,
	req *types.QueryAllExpirationsByOwnerRequest,
) (*types.QueryAllExpirationsByOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pageRes, expirations, err := k.filteredPaginate(ctx, req.Pagination, func(expiration types.Expiration) bool {
		return strings.TrimSpace(req.Owner) == expiration.Owner
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllExpirationsByOwnerResponse{Expirations: expirations, Pagination: pageRes}, nil
}

// AllExpiredExpirations queries all expired expirations
func (k Keeper) AllExpiredExpirations(
	goCtx context.Context,
	req *types.QueryAllExpiredExpirationsRequest,
) (*types.QueryAllExpiredExpirationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pageRes, expirations, err := k.filteredPaginate(ctx, req.Pagination, func(expiration types.Expiration) bool {
		return ctx.BlockTime().After(expiration.Time)
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllExpiredExpirationsResponse{Expirations: expirations, Pagination: pageRes}, nil
}

// Private method that does pagination of the filtered results.
// onMatch predicate should be used as the filter criteria.
func (k Keeper) filteredPaginate(
	ctx sdk.Context,
	pagination *query.PageRequest,
	onMatch func(expiration types.Expiration) bool,
) (*query.PageResponse, []*types.Expiration, error) {
	var expirations []*types.Expiration
	store := ctx.KVStore(k.storeKey)
	expirationStore := prefix.NewStore(store, types.ModuleAssetKeyPrefix)
	pageRes, err := query.FilteredPaginate(expirationStore, pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var expiration types.Expiration
		if err := k.cdc.Unmarshal(value, &expiration); err != nil {
			return false, err
		}
		if !onMatch(expiration) {
			return false, nil
		}
		if accumulate {
			expirations = append(expirations, &expiration)
		}
		return true, nil
	})

	if err != nil {
		return nil, nil, err
	}

	return pageRes, expirations, nil
}
