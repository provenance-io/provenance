package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/attribute/types"
)

var _ types.QueryServer = Keeper{}

// Params queries params of attribute module
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid account address: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(c)
	addrBz := types.GetAttributeAddressBytes(req.Account)
	var nameHash [32]byte
	copy(nameHash[:], types.GetNameKeyBytes(req.Name))
	blockTime := ctx.BlockTime().UTC()

	rngFn := attrAddrNameRange(addrBz, types.ReverseName(req.Name))
	attrs, pageRes, err := attrPageWalk(ctx, k.attributes, rngFn, req.Pagination, func(attr types.Attribute) bool {
		return strings.EqualFold(attr.Name, req.Name) && (attr.ExpirationDate == nil || !blockTime.After(attr.ExpirationDate.UTC()))
	},
	)
	if err != nil {
		return nil, err
	}
	return &types.QueryAttributeResponse{Account: req.Account, Attributes: attrs, Pagination: pageRes}, nil
}

// Attributes queries for all attributes on a specified account
func (k Keeper) Attributes(c context.Context, req *types.QueryAttributesRequest) (*types.QueryAttributesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if err := types.ValidateAttributeAddress(req.Account); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account address: %v", err)
	}
	ctx := sdk.UnwrapSDKContext(c)
	addrBz := types.GetAttributeAddressBytes(req.Account)
	blockTime := ctx.BlockTime().UTC()

	rngFn := attrAddrRange(addrBz)
	attrs, pageRes, err := attrPageWalk(ctx, k.attributes, rngFn, req.Pagination,
		func(attr types.Attribute) bool {
			return attr.ExpirationDate == nil || !blockTime.After(attr.ExpirationDate.UTC())
		},
	)
	if err != nil {
		return nil, err
	}
	return &types.QueryAttributesResponse{Account: req.Account, Attributes: attrs, Pagination: pageRes}, nil
}

// Scan queries all attributes associated with a specified account that contain a particular suffix in their name.
func (k Keeper) Scan(c context.Context, req *types.QueryScanRequest) (*types.QueryScanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Suffix == "" {
		return nil, status.Error(codes.InvalidArgument, "empty attribute name suffix")
	}
	if err := types.ValidateAttributeAddress(req.Account); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account address: %v", err)
	}
	ctx := sdk.UnwrapSDKContext(c)
	addrBz := types.GetAttributeAddressBytes(req.Account)
	blockTime := ctx.BlockTime().UTC()

	rngFn := attrAddrRange(addrBz)
	attrs, pageRes, err := attrPageWalk(ctx, k.attributes, rngFn, req.Pagination,
		func(attr types.Attribute) bool {
			return strings.HasSuffix(attr.Name, req.Suffix) &&
				(attr.ExpirationDate == nil || !blockTime.After(attr.ExpirationDate.UTC()))
		},
	)
	if err != nil {
		return nil, err
	}
	return &types.QueryScanResponse{Account: req.Account, Attributes: attrs, Pagination: pageRes}, nil
}

// AttributeAccounts queries for all accounts that have a specific attribute
func (k Keeper) AttributeAccounts(c context.Context, req *types.QueryAttributeAccountsRequest) (*types.QueryAttributeAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	revName := types.ReverseName(req.AttributeName)

	rngFn := napNameRange(revName)
	accounts, pageRes, err := nameAddrPageWalk(ctx, k.nameAddrCounts, rngFn, req.Pagination,
		func(key types.NameAddrPair) bool {
			return key.RevName == revName
		},
	)
	if err != nil {
		return nil, err
	}
	return &types.QueryAttributeAccountsResponse{Accounts: accounts, Pagination: pageRes}, nil
}

// AccountData returns the accountdata for a specified account.
func (k Keeper) AccountData(c context.Context, req *types.QueryAccountDataRequest) (*types.QueryAccountDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	value, err := k.GetAccountData(ctx, req.Account)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	resp := &types.QueryAccountDataResponse{
		Value: value,
	}
	return resp, nil
}

// pageParams applies the PageRequest defaults the same way the SDK's query.Paginate does, so these
// handlers page exactly like the pre-collections ones: a nil request or a zero limit means
// query.DefaultLimit with a total count, a total is only counted for offset-based requests, and
// a request can't have both an offset and a key.
func pageParams(pageReq *query.PageRequest) (limit, offset uint64, countTotal bool, err error) {
	if pageReq == nil {
		pageReq = &query.PageRequest{}
	}
	if pageReq.Offset > 0 && len(pageReq.Key) > 0 {
		return 0, 0, false, status.Error(codes.InvalidArgument, "invalid request, either offset or key is expected, got both")
	}
	limit, offset, countTotal = pageReq.Limit, pageReq.Offset, pageReq.CountTotal
	if limit == 0 {
		limit = query.DefaultLimit
		countTotal = true
	}
	if len(pageReq.Key) > 0 {
		countTotal = false
	}
	return limit, offset, countTotal, nil
}

