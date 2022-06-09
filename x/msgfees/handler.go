package msgfees

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/provenance-io/provenance/x/msgfees/keeper"
	"github.com/provenance-io/provenance/x/msgfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewProposalHandler(k keeper.Keeper, registry cdctypes.InterfaceRegistry) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddMsgFeeProposal:
			return keeper.HandleAddMsgFeeProposal(ctx, k, c, registry)
		case *types.UpdateMsgFeeProposal:
			return keeper.HandleUpdateMsgFeeProposal(ctx, k, c, registry)
		case *types.RemoveMsgFeeProposal:
			return keeper.HandleRemoveMsgFeeProposal(ctx, k, c, registry)
		case *types.UpdateNhashPerUsdMilProposal:
			return keeper.HandleUpdateNhashPerUsdMilProposal(ctx, k, c, registry)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized marker proposal content type: %T", c)
		}
	}
}
