package keeper

import (
	"context"

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

// ReverseLookup gets all names bound to an address with proper pagination
func (k Keeper) ReverseLookup(c context.Context, request *types.QueryReverseLookupRequest) (*types.QueryReverseLookupResponse, error) {

	ctx := sdk.UnwrapSDKContext(c)
	accAddr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Use the address index
	iter, err := k.nameRecords.Indexes.AddrIndex.MatchExact(ctx, accAddr)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var names []string
	var nextKey []byte
	limit := query.DefaultLimit
	offset := uint64(0)
	startAfter := ""

	// Handle pagination parameters
	if request.Pagination != nil {
		if request.Pagination.Limit > 0 {
			limit = int(request.Pagination.Limit)
		}
		offset = request.Pagination.Offset
		if len(request.Pagination.Key) > 0 {
			startAfter = string(request.Pagination.Key)
		}
	}

	count := uint64(0)
	skipping := offset > 0 || startAfter != ""

	for ; iter.Valid(); iter.Next() {
		// Get the primary key (name)
		name, err := iter.PrimaryKey()
		if err != nil {
			return nil, err
		}

		// Skip until we reach the startAfter point
		if startAfter != "" {
			if name == startAfter {
				startAfter = ""
			}
			continue
		}

		// Skip offset items
		if skipping && offset > 0 {
			offset--
			continue
		}

		// Stop if we've reached the limit
		if count >= uint64(limit) {
			nextKey = []byte(name)
			break
		}

		names = append(names, name)
		count++
	}

	pageRes := &query.PageResponse{
		NextKey: nextKey,
	}

	return &types.QueryReverseLookupResponse{
		Name:       names,
		Pagination: pageRes,
	}, nil
}
