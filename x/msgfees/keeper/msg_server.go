package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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

func (m msgServer) AssessCustomMsgFee(goCtx context.Context, req *types.MsgAssessCustomMsgFeeRequest) (*types.MsgAssessCustomMsgFeeResponse, error) {
	// method only emits that the event has been submitted, all logic is handled in the provenance custom msg handlers
	ctx := sdk.UnwrapSDKContext(goCtx)

	// if there is a recipient and bips are not set, we will want to emit the default bips with event
	recipientBips := req.RecipientBasisPoints
	if len(req.Recipient) > 0 && len(req.RecipientBasisPoints) == 0 {
		recipientBips = fmt.Sprintf("%v", types.AssessCustomMsgFeeBips)
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAssessCustomMsgFee,
			sdk.NewAttribute(types.KeyAttributeName, req.Name),
			sdk.NewAttribute(types.KeyAttributeAmount, req.Amount.String()),
			sdk.NewAttribute(types.KeyAttributeRecipient, req.Recipient),
			sdk.NewAttribute(types.KeyAttributeBips, recipientBips),
		),
	)
	return &types.MsgAssessCustomMsgFeeResponse{}, nil
}

func (m msgServer) AddMsgFeeProposal(goCtx context.Context, req *types.MsgAddMsgFeeProposalRequest) (*types.MsgAddMsgFeeProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	err := m.Keeper.AddMsgFee(sdk.UnwrapSDKContext(goCtx), req.MsgTypeUrl, req.Recipient, req.RecipientBasisPoints, req.AdditionalFee)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddMsgFeeProposalResponse{}, nil
}

func (m msgServer) UpdateMsgFeeProposal(goCtx context.Context, req *types.MsgUpdateMsgFeeProposalRequest) (*types.MsgUpdateMsgFeeProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	err := m.Keeper.UpdateMsgFee(sdk.UnwrapSDKContext(goCtx), req.MsgTypeUrl, req.Recipient, req.RecipientBasisPoints, req.AdditionalFee)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateMsgFeeProposalResponse{}, nil
}

func (m msgServer) RemoveMsgFeeProposal(goCtx context.Context, req *types.MsgRemoveMsgFeeProposalRequest) (*types.MsgRemoveMsgFeeProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	err := m.Keeper.RemoveMsgFee(sdk.UnwrapSDKContext(goCtx), req.MsgTypeUrl)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveMsgFeeProposalResponse{}, nil
}

func (m msgServer) UpdateNhashPerUsdMilProposal(goCtx context.Context, req *types.MsgUpdateNhashPerUsdMilProposalRequest) (*types.MsgUpdateNhashPerUsdMilProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	m.Keeper.UpdateNhashPerUsdMilParam(sdk.UnwrapSDKContext(goCtx), req.NhashPerUsdMil)

	return &types.MsgUpdateNhashPerUsdMilProposalResponse{}, nil
}

func (m msgServer) UpdateConversionFeeDenomProposal(goCtx context.Context, req *types.MsgUpdateConversionFeeDenomProposalRequest) (*types.MsgUpdateConversionFeeDenomProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	m.Keeper.UpdateConversionFeeDenomParam(sdk.UnwrapSDKContext(goCtx), req.ConversionFeeDenom)

	return &types.MsgUpdateConversionFeeDenomProposalResponse{}, nil
}
