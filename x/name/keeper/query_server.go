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
		return nil, err
	}
	return &types.QueryResolveResponse{Address: record.Address, Restricted: record.Restricted}, nil
}

// ReverseLookup using CollectionsPaginate with a custom filtered approach
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

	allRecords, err := k.GetRecordsByAddress(ctx, accAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	allNames := make([]string, len(allRecords))
	for i, record := range allRecords {
		allNames[i] = record.Name
	}

	limit := query.DefaultLimit
	var start = 0

	if pageReq != nil {
		if pageReq.Limit > 0 {
			limit = safeUint64ToInt(pageReq.Limit)
		}

		if len(pageReq.Key) > 0 {
			pageKey := string(pageReq.Key)
			for i, name := range allNames {
				if name == pageKey {
					start = i + 1
					break
				}
			}
		} else {
			start = safeUint64ToInt(pageReq.Offset)
		}
	}

	end := start + limit
	if end > len(allNames) {
		end = len(allNames)
	}

	if start >= len(allNames) {
		start = len(allNames)
		end = len(allNames)
	}

	var pageNames []string
	if start < end {
		pageNames = allNames[start:end]
	} else {
		pageNames = []string{}
	}

	rv := &types.QueryReverseLookupResponse{
		Name:       pageNames,
		Pagination: &query.PageResponse{},
	}

	if end < len(allNames) {
		rv.Pagination.NextKey = []byte(allNames[end-1])
	}

	return rv, nil
}
func safeUint64ToInt(u uint64) int {
	const maxInt = int(^uint(0) >> 1)
	if u > uint64(maxInt) {
		return maxInt
	}
	// #nosec G115 -- safe conversion due to explicit bound check above
	return int(u)
}
