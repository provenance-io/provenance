package keeper

import (
	"context"

	flatfees "github.com/provenance-io/provenance/x/flatfees/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

type queryServer struct {
	ffq types.FlatFeesQuerier
}

func NewQueryServer(ffq types.FlatFeesQuerier) types.QueryServer {
	return &queryServer{ffq}
}

// CalculateTxFees simulates executing a transaction for estimating gas usage and fees.
func (k queryServer) CalculateTxFees(ctx context.Context, req *types.CalculateTxFeesRequest) (*types.CalculateTxFeesResponse, error) {
	req2 := &flatfees.QueryCalculateTxFeesRequest{TxBytes: req.TxBytes, GasAdjustment: req.GasAdjustment}
	resp2, err := k.ffq.CalculateTxFees(ctx, req2)
	if err != nil {
		return nil, err
	}
	return &types.CalculateTxFeesResponse{TotalFees: resp2.TotalFees, EstimatedGas: resp2.EstimatedGas}, nil
}
