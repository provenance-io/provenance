package marker

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

// TODO[1760]: marker: Migrate the legacy gov proposals.

func NewProposalHandler(k keeper.Keeper) govtypesv1beta1.Handler {
	return func(ctx sdk.Context, content govtypesv1beta1.Content) error {
		switch c := content.(type) {
		case *types.SetDenomMetadataProposal:
			return keeper.HandleSetDenomMetadataProposal(ctx, k, c)
		default:
			return sdkerrors.ErrUnknownRequest.Wrapf("unrecognized marker proposal content type: %T", c)
		}
	}
}
