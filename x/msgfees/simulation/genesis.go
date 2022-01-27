package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	markertypes "github.com/provenance-io/provenance/x/marker/types"

	"github.com/provenance-io/provenance/x/msgfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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
		simState.Cdc, FloorGasPrice, &floorGasPrice, simState.Rand,
		func(r *rand.Rand) { floorGasPrice = FloorMinGasPrice(r) },
	)

	msgFeesGenesis := types.GenesisState{
		Params: types.Params{
			FloorGasPrice: sdk.Coin{Amount: sdk.NewIntFromUint64(floorGasPrice), Denom: "blah"},
		},
		MsgFees: []types.MsgFee{
			// Adding fees for create marker with asking for a large number of stake to make sure that
			// the call is failed without the additional fee provided.
			types.NewMsgFee(sdk.MsgTypeURL(&markertypes.MsgAddMarkerRequest{}), sdk.NewCoin("stake", sdk.NewInt(100000000000000))),
		},
	}

	bz, err := json.MarshalIndent(&msgFeesGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated msgfees parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&msgFeesGenesis)
}
