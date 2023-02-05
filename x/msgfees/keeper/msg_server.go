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

// NewMsgServerImpl returns an implementation of the msgfees MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AddMsgFeeProposal(goCtx context.Context, req *types.MsgAddMsgFeeProposalRequest) (*types.MsgAddMsgFeeProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	prop := types.AddMsgFeeProposal{
		Title:                "AddMsgFeeProposal",
		Description:          "AddMsgFeeProposal",
		MsgTypeUrl:           req.MsgTypeUrl,
		AdditionalFee:        req.AdditionalFee,
		Recipient:            req.Recipient,
		RecipientBasisPoints: req.RecipientBasisPoints,
	}

	err := HandleAddMsgFeeProposal(ctx, m.Keeper, &prop, m.registry)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddMsgFeeProposalResponse{}, nil
}

func (m msgServer) UpdateMsgFeeProposal(goCtx context.Context, req *types.MsgUpdateMsgFeeProposalRequest) (*types.MsgUpdateMsgFeeProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	prop := types.UpdateMsgFeeProposal{
		Title:                "UpdateMsgFeeProposal",
		Description:          "UpdateMsgFeeProposal",
		MsgTypeUrl:           req.MsgTypeUrl,
		AdditionalFee:        req.AdditionalFee,
		Recipient:            req.Recipient,
		RecipientBasisPoints: req.RecipientBasisPoints,
	}

	err := HandleUpdateMsgFeeProposal(ctx, m.Keeper, &prop, m.registry)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateMsgFeeProposalResponse{}, nil
}

func (m msgServer) RemoveMsgFeeProposal(goCtx context.Context, req *types.MsgRemoveMsgFeeProposalRequest) (*types.MsgRemoveMsgFeeProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	prop := types.RemoveMsgFeeProposal{
		Title:       "RemoveMsgFeeProposal",
		Description: "RemoveMsgFeeProposal",
		MsgTypeUrl:  req.MsgTypeUrl,
	}

	err := HandleRemoveMsgFeeProposal(ctx, m.Keeper, &prop, m.registry)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveMsgFeeProposalResponse{}, nil
}

func (m msgServer) UpdateNhashPerUsdMilProposal(goCtx context.Context, req *types.MsgUpdateNhashPerUsdMilProposalRequest) (*types.MsgUpdateNhashPerUsdMilProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	prop := types.UpdateNhashPerUsdMilProposal{
		Title:          "UpdateNhashPerUsdMilProposal",
		Description:    "UpdateNhashPerUsdMilProposal",
		NhashPerUsdMil: req.NhashPerUsdMil,
	}

	err := HandleUpdateNhashPerUsdMilProposal(ctx, m.Keeper, &prop)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateNhashPerUsdMilProposalResponse{}, nil
}

func (m msgServer) UpdateConversionFeeDenomProposal(goCtx context.Context, req *types.MsgUpdateConversionFeeDenomProposalRequest) (*types.MsgUpdateConversionFeeDenomProposalResponse, error) {
	if m.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", m.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	prop := types.UpdateConversionFeeDenomProposal{
		Title:              "UpdateConversionFeeDenomProposal",
		Description:        "UpdateConversionFeeDenomProposal",
		ConversionFeeDenom: req.ConversionFeeDenom,
	}

	err := HandleUpdateConversionFeeDenomProposal(ctx, m.Keeper, &prop)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateConversionFeeDenomProposalResponse{}, nil
}
