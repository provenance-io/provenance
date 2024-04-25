package keeper

import (
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
