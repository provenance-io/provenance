package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/provenance-io/provenance/x/msgfees/types"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const (
	keyMinGasPrice      = "MinGasPrice"
	keyEnableGovernance = "EnableGovernance"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyMinGasPrice,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMinGasPrice(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyEnableGovernance,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", GenEnableGovernance(r))
			},
		),
	}
}
