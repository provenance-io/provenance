package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/oracle/types"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Sequence: k.GetLastQueryPacketSeq(ctx),
		Params:   k.GetParams(ctx),
		PortId:   k.GetPort(ctx),
	}
}

// InitGenesis new trigger genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetPort(ctx, genState.PortId)
	if !k.IsBound(ctx, genState.PortId) {
		err := k.BindPort(ctx, genState.PortId)
		if err != nil {
			panic("could not claim port capability: " + err.Error())
		}
	}

	k.SetParams(ctx, genState.Params)
	k.SetLastQueryPacketSeq(ctx, genState.Sequence)
}
