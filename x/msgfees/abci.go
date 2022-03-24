package msgfees

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/msgfees/keeper"
)

// EndBlocker called every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Info(fmt.Sprintf("In endblocker for msgfees"))

	for i, s := range ctx.EventManager().GetABCIEventHistory() {
		ctx.Logger().Info(fmt.Sprintf("events in end blocker %d", i))
		ctx.Logger().Info(fmt.Sprintf("events attributes are %s", s.Type))
		for _, y := range s.Attributes {
			ctx.Logger().Info(fmt.Sprintf("event attribute key:events attribute value  %s:%s", y.Key, y.Value))
		}
	}
}
