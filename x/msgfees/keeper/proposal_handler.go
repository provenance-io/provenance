package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// HandleAddMsgBasedFeesProposal handles an Add msg based fees governance proposal request
func HandleAddMsgBasedFeesProposal(ctx sdk.Context, k Keeper, proposal *types.AddMsgBasedFeesProposal) error {
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

	msgFees := types.NewMsgFees(proposal.Msg.GetTypeUrl(), proposal.MinFee, proposal.FeeRate)

	k.SetMsgBasedFee(ctx, msgFees)

	return nil
}

// HandleUpdateMsgBasedFeesProposal handles an Update of an existing msg based fees governance proposal request
func HandleUpdateMsgBasedFeesProposal(ctx sdk.Context, k Keeper, proposal *types.UpdateMsgBasedFeesProposal) error {
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

	msgFees := types.NewMsgFees(proposal.Msg.GetTypeUrl(), proposal.MinFee, proposal.FeeRate)

	k.SetMsgBasedFee(ctx, msgFees)

	return nil
}

// HandleRemoveMsgBasedFeesProposal handles an Remove of an existing msg based fees governance proposal request
func HandleRemoveMsgBasedFeesProposal(ctx sdk.Context, k Keeper, proposal *types.RemoveMsgBasedFeesProposal) error {
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
