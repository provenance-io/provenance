package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the msgfees MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) CreateMsgBasedFee(goCtx context.Context, request *types.CreateMsgBasedFeeRequest) (*types.CreateMsgBasedFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Validate transaction message.
	err := request.ValidateBasic()

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	existing, err := k.GetMsgBasedFee(ctx, request.GetMsgFees().MsgTypeUrl)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if existing != nil {
		return nil, sdkerrors.Wrap(types.ErrMsgFeeAlreadyExists, err.Error())
	}

	k.SetMsgBasedFee(ctx, *request.MsgFees)

	return &types.CreateMsgBasedFeeResponse{
		MsgFees: request.MsgFees,
	}, nil
}

func (k msgServer) CalculateMsgBasedFees(ctx context.Context, request *types.CalculateFeePerMsgRequest) (*types.CalculateMsgBasedFeesResponse, error) {
	panic("implement me")
}
