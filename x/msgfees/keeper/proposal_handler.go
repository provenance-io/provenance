package keeper

import (
	"fmt"
	"strconv"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// TODO[1760]: marker: Migrate the legacy gov proposals.

// HandleAddMsgFeeProposal handles an Add msg fees governance proposal request
func HandleAddMsgFeeProposal(ctx sdk.Context, k Keeper, proposal *types.AddMsgFeeProposal, _ codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}

	existing, err := k.GetMsgFee(ctx, proposal.MsgTypeUrl)
	if err != nil {
		return err
	}
	if existing != nil {
		return types.ErrMsgFeeAlreadyExists
	}
	bips, err := DetermineBips(proposal.Recipient, proposal.RecipientBasisPoints)
	if err != nil {
		return err
	}

	msgFees := types.NewMsgFee(proposal.MsgTypeUrl, proposal.AdditionalFee, proposal.Recipient, bips)

	err = k.SetMsgFee(ctx, msgFees)
	if err != nil {
		return types.ErrInvalidFeeProposal
	}

	return nil
}

// DetermineBips converts basis point string to uint32
func DetermineBips(recipient string, recipientBasisPoints string) (uint32, error) {
	var bips uint32
	if len(recipientBasisPoints) > 0 && len(recipient) > 0 {
		bips64, err := strconv.ParseUint(recipientBasisPoints, 10, 32)
		if err != nil {
			return bips, types.ErrInvalidBipsValue.Wrap(err.Error())
		}
		bips = uint32(bips64)
		if bips > 10_000 {
			return 0, types.ErrInvalidBipsValue.Wrap(fmt.Errorf("recipient basis points can only be between 0 and 10,000 : %v", recipientBasisPoints).Error())
		}
	} else if len(recipientBasisPoints) == 0 && len(recipient) > 0 {
		bips = types.DefaultMsgFeeBips
	}
	return bips, nil
}

// HandleUpdateMsgFeeProposal handles an Update of an existing msg fees governance proposal request
func HandleUpdateMsgFeeProposal(ctx sdk.Context, k Keeper, proposal *types.UpdateMsgFeeProposal, _ codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	existing, err := k.GetMsgFee(ctx, proposal.MsgTypeUrl)
	if err != nil {
		return err
	}
	if existing == nil {
		return types.ErrMsgFeeDoesNotExist
	}
	bips, err := DetermineBips(proposal.Recipient, proposal.RecipientBasisPoints)
	if err != nil {
		return err
	}

	msgFees := types.NewMsgFee(proposal.MsgTypeUrl, proposal.AdditionalFee, proposal.Recipient, bips)

	err = k.SetMsgFee(ctx, msgFees)
	if err != nil {
		return types.ErrInvalidFeeProposal
	}

	return nil
}

// HandleRemoveMsgFeeProposal handles a Remove of an existing msg fees governance proposal request
func HandleRemoveMsgFeeProposal(ctx sdk.Context, k Keeper, proposal *types.RemoveMsgFeeProposal, _ codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
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
