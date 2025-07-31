package keeper

import (
	"context"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

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
	record, err := k.GetRecordByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, types.ErrNameNotBound
	}
	return &types.QueryResolveResponse{Address: record.Address, Restricted: record.Restricted}, nil
}

// ReverseLookup gets all names bound to an address with proper pagination
// ReverseLookup gets all names bound to an address with proper pagination
func (k Keeper) ReverseLookup(c context.Context, request *types.QueryReverseLookupRequest) (*types.QueryReverseLookupResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	accAddr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Get the prefix for the address
	addrPrefix, err := types.GetAddressKeyPrefix(accAddr)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Create a range for the prefix
	rng := new(collections.Range[[]byte])
	rng = rng.StartInclusive(addrPrefix)
	rng = rng.EndExclusive(storetypes.PrefixEndBytes(addrPrefix))

	// Handle pagination start key
	if request.Pagination != nil && len(request.Pagination.Key) > 0 {
		rng = rng.StartInclusive(request.Pagination.Key)
	}

	var names []string
	var nextKey []byte
	limit := query.DefaultLimit
	if request.Pagination != nil {
		limit = int(request.Pagination.Limit)
	}

	// Iterate through address index with pagination
	count := 0
	err = k.addrIndex.Walk(ctx, rng, func(key []byte, record types.NameRecord) (bool, error) {
		// Break if we've reached the limit
		if count >= limit {
			nextKey = key
			return true, nil // Stop iteration
		}

		names = append(names, record.Name)
		count++
		return false, nil
	})

	if err != nil {
		return nil, err
	}

	// Build pagination response
	pageRes := &query.PageResponse{
		NextKey: nextKey,
	}

	return &types.QueryReverseLookupResponse{Name: names, Pagination: pageRes}, nil
}
