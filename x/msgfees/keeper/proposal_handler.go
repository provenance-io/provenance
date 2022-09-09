package keeper

import (
	"fmt"
	"strconv"

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
			return 0, fmt.Errorf("recipient basis points can only be between 0 and 10,000 : %v", recipientBasisPoints)
		}
	} else if len(recipientBasisPoints) == 0 && len(recipient) > 0 {
		bips = types.DefaultMsgFeeBips
	}
	return bips, nil
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
