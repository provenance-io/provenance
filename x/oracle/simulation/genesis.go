package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	internalrand "github.com/provenance-io/provenance/internal/rand"
	"github.com/provenance-io/provenance/x/oracle/types"
)

const (
	Port = "port"
)

// PortFn randomized port name
func PortFn(r *rand.Rand) string {
	if r.Intn(2) > 0 {
		return "oracle"
	}
	length := uint64(internalrand.IntBetween(r, 6, 10))
	return strings.ToLower(simtypes.RandStringOfLength(r, int(length)))
}

// OracleFn randomized oracle address
func OracleFn(r *rand.Rand, accs []simtypes.Account) string {
	randomAccount, _ := internalrand.SelectAccounts(r, accs, 1)
	if r.Intn(2) > 0 || len(randomAccount) == 0 {
		return ""
	}
	return randomAccount[0].Address.String()
}

// RandomizedGenState generates a random GenesisState for trigger
func RandomizedGenState(simState *module.SimulationState) {
	var port string
	simState.AppParams.GetOrGenerate(
		Port, &port, simState.Rand,
		func(r *rand.Rand) { port = PortFn(r) },
	)

	var oracle string
	simState.AppParams.GetOrGenerate(
		Port, &port, simState.Rand,
		func(r *rand.Rand) { oracle = OracleFn(r, simState.Accounts) },
	)

	genesis := types.NewGenesisState(port, oracle)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)

	bz, err := json.MarshalIndent(simState.GenState[types.ModuleName], "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated oracle parameters:\n%s\n", bz)
}

// RandomChannel returns a random channel
func RandomChannel(r *rand.Rand) string {
	channelNumber := r.Intn(1000)
	return fmt.Sprintf("channel-%d", channelNumber)
}
