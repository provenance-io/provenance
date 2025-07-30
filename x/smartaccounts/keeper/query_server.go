package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

var _ types.QueryServer = Keeper{}

func (keeper Keeper) SmartAccount(ctx context.Context, request *types.SmartAccountQueryRequest) (*types.SmartAccountResponse, error) {
	address, err := keeper.addressCodec.StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}
	account, err := keeper.LookupAccountByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	return &types.SmartAccountResponse{
		Provenanceaccount: &account,
	}, nil
}

func (keeper Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	p, err := keeper.SmartAccountParams.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: &p}, nil
}
