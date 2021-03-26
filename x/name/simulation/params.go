package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/x/name/types"
)

const (
	keyMinSegmentLength  = "MinSegmentLength"
	keyMaxSegmentLength  = "MaxSegmentLength"
	keyMaxNameLevels     = "MaxNameLevels"
	keyAllowUnrestricted = "AllowUnrestrictedNames"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyMinSegmentLength,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMinSegmentLength(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMaxSegmentLength,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxSegmentLength(r))
			},
		),

		simulation.NewSimParamChange(types.ModuleName, keyMaxNameLevels,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxNameLevels(r))
			},
		),

		simulation.NewSimParamChange(types.ModuleName, keyAllowUnrestricted,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", GenAllowUnrestrictedNames(r))
			},
		),
	}
}
