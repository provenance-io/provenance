package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/flatfees/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	params := types.DefaultParams()
	genState := types.GenesisState{
		Params: params,
		MsgFees: []*types.MsgFee{
			// Adding fees for create marker with asking for a large number of stake to make sure that
			// the call is failed without the additional fee provided.
			{
				MsgTypeUrl: sdk.MsgTypeURL(&markertypes.MsgAddMarkerRequest{}),
				Cost:       sdk.Coins{sdk.NewInt64Coin(params.DefaultCost.Denom, 100_000_000_000_000)},
			},
		},
	}

	bz, err := json.MarshalIndent(&genState, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected standard flatfees parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&genState)
}
