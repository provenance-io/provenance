package app

// DONTCOVER

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"
	icagenesistypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/genesis/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	cmdconfig "github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
)

func init() {
	simcli.GetSimulatorFlags()
	pioconfig.SetProvenanceConfig("", 0)
}

// provAppStateFn wraps the simtypes.AppStateFn and sets the ICA and ICQ GenesisState if isn't yet defined in the appState.
func provAppStateFn(cdc codec.JSONCodec, simManager *module.SimulationManager, genesisState map[string]json.RawMessage) simtypes.AppStateFn {
	return func(r *rand.Rand, accs []simtypes.Account, config simtypes.Config) (json.RawMessage, []simtypes.Account, string, time.Time) {
		appState, simAccs, chainID, genesisTimestamp := simtestutil.AppStateFn(cdc, simManager, genesisState)(r, accs, config)
		appState = appStateWithICA(appState, cdc)
		appState = appStateWithICQ(appState, cdc)
		appState = appStateWithWasmSeqs(appState, cdc)
		return appState, simAccs, chainID, genesisTimestamp
	}
}

// appStateWithICA checks the given appState for an ica entry. If it's not found, it's populated with the defaults.
func appStateWithICA(appState json.RawMessage, cdc codec.JSONCodec) json.RawMessage {
	return mutateAppState(appState, cdc, icatypes.ModuleName, icagenesistypes.DefaultGenesis(), nil)
}

// appStateWithICA checks the given appState for an ica entry. If it's not found, it's populated with the defaults.
func appStateWithICQ(appState json.RawMessage, cdc codec.JSONCodec) json.RawMessage {
	return mutateAppState(appState, cdc, icqtypes.ModuleName, icqtypes.DefaultGenesis(), nil)
}

// appStateWithWasmSeqs ensures the wasm genesis state has sequence entries.
func appStateWithWasmSeqs(appState json.RawMessage, cdc codec.JSONCodec) json.RawMessage {
	// During the import/export test, the wasm module state of the first app ends up with empty sequence entries.
	// But export-genesis reads empty sequence entries as "1" and creates an entry in the Sequences list.
	// So when we call InitGenesis with the second app, those sequences are stored in state, resulting in different states between the two apps.
	// The behavior of the wasm module is the same if the value is 1 or empty, so the behavior of the two apps is the same, just not the state.
	// So in here, we make sure there are default entries for the sequences so that even the first one has them.
	// If the sims do any wasm stuff that'll change those entries, both the state and exported genesis will match, and there won't be a problem.
	return mutateAppState(appState, cdc, wasmtypes.ModuleName, &wasmtypes.GenesisState{Params: wasmtypes.DefaultParams()}, func(wasmGenState *wasmtypes.GenesisState) *wasmtypes.GenesisState {
		requiredSequences := [][]byte{wasmtypes.KeySequenceCodeID, wasmtypes.KeySequenceInstanceID}
		for _, seqKey := range requiredSequences {
			if !sequenceExists(wasmGenState.Sequences, seqKey) {
				wasmGenState.Sequences = append(wasmGenState.Sequences, wasmtypes.Sequence{IDKey: seqKey, Value: 1})
			}
		}
		return wasmGenState
	})
}

// sequenceExists checks if a sequence entry exists for the given key.
func sequenceExists(sequences []wasmtypes.Sequence, key []byte) bool {
	for _, seq := range sequences {
		if bytes.Equal(seq.IDKey, key) {
			return true
		}
	}
	return false
}

// mutateAppState returns a new appState with an updated (or new) entry for the given module.
// The mutator will receive the module's genesis state from appState if it exists, otherwise,
// it'll receive the provided defaultState. If the mutator is nil, this basically just ensures
// that the appState has an entry for the module, setting it to the default if it's not already there.
func mutateAppState[G proto.Message](appState json.RawMessage, cdc codec.JSONCodec, moduleName string, defaultState G, mutator func(state G) G) json.RawMessage {
	rawState := make(map[string]json.RawMessage)
	err := json.Unmarshal(appState, &rawState)
	if err != nil {
		panic(fmt.Sprintf("error unmarshalling appstate: %v", err))
	}

	if len(rawState[moduleName]) > 0 {
		err = cdc.UnmarshalJSON(rawState[moduleName], defaultState)
		if err != nil {
			panic(fmt.Sprintf("error unmarshalling %s genesis state from %q as %T: %v",
				moduleName, string(rawState[moduleName]), defaultState, err))
		}
	}

	if mutator != nil {
		defaultState = mutator(defaultState)
	}

	rawState[moduleName], err = cdc.MarshalJSON(defaultState)
	if err != nil {
		panic(fmt.Sprintf("error marshalling %s genesis state %#v: %v",
			moduleName, defaultState, err))
	}

	appState, err = json.Marshal(rawState)
	if err != nil {
		panic(fmt.Sprintf("error marshalling appstate: %v", err))
	}

	return appState
}

