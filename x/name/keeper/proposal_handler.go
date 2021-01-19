package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/name/types"
)

// HandleCreateRootNameProposal is a handler for executing a passed create root name proposal
func HandleCreateRootNameProposal(ctx sdk.Context, k Keeper, p *types.CreateRootNameProposal) error {
	existing, err := k.GetRecordByName(ctx, p.Name)
	if err != nil {
		return err
	}
	if existing != nil {
		return types.ErrNameAlreadyBound
	}
	addr, err := sdk.AccAddressFromBech32(p.Owner)
	if err != nil {
		return err
	}
	if err = k.setNameRecord(ctx, p.Name, addr, p.Restricted); err != nil {
		return err
	}
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("created a new root name %s and set the owner as %s", p.Name, p.Owner))
	return nil
}
