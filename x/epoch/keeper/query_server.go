package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/epoch/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) CurrentEpoch(goCtx context.Context, request *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	return nil, nil
}

func (k Keeper) EpochInfos(goCtx context.Context, request *types.QueryEpochsInfoRequest) (*types.QueryEpochsInfoResponse, error) {
	return nil, nil
}
