package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// Simulation parameter constants
const (
	Contract = "contract"
)

// ContractFn randomized conntract address
func ContractFn(r *rand.Rand, accs []simtypes.Account) string {
	randomAccount, _ := RandomAccs(r, accs, 1)
	if r.Intn(2) > 0 || len(randomAccount) == 0 {
		return ""
	}
	return randomAccount[0].Address.String()
}

// RandomizedGenState generates a random GenesisState for ibcratelimit
func RandomizedGenState(simState *module.SimulationState) {
	var contract string
	simState.AppParams.GetOrGenerate(
		Contract, &contract, simState.Rand,
		func(r *rand.Rand) { contract = ContractFn(r, simState.Accounts) },
	)

	genesis := ibcratelimit.NewGenesisState(ibcratelimit.NewParams(contract))
	simState.GenState[ibcratelimit.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)

	bz, err := json.MarshalIndent(simState.GenState[ibcratelimit.ModuleName], "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated ratelimitedibc parameters:\n%s\n", bz)
}
