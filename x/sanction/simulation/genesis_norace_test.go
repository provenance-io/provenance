//go:build !race

package simulation_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/sanction"
	"github.com/provenance-io/provenance/x/sanction/simulation"
)

func TestRandomizedGenStateImportExport(t *testing.T) {
	// This goes through 1001 seeds and:
	// 1. generates a random genesis,
	// 2. imports it into an app,
	// 3. exports the sanction genesis state from the app,
	// 4. makes sure the exported gen state is equal to the one randomly generated.
	// It will stop at the first seed that fails.

	cdc := app.MakeTestEncodingConfig(t).Marshaler
	accounts := generateAccounts(t)

	for i := int64(0); i <= 1000; i++ {
		passed := t.Run(fmt.Sprintf("seed %d", i), func(t *testing.T) {
			simState := module.SimulationState{
				AppParams:    make(simtypes.AppParams),
				Cdc:          cdc,
				Rand:         rand.New(rand.NewSource(i)),
				NumBonded:    3,
				Accounts:     make([]simtypes.Account, len(accounts)),
				InitialStake: sdkmath.NewInt(1000),
				GenState:     make(map[string]json.RawMessage),
			}
			copy(simState.Accounts, accounts)

			testRandGen := func() {
				simulation.RandomizedGenState(&simState)
			}
			require.NotPanics(t, testRandGen, "RandomizedGenState")

			var randomGenState sanction.GenesisState
			err := simState.Cdc.UnmarshalJSON(simState.GenState[sanction.ModuleName], &randomGenState)
			require.NoError(t, err, "UnmarshalJSON to sanction.GenesisState")

			if randomGenState.Params.ImmediateSanctionMinDeposit.IsZero() {
				randomGenState.Params.ImmediateSanctionMinDeposit = nil
			}
			if randomGenState.Params.ImmediateUnsanctionMinDeposit.IsZero() {
				randomGenState.Params.ImmediateUnsanctionMinDeposit = nil
			}

			provApp := app.Setup(t)
			ctx := provApp.BaseApp.NewContext(false)

			testInit := func() {
				provApp.SanctionKeeper.InitGenesis(ctx, &randomGenState)
			}
			require.NotPanics(t, testInit, "sanction InitGenesis")

			var actualGenState *sanction.GenesisState
			testExport := func() {
				actualGenState = provApp.SanctionKeeper.ExportGenesis(ctx)
			}
			require.NotPanics(t, testExport, "ExportGenesis")

			assert.ElementsMatch(t, randomGenState.SanctionedAddresses, actualGenState.SanctionedAddresses, "SanctionedAddresses, A = expected, B = actual")
			assert.ElementsMatch(t, randomGenState.TemporaryEntries, actualGenState.TemporaryEntries, "TemporaryEntries, A = expected, B = actual")
			if !assert.Equal(t, randomGenState.Params, actualGenState.Params, "Params") && randomGenState.Params != nil && actualGenState.Params != nil {
				assert.Equal(t, randomGenState.Params.ImmediateSanctionMinDeposit.String(),
					actualGenState.Params.ImmediateSanctionMinDeposit.String(),
					"Params.ImmediateSanctionMinDeposit")
				assert.Equal(t, randomGenState.Params.ImmediateUnsanctionMinDeposit.String(),
					actualGenState.Params.ImmediateUnsanctionMinDeposit.String(),
					"Params.ImmediateUnsanctionMinDeposit")
			}
		})
		if !passed {
			break
		}
	}
}
