package reward

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/provenance-io/provenance/x/reward/keeper"
	"github.com/provenance-io/provenance/x/reward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewProposalHandler(k keeper.Keeper, registry cdctypes.InterfaceRegistry) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddRewardProgramProposal:
			return keeper.HandleAddRewardProgramProposal(ctx, k, c, registry)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized reward proposal content type: %T", c)
		}
	}
}
