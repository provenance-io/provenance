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

	if proposal.AdditionalFee.Denom == k.GetDefaultFeeDenom() && (proposal.MinGasPrice.IsNil() || proposal.MinGasPrice.IsZero() || !proposal.MinGasPrice.IsValid() || proposal.GetMinGasPrice().Denom != proposal.AdditionalFee.Denom) {
		return types.ErrInvalidFeeProposal
	}

	msgFees := types.NewMsgBasedFee(proposal.Msg.GetTypeUrl(), proposal.AdditionalFee)

	k.SetMsgBasedFee(ctx, msgFees)

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

	if proposal.AdditionalFee.Denom == k.GetDefaultFeeDenom() && (proposal.MinGasPrice.IsNil() || proposal.MinGasPrice.IsZero() || !proposal.MinGasPrice.IsValid() || proposal.GetMinGasPrice().Denom != proposal.AdditionalFee.Denom) {
		return types.ErrInvalidFeeProposal
	}

	msgFees := types.NewMsgBasedFee(proposal.Msg.GetTypeUrl(), proposal.AdditionalFee)

	k.SetMsgBasedFee(ctx, msgFees)

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
