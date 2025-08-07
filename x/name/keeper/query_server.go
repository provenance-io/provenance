package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	name, err := k.Normalize(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	record, err := k.nameRecords.Get(ctx, name)
	if err != nil {
		return nil, types.ErrNameNotBound
	}
	return &types.QueryResolveResponse{Address: record.Address, Restricted: record.Restricted}, nil
}

func (k Keeper) ReverseLookup(c context.Context, request *types.QueryReverseLookupRequest) (*types.QueryReverseLookupResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	accAddr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	var pageReq *query.PageRequest
	if request.Pagination != nil {
		pageReq = request.Pagination
	}

	// Create a function that only processes records matching the address
	filterFunc := func(nameKey string, record types.NameRecord) (*string, error) {
		recordAddr, err := sdk.AccAddressFromBech32(record.Address)
		if err != nil {
			return nil, nil
		}

		if !recordAddr.Equals(accAddr) {
			return nil, nil
		}

		return &record.Name, nil
	}
	rv := &types.QueryReverseLookupResponse{}
	// Use CollectionsPaginate with filtering function
	names, pagination, err := query.CollectionPaginate(ctx, k.nameRecords, pageReq, filterFunc)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rv.Name = make([]string, 0, len(names))
	for _, namePtr := range names {
		if namePtr != nil {
			rv.Name = append(rv.Name, *namePtr)
		}
	}
	rv.Pagination = pagination

	return rv, nil
}
