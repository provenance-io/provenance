package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// HandleCreateRootNameProposal is a handler for executing a passed create root name proposal
func HandleCreateRootNameProposal(ctx sdk.Context, k Keeper, p *types.CreateRootNameProposal) error {
	// err is suppressed because it returns an error on not found.  TODO - Remove use of error for not found
	existing, _ := k.GetRecordByName(ctx, p.Name)
	if existing != nil {
		return types.ErrNameAlreadyBound
	}
	addr, err := sdk.AccAddressFromBech32(p.Owner)
	if err != nil {
		return err
	}
	logger := k.Logger(ctx)

	// Because the proposal can contain a full domain we need to ensure all intermediate pieces are create correctly
	name := ""
	segments := strings.Split(p.Name, ".")
	for i := len(segments) - 1; i >= 0; i-- {
		name = strings.Join([]string{segments[i], name}, ".")
		name = strings.TrimRight(name, ".")

		// Ensure there is not an existing record with this name that we might be over writing
		existing, _ = k.GetRecordByName(ctx, name)
		if existing == nil {
			if err = k.SetNameRecord(ctx, name, addr, p.Restricted); err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("create root name proposal: created %s and set the owner as %s", name, p.Owner))
		} else {
			logger.Info(fmt.Sprintf("create root name proposal: intermediate domain %s exists, skipping", name))
		}
	}

	return nil
}
