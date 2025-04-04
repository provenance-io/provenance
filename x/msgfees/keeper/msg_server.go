package keeper

import (
	"context"

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

func (m msgServer) AssessCustomMsgFee(_ context.Context, _ *types.MsgAssessCustomMsgFeeRequest) (*types.MsgAssessCustomMsgFeeResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("tx endpoint no longer available")
}

func (m msgServer) AddMsgFeeProposal(_ context.Context, _ *types.MsgAddMsgFeeProposalRequest) (*types.MsgAddMsgFeeProposalResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("tx endpoint no longer available")
}

func (m msgServer) UpdateMsgFeeProposal(_ context.Context, _ *types.MsgUpdateMsgFeeProposalRequest) (*types.MsgUpdateMsgFeeProposalResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("tx endpoint no longer available")
}

func (m msgServer) RemoveMsgFeeProposal(_ context.Context, _ *types.MsgRemoveMsgFeeProposalRequest) (*types.MsgRemoveMsgFeeProposalResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("tx endpoint no longer available")
}

func (m msgServer) UpdateNhashPerUsdMilProposal(_ context.Context, _ *types.MsgUpdateNhashPerUsdMilProposalRequest) (*types.MsgUpdateNhashPerUsdMilProposalResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("tx endpoint no longer available")
}

func (m msgServer) UpdateConversionFeeDenomProposal(_ context.Context, _ *types.MsgUpdateConversionFeeDenomProposalRequest) (*types.MsgUpdateConversionFeeDenomProposalResponse, error) {
	return nil, sdkerrors.ErrInvalidRequest.Wrap("tx endpoint no longer available")
}
