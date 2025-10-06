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

// ReverseLookup returns a paginated list of names owned by the specified address.
func (k Keeper) ReverseLookup(c context.Context, request *types.QueryReverseLookupRequest) (*types.QueryReverseLookupResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	accAddr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	limit := request.Pagination.GetLimit()
	if limit == 0 {
		limit = query.DefaultLimit
	}
	const maxLimit = 200
	if limit > maxLimit {
		limit = maxLimit
	}

	// Continuation key from previous page
	var continueAfterName string
	if len(request.Pagination.GetKey()) > 0 {
		continueAfterName = string(request.Pagination.GetKey())
	}

	// Create range for this address
	refKeyPrefix := collections.PairPrefix[sdk.AccAddress, string](accAddr)
	prefixRange := collections.NewPrefixedPairRange[
		collections.Pair[sdk.AccAddress, string],
		string,
	](refKeyPrefix)

	iter, err := k.nameRecords.Indexes.AddrIndex.Iterate(ctx, prefixRange)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer iter.Close()

	// If we have a continuation key, skip forward past it (inclusive)
	skipMode := continueAfterName != ""

	var names []string
	var lastKey string

	for ; iter.Valid(); iter.Next() {
		pk, err := iter.PrimaryKey()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if skipMode {
			if pk == continueAfterName {
				skipMode = false
				continue
			}
			if pk < continueAfterName {
				continue
			}
			skipMode = false
		}
		if uint64(len(names)) >= limit {
			break
		}
		record, err := k.nameRecords.Get(ctx, pk)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		names = append(names, record.Name)
		lastKey = pk
	}

	var nextKey []byte
	if iter.Valid() {
		nextKey = []byte(lastKey)
	}

	return &types.QueryReverseLookupResponse{
		Name: names,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
		},
	}, nil
}
