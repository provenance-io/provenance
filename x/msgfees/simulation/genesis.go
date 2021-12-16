package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/provenance-io/provenance/x/msgfees/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// Simulation parameter constants
const (
	FloorGasPrice    = "floor_gas_price"
	EnableGovernance = "enable_governance"
)

// FloorMinGasPrice randomized FloorGasPrice
func FloorMinGasPrice(r *rand.Rand) uint32 {
	return r.Uint32()
}

// GenEnableGovernance returns a randomized EnableGovernance parameter.
func GenEnableGovernance(r *rand.Rand) bool {
	return r.Int63n(101) <= 50 // 50% chance of enablement
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	var floorGasPrice uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, FloorGasPrice, &floorGasPrice, simState.Rand,
		func(r *rand.Rand) { floorGasPrice = FloorMinGasPrice(r) },
	)

	var enableGovernance bool
	simState.AppParams.GetOrGenerate(
		simState.Cdc, EnableGovernance, &enableGovernance, simState.Rand,
		func(r *rand.Rand) { enableGovernance = GenEnableGovernance(r) },
	)

	msgFeesGenesis := types.GenesisState{
		Params: types.Params{
			FloorGasPrice:    floorGasPrice,
			EnableGovernance: enableGovernance,
		},
		MsgBasedFees: []types.MsgBasedFee{},
	}

	bz, err := json.MarshalIndent(&msgFeesGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated msgfees parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&msgFeesGenesis)
}
