package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

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
func (k Keeper) GetAllEscrow(goCtx context.Context, req *escrow.GetAllEscrowRequest) (*escrow.GetAllEscrowResponse, error) {
	var pagination *query.PageRequest
	if req != nil {
		pagination = req.Pagination
	}

	var err error
	resp := &escrow.GetAllEscrowResponse{}
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := k.getAllEscrowCoinPrefixStore(ctx)
	// TODO[1607]: Fix this so that the count is by address instead of entry.
	resp.Pagination, err = query.Paginate(
		store, pagination,
		func(key []byte, value []byte) error {
			amount, ierr := UnmarshalEscrowCoinValue(value)
			if ierr != nil {
				return ierr
			}
			addr, denom := ParseEscrowCoinKeyUnprefixed(key)
			// TODO[1607]: Fix this so that each entry is combined by address.
			resp.Escrows = append(resp.Escrows, &escrow.AccountEscrow{
				Address: addr.String(),
				Amount:  sdk.Coins{sdk.NewCoin(denom, amount)},
			})
			return nil
		},
	)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
