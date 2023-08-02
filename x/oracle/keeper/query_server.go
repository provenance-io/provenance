package keeper

import (
	"context"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/oracle/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// QueryAddress returns the address of the oracle's contract
func (k Keeper) ContractAddress(goCtx context.Context, req *types.QueryContractAddressRequest) (*types.QueryContractAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, err := k.GetOracleContract(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryContractAddressResponse{Address: contract.String()}, nil
}

func (k Keeper) OracleContract(ctx context.Context, req *types.QueryOracleContractRequest) (*types.QueryOracleContractResponse, error) {
	// We do the logic here to actually send it to the oracle
	return &types.QueryOracleContractResponse{}, nil
}

func (k Keeper) QueryState(goCtx context.Context, req *types.QueryQueryStateRequest) (*types.QueryQueryStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	qreq, err := k.GetQueryRequest(ctx, req.Sequence)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	var anyQResp *cdctypes.Any
	qresp, err := k.GetQueryResponse(ctx, req.Sequence)
	if err == nil {
		anyQResp, err = cdctypes.NewAnyWithValue(&qresp)
		if err != nil {
			panic(err)
		}
	}

	anyQReq, err := cdctypes.NewAnyWithValue(&qreq)
	if err != nil {
		panic(err)
	}
	return &types.QueryQueryStateResponse{
		Request:  *anyQReq,
		Response: anyQResp,
	}, nil
}
