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

// RandomizedGenState generates a random GenesisState for trigger
func RandomizedGenState(simState *module.SimulationState) {
	var contract string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Contract, &contract, simState.Rand,
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

// randIntBetween generates a random number between min and max inclusive.
func randIntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}

func RandomAccs(r *rand.Rand, accs []simtypes.Account, count uint64) ([]simtypes.Account, error) {
	if uint64(len(accs)) < count {
		return nil, fmt.Errorf("cannot choose %d accounts because there are only %d", count, len(accs))
	}
	raccs := make([]simtypes.Account, 0, len(accs))
	raccs = append(raccs, accs...)
	r.Shuffle(len(raccs), func(i, j int) {
		raccs[i], raccs[j] = raccs[j], raccs[i]
	})
	return raccs[:count], nil
}
