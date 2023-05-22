package simulation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/provenance-io/provenance/x/trigger/keeper"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	_ simtypes.AppParams, _ codec.JSONCodec, _ keeper.Keeper,
) simulation.WeightedOperations {
	return simulation.WeightedOperations{}
}
