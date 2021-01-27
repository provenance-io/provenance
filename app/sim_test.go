package app_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	provenance "github.com/provenance-io/provenance/app"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	simulation2 "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	appName = "provenanced"
	chainID = "sim-provenance"
)

func init() {
	simapp.GetSimulatorFlags()
}

func TestFullAppSimulation(t *testing.T) {
	simapp.FlagEnabledValue = true
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping provenance application simulation")
	}
	require.NoError(t, err, "provenance simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := provenance.New(appName, logger, db, nil, true, map[int64]bool{}, provenance.DefaultNodeHome(appName), simapp.FlagPeriodValue, provenance.MakeEncodingConfig(), simapp.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())

	fmt.Printf("running provenance full app simulation")

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simapp.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simulation2.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simapp.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}
}

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/provenance-io/provenance -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	// uncomment to run in ide without flags.
	//simapp.FlagEnabledValue = true
	if !simapp.FlagEnabledValue {
		b.Skip("skipping provenance application benchmark simulation")
	}

	config, db, dir, logger, _, err := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			b.Fatal(err)
		}
	}()

	app := provenance.New(appName, logger, db, nil, true, map[int64]bool{}, provenance.DefaultNodeHome(appName), simapp.FlagPeriodValue, provenance.MakeEncodingConfig(), simapp.EmptyAppOptions{}, interBlockCacheOpt())

	fmt.Printf("running provenance benchmark full app simulation")

	// Run randomized simulation:w
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		simapp.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simulation2.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simapp.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = simapp.CheckExportSimulation(app, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	if config.Commit {
		simapp.PrintStats(db)
	}
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

//// TODO: Make another test for the fuzzer itself, which just has noOp txs
//// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	if !simapp.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simapp.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = chainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simapp.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			app := provenance.New(appName, logger, db, nil, true, map[int64]bool{}, provenance.DefaultNodeHome(appName), simapp.FlagPeriodValue, provenance.MakeEncodingConfig(), simapp.EmptyAppOptions{}, interBlockCacheOpt())

			fmt.Printf(
				"running provenance non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				simapp.AppStateFn(app.AppCodec(), app.SimulationManager()),
				simulation2.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simapp.SimulationOperations(app, app.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simapp.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