func setupSimulation(dirPrefix string, dbName string) (simtypes.Config, dbm.DB, string, log.Logger, bool, error) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = pioconfig.SimAppChainID
	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, dirPrefix, dbName, simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	return config, db, dir, logger, skip, err
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

// printStats outputs the config and db info.
func printStats(config simtypes.Config, db dbm.DB) {
	printConfig(config)
	if config.Commit {
		printDBInfo(db)
	}
}

// printConfig outputs the config.
func printConfig(config simtypes.Config) {
	fmt.Println("-vvv-  Config Info  -vvv-")
	cfields := cmdconfig.MakeFieldValueMap(config, true)
	for _, f := range cfields.GetSortedKeys() {
		fmt.Printf("%s: %s\n", f, cfields.GetStringOf(f))
	}
	fmt.Println("-^^^-  Config Info  -^^^-")
}

// printDBInfo outputs the db.Stats map.
func printDBInfo(db dbm.DB) {
	fmt.Println("-vvv-  Database Info  -vvv-")
	dbStats := db.Stats()
	if len(dbStats) == 0 {
		fmt.Println("No info to report.")
	} else {
		keys := make([]string, 0, len(dbStats))
		for k := range dbStats {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := dbStats[k]
			if strings.Contains(v, "\n") {
				fmt.Printf("%s:\n", k)
				fmt.Println(v)
			} else {
				fmt.Printf("%s: %q\n", k, v)
			}
		}
	}
	fmt.Println("-^^^-  Database Info  -^^^-")
}

// newSimAppOpts creates a new set of AppOptions with a temp dir for home, and the desired invariant check period.
func newSimAppOpts(t testing.TB) simtestutil.AppOptionsMap {
	return simtestutil.AppOptionsMap{
		flags.FlagHome:            t.TempDir(),
		server.FlagInvCheckPeriod: simcli.FlagPeriodValue,
	}
}

func TestFullAppSimulation(t *testing.T) {
	config, db, dir, logger, skip, err := setupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping provenance application simulation")
	}
	printConfig(config)
	require.NoError(t, err, "provenance simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOpts := newSimAppOpts(t)
	baseAppOpts := []func(*baseapp.BaseApp){
		fauxMerkleModeOpt,
		baseapp.SetChainID(config.ChainID),
	}
	app := New(logger, db, nil, true, appOpts, baseAppOpts...)
	require.Equal(t, "provenanced", app.Name())
	if !simcli.FlagSigverifyTxValue {
		app.SetNotSigverifyTx()
	}

	fmt.Printf("running provenance full app simulation\n")

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err, "CheckExportSimulation")
	require.NoError(t, simErr, "SimulateFromSeed")

	printStats(config, db)
}

