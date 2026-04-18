package keeper

import (
	"context"
	"fmt"
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

	rng, endBound := attrAddrNameRange(addrBz, nameHash)
	attrs, pageRes, err := attrPageWalk(ctx, k.attributes, rng, endBound, req.Pagination,
		func(attr types.Attribute) bool {
			return attr.ExpirationDate == nil || !blockTime.After(attr.ExpirationDate.UTC())
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

	rng, endBound := attrAddrRange(addrBz)
	attrs, pageRes, err := attrPageWalk(ctx, k.attributes, rng, endBound, req.Pagination,
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

	rng, endBound := attrAddrRange(addrBz)
	attrs, pageRes, err := attrPageWalk(ctx, k.attributes, rng, endBound, req.Pagination,
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
	var nameHash [32]byte
	copy(nameHash[:], types.GetNameKeyBytes(req.AttributeName))

	rng, endBound := nameHashRange(nameHash)
	accounts, pageRes, err := nameAddrPageWalk(ctx, k.nameAddrCounts, rng, endBound, req.Pagination)
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

// attrPageWalk walks col over rng with full pagination:
func attrPageWalk(
	ctx sdk.Context,
	col collections.Map[types.AttrTriple, types.Attribute],
	rng *collections.Range[types.AttrTriple],
	endBound *types.AttrTriple,
	pageReq *query.PageRequest,
	accept func(types.Attribute) bool,
) ([]types.Attribute, *query.PageResponse, error) {
	limit := uint64(query.DefaultLimit)
	offset := uint64(0)
	countTotal := false

	if pageReq == nil {
		pageReq = &query.PageRequest{CountTotal: true}
	}

	if pageReq != nil {
		if pageReq.Limit > 0 {
			limit = pageReq.Limit
		}
		offset = pageReq.Offset
		countTotal = pageReq.CountTotal

		// Cursor resume: start from the key returned as NextKey by the previous page.
		if len(pageReq.Key) > 0 {
			_, startKey, err := types.AttrTripleKey.Decode(pageReq.Key)
			if err != nil {
				return nil, nil, fmt.Errorf("attribute: invalid pagination key: %w", err)
			}
			newRng := new(collections.Range[types.AttrTriple]).StartInclusive(startKey)
			if endBound != nil {
				newRng = newRng.EndExclusive(*endBound)
			}
			rng = newRng
			offset = 0
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
	rng *collections.Range[types.NameAddrPair],
	endBound *types.NameAddrPair,
	pageReq *query.PageRequest,
) ([]string, *query.PageResponse, error) {
	limit := uint64(query.DefaultLimit)
	offset := uint64(0)
	countTotal := false

	if pageReq == nil {
		pageReq = &query.PageRequest{CountTotal: true}
	}

	if pageReq != nil {
		if pageReq.Limit > 0 {
			limit = pageReq.Limit
		}
		offset = pageReq.Offset
		countTotal = pageReq.CountTotal

		if len(pageReq.Key) > 0 {
			_, startKey, err := types.NameAddrPairKey.Decode(pageReq.Key)
			if err != nil {
				return nil, nil, fmt.Errorf("attribute: invalid pagination key: %w", err)
			}
			newRng := new(collections.Range[types.NameAddrPair]).StartInclusive(startKey)
			if endBound != nil {
				newRng = newRng.EndExclusive(*endBound)
			}
			rng = newRng
			offset = 0
		}
	}

	var (
		accounts []string
		total    uint64
		skipped  uint64
		nextKey  []byte
	)

	if err := col.Walk(ctx, rng, func(key types.NameAddrPair, _ uint64) (bool, error) {
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
