package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) SmartAccount(ctx context.Context, request *types.SmartAccountQueryRequest) (*types.SmartAccountResponse, error) {
	address, err := k.addressCodec.StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}
	account, err := k.LookupAccountByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	return &types.SmartAccountResponse{
		Provenanceaccount: &account,
	}, nil
}

func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	p, err := k.SmartAccountParams.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: &p}, nil
}
