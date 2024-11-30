package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/provenance-io/provenance/x/nav"
)

type QueryServer struct {
	Keeper
}

func NewQueryServer(k Keeper) nav.QueryServer {
	return QueryServer{Keeper: k}
}

// GetNAV returns the single Net Asset Value entry requested.
func (q QueryServer) GetNAV(ctx context.Context, req *nav.QueryGetNAVRequest) (*nav.QueryGetNAVResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.AssetDenom) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty asset denom")
	}
	if len(req.PriceDenom) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty price denom")
	}
	rv := &nav.QueryGetNAVResponse{
		Nav: q.GetNAVRecord(ctx, req.AssetDenom, req.PriceDenom),
	}
	return rv, nil
}

// GetAllNAVs returns a page of all Net Asset Value entries, possibly limited
// to a single asset denom.
func (q QueryServer) GetAllNAVs(ctx context.Context, req *nav.QueryGetAllNAVsRequest) (*nav.QueryGetAllNAVsResponse, error) {
	var assetDenom string
	var pagination *query.PageRequest
	if req != nil {
		assetDenom = req.AssetDenom
		pagination = req.Pagination
	}

	var opts []func(opt *query.CollectionsPaginateOptions[collections.Pair[string, string]])
	if len(assetDenom) > 0 {
		opts = append(opts, query.WithCollectionPaginationPairPrefix[string, string](assetDenom))
	}

	rv := &nav.QueryGetAllNAVsResponse{}
	var err error
	rv.Navs, rv.Pagination, err = query.CollectionPaginate(ctx, q.navs, pagination,
		func(_ collections.Pair[string, string], value nav.NetAssetValueRecord) (*nav.NetAssetValueRecord, error) {
			return &value, nil
		},
		opts...,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return rv, nil
}
