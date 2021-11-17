package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

// HandleAddMsgBasedFeeProposal handles an Add msg based fees governance proposal request
func HandleAddMsgBasedFeeProposal(ctx sdk.Context, k Keeper, proposal *types.AddMsgBasedFeeProposal) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	existing, err := k.GetMsgBasedFee(ctx, proposal.Msg.GetTypeUrl())
	if err != nil {
		return err
	}
	if existing != nil {
		return types.ErrMsgFeeAlreadyExists
	}

	msgFees := types.NewMsgBasedFee(proposal.Msg.GetTypeUrl(), proposal.AdditionalFee)

	err = k.SetMsgBasedFee(ctx, msgFees)
	if err != nil {
		return types.ErrInvalidFeeProposal
	}

	return nil
}

// HandleUpdateMsgBasedFeeProposal handles an Update of an existing msg based fees governance proposal request
func HandleUpdateMsgBasedFeeProposal(ctx sdk.Context, k Keeper, proposal *types.UpdateMsgBasedFeeProposal) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	existing, err := k.GetMsgBasedFee(ctx, proposal.Msg.GetTypeUrl())
	if err != nil {
		return err
	}
	if existing == nil {
		return types.ErrMsgFeeDoesNotExist
	}

	msgFees := types.NewMsgBasedFee(proposal.Msg.GetTypeUrl(), proposal.AdditionalFee)

	err = k.SetMsgBasedFee(ctx, msgFees)
	if err != nil {
		return types.ErrInvalidFeeProposal
	}

	return nil
}

// HandleRemoveMsgBasedFeeProposal handles an Remove of an existing msg based fees governance proposal request
func HandleRemoveMsgBasedFeeProposal(ctx sdk.Context, k Keeper, proposal *types.RemoveMsgBasedFeeProposal) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	existing, err := k.GetMsgBasedFee(ctx, proposal.Msg.GetTypeUrl())
	if err != nil {
		return err
	}
	if existing == nil {
		return types.ErrMsgFeeDoesNotExist
	}

	return k.RemoveMsgBasedFee(ctx, proposal.Msg.GetTypeUrl())
}