// attrPageWalk walks col over rng with full pagination:
func attrPageWalk(
	ctx sdk.Context,
	col collections.Map[types.AttrTriple, types.Attribute],
	rngFn attrRange,
	pageReq *query.PageRequest,
	accept func(types.Attribute) bool,
) ([]types.Attribute, *query.PageResponse, error) {
	limit, offset, countTotal, err := pageParams(pageReq)
	if err != nil {
		return nil, nil, err
	}
	rng := rngFn(nil)

	// start from the key returned as NextKey by the previous page.
	// The factory rebuilds the same end bound, so only the start changes.
	if pageReq != nil && len(pageReq.Key) > 0 {
		_, startKey, decErr := types.AttrTripleKey.Decode(pageReq.Key)
		if decErr != nil {
			return nil, nil, status.Errorf(codes.InvalidArgument, "invalid pagination key: %v", decErr)
		}
		if rng = rngFn(&startKey); rng == nil {
			return nil, nil, status.Error(codes.InvalidArgument, "pagination key is outside the requested range")
		}
	}

	var (
		attrs   []types.Attribute
		total   uint64
		skipped uint64
		nextKey []byte
	)

	if err := col.Walk(ctx, rng, func(key types.AttrTriple, attr types.Attribute) (bool, error) {
		if !accept(attr) {
			return false, nil
		}
		total++
		if skipped < offset {
			skipped++
			return false, nil
		}
		if uint64(len(attrs)) < limit {
			attrs = append(attrs, attr)
			return false, nil
		}
		if nextKey == nil {
			buf := make([]byte, types.AttrTripleKey.Size(key))
			if n, encErr := types.AttrTripleKey.Encode(buf, key); encErr == nil {
				nextKey = buf[:n]
			} else {
				ctx.Logger().Error("attribute: failed to encode next page key", "error", encErr)
			}
		}
		if !countTotal {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return nil, nil, err
	}

	pageRes := &query.PageResponse{NextKey: nextKey}
	if countTotal {
		pageRes.Total = total
	}
	return attrs, pageRes, nil
}

// nameAddrPageWalk is the equivalent helper for the nameAddrCounts map.
func nameAddrPageWalk(
	ctx sdk.Context,
	col collections.Map[types.NameAddrPair, uint64],
	rngFn napRange,
	pageReq *query.PageRequest,
	accept func(types.NameAddrPair) bool,
) ([]string, *query.PageResponse, error) {
	limit, offset, countTotal, err := pageParams(pageReq)
	if err != nil {
		return nil, nil, err
	}

	rng := rngFn(nil)
	if pageReq != nil && len(pageReq.Key) > 0 {
		_, startKey, decErr := types.NameAddrPairKey.Decode(pageReq.Key)
		if decErr != nil {
			return nil, nil, status.Errorf(codes.InvalidArgument, "invalid pagination key: %v", decErr)
		}
		if rng = rngFn(&startKey); rng == nil {
			return nil, nil, status.Error(codes.InvalidArgument, "pagination key is outside the requested range")
		}
	}

	var (
		accounts []string
		total    uint64
		skipped  uint64
		nextKey  []byte
	)

	if err := col.Walk(ctx, rng, func(key types.NameAddrPair, _ uint64) (bool, error) {
		if !accept(key) {
			return false, nil
		}
		total++
		if skipped < offset {
			skipped++
			return false, nil
		}
		if uint64(len(accounts)) < limit {
			accounts = append(accounts, sdk.AccAddress(key.AddrBytes).String())
			return false, nil
		}
		if nextKey == nil {
			buf := make([]byte, types.NameAddrPairKey.Size(key))
			if n, encErr := types.NameAddrPairKey.Encode(buf, key); encErr == nil {
				nextKey = buf[:n]
			} else {
				ctx.Logger().Error("attribute: failed to encode next page key", "error", encErr)
			}
		}
		if !countTotal {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return nil, nil, err
	}

	pageRes := &query.PageResponse{NextKey: nextKey}
	if countTotal {
		pageRes.Total = total
	}
	return accounts, pageRes, nil
}
