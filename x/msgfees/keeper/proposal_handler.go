package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// HandleAddMsgBasedFeesProposal handles an Add msg based fees governance proposal request
func HandleAddMsgBasedFeesProposal(ctx sdk.Context, k Keeper, c *types.AddMsgBasedFeesProposal) error {
	return nil
}

// HandleUpdateMsgBasedFeesProposal handles an Update of an existing msg based fees governance proposal request
func HandleUpdateMsgBasedFeesProposal(ctx sdk.Context, k Keeper, c *types.UpdateMsgBasedFeesProposal) error {
	return nil
}

// HandleRemoveMsgBasedFeesProposal handles an Remove of an existing msg based fees governance proposal request
func HandleRemoveMsgBasedFeesProposal(ctx sdk.Context, k Keeper, c *types.RemoveMsgBasedFeesProposal) error {
	return nil
}
