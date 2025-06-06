package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/flatfees/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServer returns an implementation of the x/flatfees QueryServer interface for the provided keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Params queries the parameters for x/flatfees.
func (k queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	c := sdk.UnwrapSDKContext(ctx)
	return &types.QueryParamsResponse{Params: k.GetParams(c)}, nil
}

// AllMsgFees returns info on all msg types that have a customized msg fee.
func (k queryServer) AllMsgFees(ctx context.Context, req *types.QueryAllMsgFeesRequest) (*types.QueryAllMsgFeesResponse, error) {
	convert := true
	var pageReq *query.PageRequest
	if req != nil {
		convert = !req.DoNotConvert
		pageReq = req.Pagination
	}

	var cf *types.ConversionFactor
	if convert {
		params := k.GetParams(sdk.UnwrapSDKContext(ctx))
		cf = &params.ConversionFactor
	}

	rv := &types.QueryAllMsgFeesResponse{}
	var err error
	rv.MsgFees, rv.Pagination, err = query.CollectionPaginate(ctx, k.msgFees, pageReq, func(_ string, msgFee types.MsgFee) (*types.MsgFee, error) {
		if convert {
			return cf.ConvertMsgFee(&msgFee), nil
		}
		return &msgFee, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return rv, nil
}

// MsgFee will return information about what it will cost to execute a given msg type.
// If the provided msg type does not have a specific fee defined, the default is returned.
func (k queryServer) MsgFee(ctx context.Context, req *types.QueryMsgFeeRequest) (*types.QueryMsgFeeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msgFee, err := k.GetMsgFee(sdkCtx, req.MsgTypeUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if msgFee != nil && req.DoNotConvert {
		return &types.QueryMsgFeeResponse{MsgFee: msgFee}, nil
	}

	// If we don't have an entry for it, make sure it's a valid msg type url.
	if msgFee == nil {
		if _, err = k.cdc.InterfaceRegistry().Resolve(req.MsgTypeUrl); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown msg type url %q", req.MsgTypeUrl)
		}
	}

	params := k.GetParams(sdkCtx)
	if msgFee == nil {
		// If we don't have an entry, we know it's an actual msg type, so return the default cost.
		msgFee = &types.MsgFee{
			MsgTypeUrl: req.MsgTypeUrl,
			Cost:       params.DefaultCostCoins(),
		}
	}

	if !req.DoNotConvert {
		msgFee = params.ConversionFactor.ConvertMsgFee(msgFee)
	}

	return &types.QueryMsgFeeResponse{MsgFee: msgFee}, nil
}
