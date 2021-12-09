package keeper

import (
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

// HandleAddMsgBasedFeeProposal handles an Add msg based fees governance proposal request
func HandleAddMsgBasedFeeProposal(ctx sdk.Context, k Keeper, proposal *types.AddMsgBasedFeeProposal, registry codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	err := checkMsgTypeValid(registry, proposal.MsgTypeURL)
	if err != nil {
		return fmt.Errorf("message type is not a sdk message: %v", proposal.MsgTypeURL)
	}

	existing, err := k.GetMsgBasedFee(ctx, proposal.MsgTypeURL)
	if err != nil {
		return err
	}
	if existing != nil {
		return types.ErrMsgFeeAlreadyExists
	}

	msgFees := types.NewMsgBasedFee(proposal.MsgTypeURL, proposal.AdditionalFee)

	err = k.SetMsgBasedFee(ctx, msgFees)
	if err != nil {
		return types.ErrInvalidFeeProposal
	}

	return nil
}

func checkMsgTypeValid(registry codectypes.InterfaceRegistry, msgTypeUrl string) (error) {
	msgFee, err := registry.Resolve(msgTypeUrl)
	if err != nil {
		return err
	}

	_, ok := msgFee.(sdk.Msg)
	if !ok {
		return fmt.Errorf("message type is not a sdk message: %v", msgTypeUrl)
	}
	return err
}

// HandleUpdateMsgBasedFeeProposal handles an Update of an existing msg based fees governance proposal request
func HandleUpdateMsgBasedFeeProposal(ctx sdk.Context, k Keeper, proposal *types.UpdateMsgBasedFeeProposal, registry codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	err := checkMsgTypeValid(registry, proposal.MsgTypeURL)
	if err != nil {
		return fmt.Errorf("message type is not a sdk message: %v", proposal.MsgTypeURL)
	}
	existing, err := k.GetMsgBasedFee(ctx, proposal.MsgTypeURL)
	if err != nil {
		return err
	}
	if existing == nil {
		return types.ErrMsgFeeDoesNotExist
	}

	msgFees := types.NewMsgBasedFee(proposal.MsgTypeURL, proposal.AdditionalFee)

	err = k.SetMsgBasedFee(ctx, msgFees)
	if err != nil {
		return types.ErrInvalidFeeProposal
	}

	return nil
}

// HandleRemoveMsgBasedFeeProposal handles an Remove of an existing msg based fees governance proposal request
func HandleRemoveMsgBasedFeeProposal(ctx sdk.Context, k Keeper, proposal *types.RemoveMsgBasedFeeProposal, registry codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	err := checkMsgTypeValid(registry, proposal.MsgTypeURL)
	existing, err := k.GetMsgBasedFee(ctx, proposal.MsgTypeURL)
	if err != nil {
		return err
	}
	if existing == nil {
		return types.ErrMsgFeeDoesNotExist
	}

	return k.RemoveMsgBasedFee(ctx, proposal.MsgTypeURL)
}
