package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

func (k Keeper) InitGenesis(origCtx sdk.Context, genState *exchange.GenesisState) {
	// TODO[1658]: Implement InitGenesis
	panic("not implemented")
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *exchange.GenesisState {
	// TODO[1658]: Implement ExportGenesis
	panic("not implemented")
}
