package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/epoch/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) CurrentEpoch(goCtx context.Context, request *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	info := k.GetEpochInfo(ctx, request.Identifier)
	return &types.QueryCurrentEpochResponse{
		CurrentEpoch: info.CurrentEpoch,
	}, nil
}

func (k Keeper) EpochInfos(goCtx context.Context, request *types.QueryEpochInfosRequest) (*types.QueryEpochInfosResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryEpochInfosResponse{
		Epochs: k.AllEpochInfos(ctx),
	}, nil
}
