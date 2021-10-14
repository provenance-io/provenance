package msgfees

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/msgfees/keeper"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Info("NOTICE: Inside EndBlocker##")
}
