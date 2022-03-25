package reward

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/reward/keeper"
)

// EndBlocker called every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Info(fmt.Sprintf("In endblocker"))

	// check if epoch has ended
	ctx.Logger().Info(fmt.Sprintf("Size of events is %d", len(ctx.EventManager().GetABCIEventHistory())))
	for _, s := range ctx.EventManager().GetABCIEventHistory() {
		ctx.Logger().Info(fmt.Sprintf("events type is %s", s.Type))
		for _, y := range s.Attributes {
			ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", s.Type, y.Key, y.Value))
			//4:24PM INF events type is coin_spent
			//4:24PM INF event attribute is coin_spent attribute_key:attribute_value  spender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
			//4:24PM INF event attribute is coin_spent attribute_key:attribute_value  amount:76200000000000nhash
			//4:24PM INF events type is coin_received
			//4:24PM INF event attribute is coin_received attribute_key:attribute_value  receiver:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
			//4:24PM INF event attribute is coin_received attribute_key:attribute_value  amount:76200000000000nhash
			//4:24PM INF events type is transfer
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  recipient:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  sender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  amount:76200000000000nhash

		}
	}
}
