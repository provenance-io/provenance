package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/escrow"
)

// GetEscrow looks up the funds that are in escrow for an address.
func (k Keeper) GetEscrow(goCtx context.Context, req *escrow.GetEscrowRequest) (*escrow.GetEscrowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Address) == 0 {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &escrow.GetEscrowResponse{}
	resp.Amount, err = k.GetEscrowCoins(ctx, addr)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// GetAllEscrow returns all addresses with funds in escrow, and the amount in escrow.
func (k Keeper) GetAllEscrow(goCtx context.Context, _ *escrow.GetAllEscrowRequest) (*escrow.GetAllEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	escrows, err := k.GetAllAccountEscrows(ctx)
	if err != nil {
		return nil, err
	}

	return &escrow.GetAllEscrowResponse{Escrows: escrows}, nil
}
