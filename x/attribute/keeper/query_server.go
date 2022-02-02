package keeper

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/attribute/types"
)

var _ types.QueryServer = Keeper{}

// Params queries params of attribute module
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Attribute queries for a specific attribute
func (k Keeper) Attribute(c context.Context, req *types.QueryAttributeRequest) (*types.QueryAttributeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "empty attribute name")
	}
	if err := types.ValidateAttributeAddress(req.Account); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid account address: %v", err))
	}
	ctx := sdk.UnwrapSDKContext(c)
	attributes := make([]types.Attribute, 0)
	store := ctx.KVStore(k.storeKey)
	attributeStore := prefix.NewStore(store, types.AddrStrAttributesNameKeyPrefix(req.Account, req.Name))
	pageRes, err := query.Paginate(attributeStore, req.Pagination, func(key []byte, value []byte) error {
		var result types.Attribute
		err := k.cdc.Unmarshal(value, &result)
		if err != nil {
			return err
		}
		attributes = append(attributes, result)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.QueryAttributeResponse{Account: req.Account, Attributes: attributes, Pagination: pageRes}, nil
}

// Attributes queries for all attributes on a specified account
func (k Keeper) Attributes(c context.Context, req *types.QueryAttributesRequest) (*types.QueryAttributesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if err := types.ValidateAttributeAddress(req.Account); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid account address: %v", err))
	}
	ctx := sdk.UnwrapSDKContext(c)
	attributes := make([]types.Attribute, 0)
	store := ctx.KVStore(k.storeKey)
	attributeStore := prefix.NewStore(store, types.AddrStrAttributesKeyPrefix(req.Account))

	pageRes, err := query.Paginate(attributeStore, req.Pagination, func(key []byte, value []byte) error {
		var result types.Attribute
		err := k.cdc.Unmarshal(value, &result)
		if err != nil {
			return err
		}
		attributes = append(attributes, result)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryAttributesResponse{Account: req.Account, Attributes: attributes, Pagination: pageRes}, nil
}

// Scan queries for all attributes on a specied account that have a given suffix in their name
func (k Keeper) Scan(c context.Context, req *types.QueryScanRequest) (*types.QueryScanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Suffix == "" {
		return nil, status.Error(codes.InvalidArgument, "empty attribute name suffix")
	}
	if err := types.ValidateAttributeAddress(req.Account); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid account address: %v", err))
	}
	ctx := sdk.UnwrapSDKContext(c)
	attributes := make([]types.Attribute, 0)
	store := ctx.KVStore(k.storeKey)
	attributeStore := prefix.NewStore(store, types.AddrStrAttributesKeyPrefix(req.Account))

	pageRes, err := query.FilteredPaginate(attributeStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var result types.Attribute
		err := k.cdc.Unmarshal(value, &result)
		if err != nil {
			return false, err
		}
		if !strings.HasSuffix(result.Name, req.Suffix) {
			return false, nil
		}
		if accumulate {
			attributes = append(attributes, result)
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryScanResponse{Account: req.Account, Attributes: attributes, Pagination: pageRes}, nil
}
