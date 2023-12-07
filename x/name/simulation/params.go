package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
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
func ParamChanges(_ *rand.Rand) []proposal.ParamChange {
	return []proposal.ParamChange{
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
