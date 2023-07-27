package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/oracle/types"
)

var _ types.QueryServer = Keeper{}

// QueryAddress returns the address of the oracle's contract
func (k Keeper) ContractAddress(ctx context.Context, req *types.QueryContractAddressRequest) (*types.QueryContractAddressResponse, error) {
	return &types.QueryContractAddressResponse{}, nil
}

// Oracle sends an ICQ to an oracle
func (k Keeper) Oracle(ctx context.Context, req *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	return &types.QueryOracleResponse{}, nil
}

func (k Keeper) OracleContract(ctx context.Context, req *types.QueryOracleContractRequest) (*types.QueryOracleContractResponse, error) {
	return &types.QueryOracleContractResponse{}, nil
}

func (k Keeper) OracleResult(ctx context.Context, req *types.QueryOracleResult) (*types.QueryOracleResultResponse, error) {
	return &types.QueryOracleResultResponse{}, nil
}
