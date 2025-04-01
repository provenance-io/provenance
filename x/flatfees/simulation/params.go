package simulation

// DONTCOVER

import (
	"encoding/json"
	"math/rand"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/provenance-io/provenance/x/flatfees/types"
)

const (
	keyFloorGasPrice = "FloorGasPrice"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(_ *rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		simulation.NewSimLegacyParamChange(types.ModuleName, keyFloorGasPrice,
			func(r *rand.Rand) string {
				jsonResp, err := json.Marshal(sdk.Coin{
					Denom:  "stake",
					Amount: sdkmath.NewIntFromUint64(FloorMinGasPrice(r)),
				})
				if err != nil {
					panic("Error happened in JSON marshal")
				}
				return string(jsonResp)
			},
		),
	}
}
