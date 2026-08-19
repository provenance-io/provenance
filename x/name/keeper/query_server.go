package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/name/types"
)

var _ types.QueryServer = Keeper{}

// Params queries params of distribution module
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

// Resolve returns the address a name resolves to or an error.
func (k Keeper) Resolve(c context.Context, request *types.QueryResolveRequest) (*types.QueryResolveResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	record, err := k.GetRecordByName(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	return &types.QueryResolveResponse{Address: record.Address, Restricted: record.Restricted}, nil
}

// ReverseLookup returns a paginated list of names owned by the specified address.
func (k Keeper) ReverseLookup(c context.Context, request *types.QueryReverseLookupRequest) (*types.QueryReverseLookupResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	accAddr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// indexes.Multi satisfies query.Collection, and the primary key of each index entry is
	// the name, so the names come straight off the index keys — no record read per result.
	names, pageRes, err := query.CollectionPaginate(
		ctx,
		k.nameRecords.Indexes.AddrIndex,
		request.Pagination,
		func(key collections.Pair[sdk.AccAddress, string], _ collections.NoValue) (string, error) {
			return key.K2(), nil
		},
		query.WithCollectionPaginationPairPrefix[sdk.AccAddress, string](accAddr),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if names == nil {
		names = []string{} // so an empty result marshals as [] rather than null
	}

	return &types.QueryReverseLookupResponse{Name: names, Pagination: pageRes}, nil
}
