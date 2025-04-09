package keeper

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(_ context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("query no longer available")
}

func (k Keeper) QueryAllMsgFees(_ context.Context, _ *types.QueryAllMsgFeesRequest) (*types.QueryAllMsgFeesResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("query no longer available")
}

func (k Keeper) CalculateTxFees(_ context.Context, _ *types.CalculateTxFeesRequest) (*types.CalculateTxFeesResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("query no longer available")
}
