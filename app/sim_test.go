//go:build sims

package app

// DONTCOVER

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	vaulttypes "github.com/provlabs/vault/types"
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
	sims "github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	icagenesistypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/genesis/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	cmdconfig "github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
)

func init() {
	simcli.GetSimulatorFlags()
}

// provAppStateFn wraps the simtypes.AppStateFn and sets the ICA and ICQ GenesisState if isn't yet defined in the appState.
func provAppStateFn(cdc codec.JSONCodec, simManager *module.SimulationManager, genesisState map[string]json.RawMessage) simtypes.AppStateFn {
	return func(r *rand.Rand, accs []simtypes.Account, config simtypes.Config) (json.RawMessage, []simtypes.Account, string, time.Time) {
		pioconfig.SetProvConfig(sdk.DefaultBondDenom)
		appState, simAccs, chainID, genesisTimestamp := simtestutil.AppStateFn(cdc, simManager, genesisState)(r, accs, config)
		appState = appStateWithICA(appState, cdc)
		appState = appStateWithWasmSeqs(appState, cdc)
		return appState, simAccs, chainID, genesisTimestamp
	}
}

// appStateWithICA checks the given appState for an ica entry. If it's not found, it's populated with the defaults.
func appStateWithICA(appState json.RawMessage, cdc codec.JSONCodec) json.RawMessage {
	return mutateAppState(appState, cdc, icatypes.ModuleName, icagenesistypes.DefaultGenesis(), nil)
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
	sims.Run(t, New, setupStateFactory)
}

func TestSimple(t *testing.T) {
	sims.Run(t, New, setupStateFactory)
}

// setupStateFactory is the Provenance-custom SimStateFactory builder.
// It uses provAppStateFn (wraps with ICA genesis + wasm sequence defaults).
func setupStateFactory(app *App) sims.SimStateFactory {
	return sims.SimStateFactory{
		Codec:         app.AppCodec(),
		AppStateFn:    provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		BlockedAddr:   app.ModuleAccountAddrs(),
		AccountSource: app.AccountKeeper,
		BalanceSource: app.BankKeeper,
	}
}

var (
	exportAllModules       = []string{}
	exportWithValidatorSet = []string{}
)

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

	sims.Run(t, New, setupStateFactory, func(tb testing.TB, ti sims.TestInstance[*App], accs []simtypes.Account) {
		tb.Helper()
		app := ti.App

		tb.Log("exporting genesis...\n")
		exported, err := app.ExportAppStateAndValidators(false, exportWithValidatorSet, exportAllModules)
		require.NoError(tb, err)

		tb.Log("importing genesis...\n")
		newTestInstance := sims.NewSimulationAppInstance(tb, ti.Cfg, New)
		newApp := newTestInstance.App

		var genesisState map[string]json.RawMessage
		require.NoError(tb, json.Unmarshal(exported.AppState, &genesisState))

		ctxB := newApp.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight(), Time: ti.EndBlockTime})
		_, err = newApp.mm.InitGenesis(ctxB, newApp.appCodec, genesisState)
		if IsEmptyValidatorSetErr(err) {
			tb.Skip("Skipping simulation as all validators have been unbonded")
			return
		}
		require.NoError(tb, err)

		err = newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)
		require.NoError(tb, err)

		tb.Log("comparing stores...")
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
			wasmtypes.StoreKey:     {wasmtypes.TXCounterPrefix, wasmtypes.ContractCodeHistoryElementPrefix},
			vaulttypes.StoreKey:    {vaulttypes.VaultPayoutVerificationSetPrefix},
			attrtypes.StoreKey:     {attrtypes.AttributeAddrLookupKeyPrefix, attrtypes.AttributeExpirationKeyPrefix}, //0x03 -derived the secondary index rebuilt on import
		}
		AssertEqualStores(tb, app, newApp, app.SimulationManager().StoreDecoders, skipPrefixes)
	})
}

