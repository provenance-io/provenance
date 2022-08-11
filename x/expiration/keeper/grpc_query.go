package keeper

import (
	"context"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/provenance-io/provenance/x/expiration/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	var expirations []*types.Expiration
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	expirationStore := prefix.NewStore(store, types.ModuleAssetKeyPrefix)
	pageRes, err := query.Paginate(expirationStore, req.Pagination, func(key []byte, value []byte) error {
		var expiration types.Expiration
		if err := k.cdc.Unmarshal(value, &expiration); err != nil {
			return err
		}
		expirations = append(expirations, &expiration)
		return nil
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

	var expirations []*types.Expiration
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	expirationStore := prefix.NewStore(store, types.ModuleAssetKeyPrefix)
	pageRes, err := query.FilteredPaginate(expirationStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var expiration types.Expiration
		if err := k.cdc.Unmarshal(value, &expiration); err != nil {
			return false, err
		}
		if strings.TrimSpace(req.Owner) != expiration.Owner {
			return false, nil
		}
		if accumulate {
			expirations = append(expirations, &expiration)

		}
		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllExpirationsByOwnerResponse{Expirations: expirations, Pagination: pageRes}, nil
}
