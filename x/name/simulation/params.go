package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

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
func ParamChanges(_ *rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		simulation.NewSimLegacyParamChange(types.ModuleName, keyMinSegmentLength,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMinSegmentLength(r))
			},
		),
		simulation.NewSimLegacyParamChange(types.ModuleName, keyMaxSegmentLength,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxSegmentLength(r))
			},
		),

		simulation.NewSimLegacyParamChange(types.ModuleName, keyMaxNameLevels,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxNameLevels(r))
			},
		),

		simulation.NewSimLegacyParamChange(types.ModuleName, keyAllowUnrestricted,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", GenAllowUnrestrictedNames(r))
			},
		),
	}
}
