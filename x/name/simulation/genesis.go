package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/name/types"
)

// Simulation parameter constants
const (
	MaxSegmentLength       = "max_segment_length"
	MinSegmentLength       = "min_segment_length"
	MaxNameLevels          = "max_namne_levels"
	AllowUnrestrictedNames = "allow_unrestricted_names"
	RootNameSegment        = "root_name_segment"
)

// GenMaxSegmentLength randomized Max Segment Length
func GenMaxSegmentLength(r *rand.Rand) uint32 {
	return uint32(r.Intn(22) + 11) // ensures that max is always more than range of min values (1-11)
}

// GenMaxNameLevels randomized Maximum number of segment levels
func GenMaxNameLevels(r *rand.Rand) uint32 {
	return uint32(r.Intn(9) + 2)
}

// GenMinSegmentLength randomized minimum segment name length
func GenMinSegmentLength(r *rand.Rand) uint32 {
	return uint32(r.Intn(10) + 1)
}

// GenAllowUnrestrictedNames returns a randomized AllowUnrestrictedNames parameter.
func GenAllowUnrestrictedNames(r *rand.Rand) bool {
	return r.Int63n(101) <= 50 // 50% chance of unrestricted names being enabled
}

// GenRootNameSegment returns a randomized String to use for the root name binding
func GenRootNameSegment(r *rand.Rand, minSegmentLength uint32) string {
	return strings.ToLower(simtypes.RandStringOfLength(r, int(minSegmentLength)))
}

// RandomizedGenState generates a random GenesisState for name
func RandomizedGenState(simState *module.SimulationState) {
	var maxValueLength uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxSegmentLength, &maxValueLength, simState.Rand,
		func(r *rand.Rand) { maxValueLength = GenMaxSegmentLength(r) },
	)

	var maxNameLevels uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxSegmentLength, &maxNameLevels, simState.Rand,
		func(r *rand.Rand) { maxNameLevels = GenMaxNameLevels(r) },
	)

	var minValueLength uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxSegmentLength, &minValueLength, simState.Rand,
		func(r *rand.Rand) { minValueLength = GenMinSegmentLength(r) },
	)

	var allowUnrestrictedNames bool
	simState.AppParams.GetOrGenerate(
		simState.Cdc, AllowUnrestrictedNames, &allowUnrestrictedNames, simState.Rand,
		func(r *rand.Rand) { allowUnrestrictedNames = GenAllowUnrestrictedNames(r) },
	)

	var rootNameSegment string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, RootNameSegment, &rootNameSegment, simState.Rand,
		func(r *rand.Rand) { rootNameSegment = GenRootNameSegment(r, minValueLength) },
	)

	accountGenesis := types.GenesisState{
		Params: types.Params{
			MaxSegmentLength:       maxValueLength,
			MaxNameLevels:          maxNameLevels,
			MinSegmentLength:       minValueLength,
			AllowUnrestrictedNames: allowUnrestrictedNames,
		},
		Bindings: []types.NameRecord{
			types.NewNameRecord(rootNameSegment, simState.Accounts[0].Address, false),
		},
	}

	bz, err := json.MarshalIndent(&accountGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated name parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&accountGenesis)
}
