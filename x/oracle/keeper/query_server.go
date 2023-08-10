package keeper

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/oracle/types"
)

var _ types.QueryServer = Keeper{}

// QueryAddress returns the address of the module's oracle
func (k Keeper) OracleAddress(goCtx context.Context, req *types.QueryOracleAddressRequest) (*types.QueryOracleAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	oracle, _ := k.GetOracle(ctx)

	return &types.QueryOracleAddressResponse{Address: oracle.String()}, nil
}

// Oracle queries module's oracle
func (k Keeper) Oracle(goCtx context.Context, req *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if err := req.Query.ValidateBasic(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid query data")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.GetOracle(ctx)
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
