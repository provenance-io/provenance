package app

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/module"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeetype "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

const (
	chainID = "sim-provenance"
)

func init() {
	sdksim.GetSimulatorFlags()
}

type StoreKeysPrefixes struct {
	A        sdk.StoreKey
	B        sdk.StoreKey
	Prefixes [][]byte
}

func TestFullAppSimulation(t *testing.T) {
	config, db, dir, logger, skip, err := sdksim.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping provenance application simulation")
	}
	require.NoError(t, err, "provenance simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), sdksim.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "provenanced", app.Name())

	fmt.Printf("running provenance full app simulation")

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		sdksim.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksim.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sdksim.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sdksim.PrintStats(db)
	}
}

func TestSmartContracts(t *testing.T) {
	config, db, dir, logger, skip, err := sdksim.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping provenance application simulation")
	}
	require.NoError(t, err, "provenance simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), sdksim.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "provenanced", app.Name())



	fmt.Printf("running provenance full app simulation for smart contracts")

	header := tmproto.Header{Height: app.LastBlockHeight()}
	ctx := app.NewContext(true, header)
	ctx = ctx.WithBlockTime(time.Now())

	nano := ctx.BlockTime().UnixNano()

	if nano < 1 {

		fmt.Println(("Block (unix) time must never be empty or negative "))
		fmt.Println(ctx.BlockTime())
		fmt.Println(header.Time)
		fmt.Println(nano)
		fmt.Println("--------")
		panic(nano)
	}

	//ctx := sdk.NewContext(nil, tmproto.Header{Height: app.LastBlockHeight()}, false, logger)

	fmt.Println("ctx: ")
	fmt.Println(ctx)

	// do we want to just use the weighted operations from the module or do our own stuff afterwards???

	simState := module.SimulationState{
		AppParams: make(simtypes.AppParams),
		Cdc:       app.AppCodec(),
	}

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		sdksim.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		//sdksim.SimulationOperations(app, app.AppCodec(), config),
		app.sm.Modules[16].WeightedOperations(simState),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	//err = sdksim.CheckExportSimulation(app, config, simParams)
	//require.NoError(t, err)
	require.NoError(t, simErr)
	fmt.Println(simParams)
	sdksim.PrintStats(db)
}

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/provenance-io/provenance -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func TestAppImportExport(t *testing.T) {
	// uncomment to run in ide without flags.
	//sdksim.FlagEnabledValue = true

	config, db, dir, logger, skip, err := sdksim.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), sdksim.EmptyAppOptions{}, fauxMerkleModeOpt)

	fmt.Printf("running provenance benchmark full app simulation")

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		sdksim.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksim.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sdksim.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sdksim.PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := sdksim.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := New(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), sdksim.EmptyAppOptions{}, fauxMerkleModeOpt)

	var genesisState sdksim.GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	newApp.mm.InitGenesis(ctxB, app.AppCodec(), genesisState)
	newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)

	fmt.Printf("comparing stores...\n")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keys[authtypes.StoreKey], newApp.keys[authtypes.StoreKey], [][]byte{}},
		{app.keys[stakingtypes.StoreKey], newApp.keys[stakingtypes.StoreKey],
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey,
			}}, // ordering may change but it doesn't matter
		{app.keys[slashingtypes.StoreKey], newApp.keys[slashingtypes.StoreKey], [][]byte{}},
		{app.keys[minttypes.StoreKey], newApp.keys[minttypes.StoreKey], [][]byte{}},
		{app.keys[distrtypes.StoreKey], newApp.keys[distrtypes.StoreKey], [][]byte{}},
		{app.keys[banktypes.StoreKey], newApp.keys[banktypes.StoreKey], [][]byte{banktypes.BalancesPrefix}},
		{app.keys[paramtypes.StoreKey], newApp.keys[paramtypes.StoreKey], [][]byte{}},
		{app.keys[govtypes.StoreKey], newApp.keys[govtypes.StoreKey], [][]byte{}},
		{app.keys[evidencetypes.StoreKey], newApp.keys[evidencetypes.StoreKey], [][]byte{}},
		{app.keys[capabilitytypes.StoreKey], newApp.keys[capabilitytypes.StoreKey], [][]byte{}},
		{app.keys[authzkeeper.StoreKey], newApp.keys[authzkeeper.StoreKey], [][]byte{}},

		{app.keys[markertypes.StoreKey], newApp.keys[markertypes.StoreKey], [][]byte{}},
		{app.keys[msgfeetype.StoreKey], newApp.keys[msgfeetype.StoreKey], [][]byte{}},
		{app.keys[attributetypes.StoreKey], newApp.keys[attributetypes.StoreKey], [][]byte{}},
		{app.keys[nametypes.StoreKey], newApp.keys[nametypes.StoreKey], [][]byte{}},
		{app.keys[metadatatypes.StoreKey], newApp.keys[metadatatypes.StoreKey], [][]byte{}},
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, 0, len(failedKVAs), sdksim.GetSimulationLog(skp.A.Name(), app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, dir, logger, skip, err := sdksim.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation after import")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), sdksim.EmptyAppOptions{}, fauxMerkleModeOpt)

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		sdksim.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksim.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sdksim.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sdksim.PrintStats(db)
	}

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(true, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := sdksim.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := New(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), sdksim.EmptyAppOptions{}, fauxMerkleModeOpt)

	newApp.InitChain(abci.RequestInitChain{
		AppStateBytes: exported.AppState,
	})

	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		newApp.BaseApp,
		sdksim.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksim.SimulationOperations(newApp, newApp.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)
	require.NoError(t, err)
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	if !sdksim.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := sdksim.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = helpers.SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if sdksim.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			app := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), sdksim.EmptyAppOptions{}, interBlockCacheOpt())

			fmt.Printf(
				"running provenance non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				sdksim.AppStateFn(app.AppCodec(), app.SimulationManager()),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				sdksim.SimulationOperations(app, app.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				sdksim.PrintStats(db)
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

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}
