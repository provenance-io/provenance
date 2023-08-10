package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/x/oracle/types"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const (
	SequenceID = "sequence_id"
	Port       = "port"
)

// SequenceIDFn randomized sequence id
func SequenceIDFn(r *rand.Rand) uint64 {
	return uint64(randIntBetween(r, 0, 10000000000))
}

// PortFn randomized port name
func PortFn(r *rand.Rand) string {
	if r.Intn(2) > 0 {
		return "oracle"
	}
	length := uint64(randIntBetween(r, 6, 10))
	return strings.ToLower(simtypes.RandStringOfLength(r, int(length)))
}

// OracleFn randomized oracle address
func OracleFn(r *rand.Rand, accs []simtypes.Account) string {
	randomAccount, _ := RandomAccs(r, accs, 1)
	if r.Intn(2) > 0 {
		return ""
	}
	return randomAccount[0].Address.String()
}

// RandomizedGenState generates a random GenesisState for trigger
func RandomizedGenState(simState *module.SimulationState) {
	var sequenceID uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SequenceID, &sequenceID, simState.Rand,
		func(r *rand.Rand) { sequenceID = SequenceIDFn(r) },
	)

	var port string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Port, &port, simState.Rand,
		func(r *rand.Rand) { port = PortFn(r) },
	)

	params := types.DefaultParams()

	var oracle string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Port, &port, simState.Rand,
		func(r *rand.Rand) { oracle = OracleFn(r, simState.Accounts) },
	)

	genesis := types.NewGenesisState(port, params, sequenceID, oracle)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)

	bz, err := json.MarshalIndent(simState.GenState[types.ModuleName], "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated oracle parameters:\n%s\n", bz)
}

// randIntBetween generates a random number between min and max inclusive.
func randIntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}

// RandomChannel returns a random channel
func RandomChannel(r *rand.Rand) string {
	channelNumber := r.Intn(1000)
	return fmt.Sprintf("channel-%d", channelNumber)
}
