package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// Simulation parameter constants
const (
	MaxTotalSupply         = "max_total_supply"
	MaxSupply              = "max_supply"
	EnableGovernance       = "enable_governance"
	UnrestrictedDenomRegex = "unresticted_denom_regex"
)

// GenMaxTotalSupply is deprecated.
// Deprecated: GenMaxTotalSupply is kept for backwards compatability.
func GenMaxTotalSupply(r *rand.Rand) uint64 {
	return r.Uint64()
}

// GenMaxSupply randomized Maximum amount of supply to allow for markers
func GenMaxSupply(r *rand.Rand) math.Int {
	return math.NewIntFromUint64(r.Uint64())
}

// GenEnableGovernance returns a randomized EnableGovernance parameter.
func GenEnableGovernance(r *rand.Rand) bool {
	return r.Int63n(100) < 50 // 50% chance of unrestricted names being enabled
}

// GenUnrestrictedDenomRegex returns a randomized length focused string for the unrestricted denom validation expression
func GenUnrestrictedDenomRegex(r *rand.Rand) string {
	min := r.Int31n(16) + 3
	max := r.Int31n(64-min) + min
	return fmt.Sprintf(`[a-zA-Z][a-zA-Z0-9\\-\\.]{%d,%d}`, min, max)
}

// RandomizedGenState generates a random GenesisState for marker
func RandomizedGenState(simState *module.SimulationState) {
	var maxTotalSupply uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxTotalSupply, &maxTotalSupply, simState.Rand,
		func(r *rand.Rand) { maxTotalSupply = GenMaxTotalSupply(r) },
	)

	var maxSupply math.Int
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxSupply, &maxSupply, simState.Rand,
		func(r *rand.Rand) { maxSupply = GenMaxSupply(r) },
	)

	var enableGovernance bool
	simState.AppParams.GetOrGenerate(
		simState.Cdc, EnableGovernance, &enableGovernance, simState.Rand,
		func(r *rand.Rand) { enableGovernance = GenEnableGovernance(r) },
	)

	var unrestrictedDenomRegex string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, UnrestrictedDenomRegex, &unrestrictedDenomRegex, simState.Rand,
		func(r *rand.Rand) { unrestrictedDenomRegex = GenUnrestrictedDenomRegex(r) },
	)

	markerGenesis := types.GenesisState{
		Params: types.Params{
			MaxTotalSupply:         maxTotalSupply,
			MaxSupply:              maxSupply,
			EnableGovernance:       enableGovernance,
			UnrestrictedDenomRegex: unrestrictedDenomRegex,
		},
		Markers: []types.MarkerAccount{
			{
				BaseAccount: &authtypes.BaseAccount{
					Address: types.MustGetMarkerAddress(sdk.DefaultBondDenom).String(),
				},
				AccessControl: []types.AccessGrant{
					{
						Address:     simState.Accounts[0].Address.String(),
						Permissions: types.AccessListByNames("mint,burn,delete,admin"),
					},
				},
				Status:                 types.StatusActive,
				Denom:                  sdk.DefaultBondDenom,
				MarkerType:             types.MarkerType_Coin,
				SupplyFixed:            false,
				AllowGovernanceControl: true,
				AllowForcedTransfer:    false,
				RequiredAttributes:     []string{},
			},
		},
	}

	bz, err := json.MarshalIndent(&markerGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated marker parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&markerGenesis)
}
