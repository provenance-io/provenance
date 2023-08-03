package keeper

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/oracle/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// QueryAddress returns the address of the oracle's contract
func (k Keeper) OracleAddress(goCtx context.Context, req *types.QueryOracleAddressRequest) (*types.QueryOracleAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, err := k.GetOracleContract(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryOracleAddressResponse{Address: contract.String()}, nil
}

func (k Keeper) Oracle(goCtx context.Context, req *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.GetOracleContract(ctx)
	if err != nil {
		return nil, err
	}
	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   addr.String(),
		QueryData: req.Query,
	}
	resp, err := k.wasmQueryServer.SmartContractState(ctx, query)
	if err != nil {
		return nil, err
	}
	return &types.QueryOracleResponse{Data: resp.Data}, nil
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
