package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/flatfees/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// Simulation parameter constants
const (
	FloorGasPrice = "floor_gas_price"
)

// FloorMinGasPrice randomized FloorGasPrice
func FloorMinGasPrice(r *rand.Rand) uint64 {
	return r.Uint64()
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	var floorGasPrice uint64
	simState.AppParams.GetOrGenerate(
		FloorGasPrice, &floorGasPrice, simState.Rand,
		func(r *rand.Rand) { floorGasPrice = FloorMinGasPrice(r) },
	)

	genState := types.GenesisState{
		Params: types.Params{
			// TODO[fees]: Better randomized params.
		},
		MsgFees: []*types.MsgFee{
			// Adding fees for create marker with asking for a large number of stake to make sure that
			// the call is failed without the additional fee provided.
			{
				MsgTypeUrl: sdk.MsgTypeURL(&markertypes.MsgAddMarkerRequest{}),
				Cost:       sdk.Coins{sdk.NewInt64Coin("stake", 100_000_000_000_000)},
			},
		},
	}

	bz, err := json.MarshalIndent(&genState, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated flatfees parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&genState)
}