func TestAppSimulationAfterImport(t *testing.T) {
	sims.Run(t, New, setupStateFactory, func(tb testing.TB, ti sims.TestInstance[*App], accs []simtypes.Account) {
		tb.Helper()
		app := ti.App

		tb.Log("exporting genesis...\n")
		exported, err := app.ExportAppStateAndValidators(false, exportWithValidatorSet, exportAllModules)
		require.NoError(tb, err)

		tb.Log("importing genesis...\n")
		newCfg := ti.Cfg
		newCfg.GenesisTime = ti.EndBlockTime.Unix()
		newTestInstance := sims.NewSimulationAppInstance(tb, newCfg, New)
		newApp := newTestInstance.App

		_, err = newApp.InitChain(&abci.RequestInitChain{
			AppStateBytes: exported.AppState,
			ChainId:       sims.SimAppChainID,
			Time:          ti.EndBlockTime,
		})
		if IsEmptyValidatorSetErr(err) {
			tb.Skip("Skipping simulation as all validators have been unbonded")
			return
		}
		require.NoError(tb, err)

		newStateFactory := setupStateFactory(newApp)
		_, _, err = simulation.SimulateFromSeedX(
			tb,
			newTestInstance.AppLogger,
			sims.WriteToDebugLog(newTestInstance.AppLogger),
			newApp.BaseApp,
			newStateFactory.AppStateFn,
			simtypes.RandomAccounts,
			simtestutil.BuildSimulationOperations(newApp, newApp.AppCodec(), newTestInstance.Cfg, newApp.TxConfig()),
			newStateFactory.BlockedAddr,
			newTestInstance.Cfg,
			newStateFactory.Codec,
			ti.ExecLogWriter,
		)
		require.NoError(tb, err)
	})
}

func IsEmptyValidatorSetErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "validator set is empty after InitGenesis")
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	// uncomment these to run in ide without flags.
	//simcli.FlagEnabledValue = true
	//simcli.FlagBlockSizeValue = 100
	//simcli.FlagNumBlocksValue = 50

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = pioconfig.SimAppChainID
	config.Commit = true

	numSeeds := 3
	numTimesToRunPerSeed := 3
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

	pioconfig.SetProvConfig(sdk.DefaultBondDenom)

	for i, seed := range seeds {
		config.Seed = seed

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simcli.FlagVerboseValue {
				logger = log.NewTestLogger(t)
			} else {
				logger = log.NewNopLogger()
			}

			// Reset any global state that prior tests may have set.
			simcli.FlagGenesisTimeValue = 0

			// Create a new temp dir for the app to fix wasmvm data dir lockfile contention.
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
				simtypes.RandomAccounts,
				simtestutil.SimulationOperations(app, app.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if simcli.FlagVerboseValue {
				printStats(config, db)
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

type ComparableStoreApp interface {
	LastBlockHeight() int64
	NewContextLegacy(isCheckTx bool, header cmtproto.Header) sdk.Context
	GetKey(storeKey string) *storetypes.KVStoreKey
	GetStoreKeys() []storetypes.StoreKey
}

func AssertEqualStores(
	tb testing.TB,
	app, newApp ComparableStoreApp,
	storeDecoders simtypes.StoreDecoderRegistry,
	skipPrefixes map[string][][]byte,
) {
	tb.Helper()
	ctxA := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})

	storeKeys := app.GetStoreKeys()
	require.NotEmpty(tb, storeKeys)

	for _, appKeyA := range storeKeys {
		// only compare kvstores
		if _, ok := appKeyA.(*storetypes.KVStoreKey); !ok {
			continue
		}

		keyName := appKeyA.Name()
		appKeyB := newApp.GetKey(keyName)

		storeA := ctxA.KVStore(appKeyA)
		storeB := ctxB.KVStore(appKeyB)

		failedKVAs, failedKVBs := simtestutil.DiffKVStores(storeA, storeB, skipPrefixes[keyName])
		require.Equal(tb, len(failedKVAs), len(failedKVBs),
			"unequal sets of key-values to compare %s, key stores %s and %s", keyName, appKeyA, appKeyB)

		tb.Logf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), appKeyA, appKeyB)
		if !assert.Equal(tb, 0, len(failedKVAs), simtestutil.GetSimulationLog(keyName, storeDecoders, failedKVAs, failedKVBs)) {
			for _, v := range failedKVAs {
				tb.Logf("store mismatch: %q\n", v)
			}
			tb.FailNow()
		}
	}
}