func TestSimple(t *testing.T) {
	config, db, dir, logger, skip, err := setupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping provenance application simulation")
	}
	printConfig(config)
	require.NoError(t, err, "provenance simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOpts := newSimAppOpts(t)
	baseAppOpts := []func(*baseapp.BaseApp){
		fauxMerkleModeOpt,
		baseapp.SetChainID(config.ChainID),
	}
	app := New(logger, db, nil, true, appOpts, baseAppOpts...)
	require.Equal(t, "provenanced", app.Name())
	if !simcli.FlagSigverifyTxValue {
		app.SetNotSigverifyTx()
	}

	// run randomized simulation
	_, _, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	require.NoError(t, simErr, "SimulateFromSeed")
	printStats(config, db)
}

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/provenance-io/provenance -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func TestAppImportExport(t *testing.T) {
	// uncomment to run in ide without flags.
	//simcli.FlagEnabledValue = true
	//tempDir, err := os.MkdirTemp("", "sim-log-*")
	//require.NoError(t, err, "MkdirTemp")
	//t.Logf("tempDir: %s", tempDir)
	//simcli.FlagNumBlocksValue = 30
	//simcli.FlagVerboseValue = true
	//simcli.FlagCommitValue = true
	//simcli.FlagSeedValue = 2
	//simcli.FlagPeriodValue = 3
	//simcli.FlagExportParamsPathValue = filepath.Join(tempDir, fmt.Sprintf("sim_params-%d.json", simcli.FlagSeedValue))
	//simcli.FlagExportStatePathValue = filepath.Join(tempDir, fmt.Sprintf("sim_state-%d.json", simcli.FlagSeedValue))

	config, db, dir, logger, skip, err := setupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	printConfig(config)
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOpts := newSimAppOpts(t)
	baseAppOpts := []func(*baseapp.BaseApp){
		fauxMerkleModeOpt,
		baseapp.SetChainID(config.ChainID),
	}
	app := New(logger, db, nil, true, appOpts, baseAppOpts...)
	require.Equal(t, "provenanced", app.Name())
	if !simcli.FlagSigverifyTxValue {
		app.SetNotSigverifyTx()
	}

	fmt.Printf("running provenance test import export\n")

	// Run randomized simulation
	_, lastBlockTime, simParams, simErr := simulation.SimulateFromSeedProv(
		t,
		os.Stdout,
		app.BaseApp,
		provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err, "CheckExportSimulation")
	require.NoError(t, simErr, "SimulateFromSeedProv")

	printStats(config, db)

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err, "ExportAppStateAndValidators")

	fmt.Printf("importing genesis...\n")

	newDB, newDir, newLogger, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim-2", "Simulation-2", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(t, err, "simulation setup 2 failed")

	defer func() {
		require.NoError(t, newDB.Close())
		require.NoError(t, os.RemoveAll(newDir))
	}()

	// create a new temp dir for the app to fix wasmvm data dir lockfile contention
	appOpts = newSimAppOpts(t)
	newApp := New(newLogger, newDB, nil, true, appOpts, baseAppOpts...)

	var genesisState map[string]json.RawMessage
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight(), Time: lastBlockTime})
	ctxB := newApp.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight(), Time: lastBlockTime})
	_, err = newApp.mm.InitGenesis(ctxB, app.AppCodec(), genesisState)
	if err != nil {
		if strings.Contains(err.Error(), "validator set is empty after InitGenesis") {
			logger.Info("Skipping simulation as all validators have been unbonded")
			logger.Info("err", err, "stacktrace", string(debug.Stack()))
			return
		}
	}
	require.NoError(t, err, "InitGenesis")

	err = newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)
	require.NoError(t, err, "StoreConsensusParams")

	fmt.Printf("comparing stores...\n")

	// skip certain prefixes
	skipPrefixes := map[string][][]byte{
		stakingtypes.StoreKey: {
			stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
			stakingtypes.HistoricalInfoKey, stakingtypes.UnbondingIDKey, stakingtypes.UnbondingIndexKey,
			stakingtypes.UnbondingTypeKey, stakingtypes.ValidatorUpdatesKey,
		},
		authzkeeper.StoreKey:   {authzkeeper.GrantQueuePrefix},
		feegrant.StoreKey:      {feegrant.FeeAllowanceQueueKeyPrefix},
		slashingtypes.StoreKey: {slashingtypes.ValidatorMissedBlockBitmapKeyPrefix},
		wasmtypes.StoreKey:     {wasmtypes.TXCounterPrefix},
	}

	storeKeys := app.GetStoreKeys()
	require.NotEmpty(t, storeKeys, "storeKeys")

	for _, appKeyA := range storeKeys {
		keyName := appKeyA.Name()
		t.Run(keyName, func(t *testing.T) {
			// only compare kvstores
			if _, ok := appKeyA.(*storetypes.KVStoreKey); !ok {
				t.Skipf("Skipping because the key is a %T (not a KVStoreKey)", appKeyA)
				return
			}

			appKeyB := newApp.GetKey(keyName)

			storeA := ctxA.KVStore(appKeyA)
			storeB := ctxB.KVStore(appKeyB)

			failedKVAs, failedKVBs := simtestutil.DiffKVStores(storeA, storeB, skipPrefixes[keyName])
			assert.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare: %s", keyName)
			fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), appKeyA, appKeyB)

			// Make the lists the same length because GetSimulationLog assumes they're that way.
			for len(failedKVBs) < len(failedKVAs) {
				failedKVBs = append(failedKVBs, kv.Pair{Key: []byte{}, Value: []byte{}})
			}
			for len(failedKVBs) > len(failedKVAs) {
				failedKVAs = append(failedKVAs, kv.Pair{Key: []byte{}, Value: []byte{}})
			}

			assert.Equal(t, 0, len(failedKVAs), simtestutil.GetSimulationLog(keyName, app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
		})
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, dir, logger, skip, err := setupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation after import")
	}
	printConfig(config)
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOpts := newSimAppOpts(t)
	baseAppOpts := []func(*baseapp.BaseApp){
		fauxMerkleModeOpt,
		baseapp.SetChainID(config.ChainID),
	}
	app := New(logger, db, nil, true, appOpts, baseAppOpts...)
	if !simcli.FlagSigverifyTxValue {
		app.SetNotSigverifyTx()
	}

	// Run randomized simulation
	stopEarly, lastBlockTime, simParams, simErr := simulation.SimulateFromSeedProv(
		t,
		os.Stdout,
		app.BaseApp,
		provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err, "CheckExportSimulation")
	require.NoError(t, simErr, "SimulateFromSeedProv")

	printStats(config, db)

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(true, nil, nil)
	require.NoError(t, err, "ExportAppStateAndValidators")

	fmt.Printf("importing genesis...\n")

	newDB, newDir, newLogger, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim-2", "Simulation-2", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(t, err, "simulation setup 2 failed")

	defer func() {
		require.NoError(t, newDB.Close())
		require.NoError(t, os.RemoveAll(newDir))
	}()

	// create a new temp dir for the app to fix wasmvm data dir lockfile contention
	appOpts = newSimAppOpts(t)
	newApp := New(newLogger, newDB, nil, true, appOpts, baseAppOpts...)

	_, err = newApp.InitChain(&abci.RequestInitChain{
		AppStateBytes: exported.AppState,
		ChainId:       config.ChainID,
		Time:          lastBlockTime,
	})
	require.NoError(t, err, "InitChain")

	simcli.FlagGenesisTimeValue = lastBlockTime.Unix()
	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		newApp.BaseApp,
		provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(newApp, newApp.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)
	require.NoError(t, err)
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	// uncomment these to run in ide without flags.
	//simcli.FlagEnabledValue = true
	//simcli.FlagBlockSizeValue = 100
	//simcli.FlagNumBlocksValue = 50

	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = pioconfig.SimAppChainID
	config.DBBackend = "memdb"
	config.Commit = true

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	var seeds []int64
	if config.Seed != simcli.DefaultSeedValue {
		// If a seed was provided, just do that one.
		numSeeds = 1
		seeds = append(seeds, config.Seed)
	} else {
		// Otherwise, pick random seeds to use.
		seeds = make([]int64, numSeeds)
		for i := range seeds {
			seeds[i] = rand.Int63()
		}
	}

	for i, seed := range seeds {
		config.Seed = seed
		printConfig(config)

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simcli.FlagVerboseValue {
				logger = log.NewTestLogger(t)
			} else {
				logger = log.NewNopLogger()
			}

			// create a new temp dir for the app to fix wasmvm data dir lockfile contention
			appOpts := newSimAppOpts(t)
			if simcli.FlagVerboseValue {
				appOpts[flags.FlagLogLevel] = "debug"
			}

			db := dbm.NewMemDB()
			app := New(logger, db, nil, true, appOpts, interBlockCacheOpt(), baseapp.SetChainID(config.ChainID))
			if !simcli.FlagSigverifyTxValue {
				app.SetNotSigverifyTx()
			}

			fmt.Printf(
				"running provenance non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simtestutil.SimulationOperations(app, app.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			printStats(config, db)

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
