package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/x/attribute/types"
)

const (
	keyMaxValueLength = "MaxValueLength"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyMaxValueLength,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxValueLength(r))
			},
		),
	}
}
