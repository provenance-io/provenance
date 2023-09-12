package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/provenance-io/provenance/x/marker/types"
)

const (
	keyMaxTotalSupply         = "MaxTotalSupply"
	keyMaxSupply              = "MaxSupply"
	keyEnableGovernance       = "EnableGovernance"
	keyUnrestrictedDenomRegex = "UnrestrictedDenomRegex"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(_ *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyMaxSupply,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMaxSupply(r).String())
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyEnableGovernance,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", GenEnableGovernance(r))
			},
		),

		simulation.NewSimParamChange(types.ModuleName, keyUnrestrictedDenomRegex,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenUnrestrictedDenomRegex(r))
			},
		),
	}
}
