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
	keyMaxSupply              = "MaxSupply"
	keyEnableGovernance       = "EnableGovernance"
	keyUnrestrictedDenomRegex = "UnrestrictedDenomRegex"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(_ *rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		simulation.NewSimLegacyParamChange(types.ModuleName, keyMaxSupply,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMaxSupply(r).String())
			},
		),
		simulation.NewSimLegacyParamChange(types.ModuleName, keyEnableGovernance,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", GenEnableGovernance(r))
			},
		),

		simulation.NewSimLegacyParamChange(types.ModuleName, keyUnrestrictedDenomRegex,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenUnrestrictedDenomRegex(r))
			},
		),
	}
}
