package keeper

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// HandleAddMsgFeeProposal handles an Add msg fees governance proposal request
func HandleAddMsgFeeProposal(ctx sdk.Context, k Keeper, proposal *types.AddMsgFeeProposal, registry codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	err := checkMsgTypeValid(registry, proposal.MsgTypeUrl)
	if err != nil {
		return fmt.Errorf("message type is not a sdk message: %v", proposal.MsgTypeUrl)
	}

	existing, err := k.GetMsgFee(ctx, proposal.MsgTypeUrl)
	if err != nil {
		return err
	}
	if existing != nil {
		return types.ErrMsgFeeAlreadyExists
	}

	msgFees := types.NewMsgFee(proposal.MsgTypeUrl, proposal.AdditionalFee)

	err = k.SetMsgFee(ctx, msgFees)
	if err != nil {
		return types.ErrInvalidFeeProposal
	}

	return nil
}

func checkMsgTypeValid(registry codectypes.InterfaceRegistry, msgTypeURL string) error {
	msgFee, err := registry.Resolve(msgTypeURL)
	if err != nil {
		return err
	}

	_, ok := msgFee.(sdk.Msg)
	if !ok {
		return fmt.Errorf("message type is not a sdk message: %v", msgTypeURL)
	}
	return err
}

// HandleUpdateMsgFeeProposal handles an Update of an existing msg fees governance proposal request
func HandleUpdateMsgFeeProposal(ctx sdk.Context, k Keeper, proposal *types.UpdateMsgFeeProposal, registry codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	err := checkMsgTypeValid(registry, proposal.MsgTypeUrl)
	if err != nil {
		return fmt.Errorf("message type is not a sdk message: %v", proposal.MsgTypeUrl)
	}
	existing, err := k.GetMsgFee(ctx, proposal.MsgTypeUrl)
	if err != nil {
		return err
	}
	if existing == nil {
		return types.ErrMsgFeeDoesNotExist
	}

	msgFees := types.NewMsgFee(proposal.MsgTypeUrl, proposal.AdditionalFee)

	err = k.SetMsgFee(ctx, msgFees)
	if err != nil {
		return types.ErrInvalidFeeProposal
	}

	return nil
}

// HandleRemoveMsgFeeProposal handles a Remove of an existing msg fees governance proposal request
func HandleRemoveMsgFeeProposal(ctx sdk.Context, k Keeper, proposal *types.RemoveMsgFeeProposal, registry codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	if err := checkMsgTypeValid(registry, proposal.MsgTypeUrl); err != nil {
		return err
	}
	existing, err := k.GetMsgFee(ctx, proposal.MsgTypeUrl)
	if err != nil {
		return err
	}
	if existing == nil {
		return types.ErrMsgFeeDoesNotExist
	}

	return k.RemoveMsgFee(ctx, proposal.MsgTypeUrl)
}

// HandleUpdateNhashPerUsdMilProposal handles update of nhash per usd mil governance proposal request
func HandleUpdateNhashPerUsdMilProposal(ctx sdk.Context, k Keeper, proposal *types.UpdateNhashPerUsdMilProposal, registry codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	params := k.GetParams(ctx)
	params.NhashPerUsdMil = proposal.NhashPerUsdMil
	k.SetParams(ctx, params)
	return nil
}
