package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/provenance-io/provenance/x/msgfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
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
			// changed it to some obscure vesting type since sim tests as they are written will fail if a message actually has a fee on it :heavysigh:
			//  however it does help create an app with a genesis state so not totally useless.
			types.NewMsgFee(sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}), sdk.NewCoin("stake", sdk.NewInt(1))),
		},
	}

	bz, err := json.MarshalIndent(&msgFeesGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated msgfees parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&msgFeesGenesis)
}
