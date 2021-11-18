package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

)

// Simulation parameter constants
const (
	MinGasPrice = "min_gas_price"
	EnableGovernance       = "enable_governance"
)

// GenMinGasPrice randomized MinGasPrice
func GenMinGasPrice(r *rand.Rand) uint32 {
	return r.Uint32()
}

// GenEnableGovernance returns a randomized EnableGovernance parameter.
func GenEnableGovernance(r *rand.Rand) bool {
	return r.Int63n(101) <= 50 // 50% chance of enablement
}



// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	var minGasPrice uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MinGasPrice, &minGasPrice, simState.Rand,
		func(r *rand.Rand) { minGasPrice = GenMinGasPrice(r) },
	)

	var enableGovernance bool
	simState.AppParams.GetOrGenerate(
		simState.Cdc, EnableGovernance, &enableGovernance, simState.Rand,
		func(r *rand.Rand) { enableGovernance = GenEnableGovernance(r) },
	)

	attributeGenesis := types.GenesisState{
		Params: types.Params{
			MinGasPrice: minGasPrice,
			EnableGovernance: enableGovernance,
		},
		MsgBasedFees: []types.MsgBasedFee{},
	}

	bz, err := json.MarshalIndent(&attributeGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated attribute parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&attributeGenesis)
}
