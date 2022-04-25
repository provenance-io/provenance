package app

// DONTCOVER

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/plugin"
	kafkaplugin "github.com/cosmos/cosmos-sdk/plugin/plugins/kafka"
	kafkaservice "github.com/cosmos/cosmos-sdk/plugin/plugins/kafka/service"
	"github.com/cosmos/cosmos-sdk/server/types"
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

	cmdconfig "github.com/provenance-io/provenance/cmd/provenanced/config"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeetype "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

const (
	chainID = "sim-provenance"
)

var (
	StateListeningPlugin   string
	HaltAppOnDeliveryError bool
)

func init() {
	sdksim.GetSimulatorFlags()
	// State listening flags
	flag.StringVar(&StateListeningPlugin, "StateListeningPlugin", "", "State listening plugin name")
	flag.BoolVar(&HaltAppOnDeliveryError, "HaltAppOnDeliveryError", true, "Halt app when state listeners fail")
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
	PrintConfig(config)
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

	PrintStats(config, db)
}

func TestSimple(t *testing.T) {
	config, db, dir, logger, skip, err := sdksim.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping provenance application simulation")
	}
	PrintConfig(config)
	require.NoError(t, err, "provenance simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), sdksim.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "provenanced", app.Name())

	// run randomized simulation
	_, _, simErr := simulation.SimulateFromSeed(
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

	require.NoError(t, simErr)
	PrintStats(config, db)
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
	PrintConfig(config)
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

	PrintStats(config, db)

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
	PrintConfig(config)
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

	PrintStats(config, db)

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
	config.DBBackend = "memdb"

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()
		PrintConfig(config)

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

			PrintStats(config, db)

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

// TestAppStateDeterminismStateListening non-deterministic
// testing with state listing indexing plugins enabled
func TestAppStateDeterminismWithStateListening(t *testing.T) {
	if !sdksim.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	if StateListeningPlugin == "" {
		t.Skip("state listening plugin flag not provided: -StateListeningPlugin=name")
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
		PrintConfig(config)

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if sdksim.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			// load listening plugin(s)
			appOpts := loadAppOptions()
			key := fmt.Sprintf("%s.%s", plugin.PLUGINS_TOML_KEY, plugin.PLUGINS_ENABLED_TOML_KEY)
			enabledPlugins := cast.ToStringSlice(appOpts.Get(key))
			for _, p := range enabledPlugins {
				if kafkaplugin.PLUGIN_NAME == p {
					prepKafkaTopics(appOpts)
					break
				}
			}

			db := dbm.NewMemDB()
			app := New(logger,
				db,
				nil,
				true, map[int64]bool{},
				DefaultNodeHome,
				sdksim.FlagPeriodValue,
				MakeEncodingConfig(),
				//sdksim.EmptyAppOptions{},
				appOpts,
				interBlockCacheOpt(),
			)

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
				PrintStats(config, db)
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

func loadAppOptions() types.AppOptions {
	// load plugin config
	keys := make([]string, 0) // leave empty to listen to all store keys
	m := make(map[string]interface{})
	m["plugins.on"] = true
	m["plugins.enabled"] = []string{StateListeningPlugin}
	m["plugins.dir"] = ""
	// file plugin
	m["plugins.streaming.file.keys"] = keys
	m["plugins.streaming.file.write_dir"] = ""
	m["plugins.streaming.file.prefix"] = ""
	m["plugins.streaming.file.halt_app_on_delivery_error"] = HaltAppOnDeliveryError
	// trace plugin
	m["plugins.streaming.trace.keys"] = keys
	m["plugins.streaming.trace.print_data_to_stdout"] = false
	m["plugins.streaming.trace.halt_app_on_delivery_error"] = HaltAppOnDeliveryError
	// kafka plugin
	m["plugins.streaming.kafka.keys"] = keys
	m["plugins.streaming.kafka.topic_prefix"] = "sim"
	m["plugins.streaming.kafka.flush_timeout_ms"] = 5000
	m["plugins.streaming.kafka.halt_app_on_delivery_error"] = HaltAppOnDeliveryError
	// Kafka plugin producer
	m["plugins.streaming.kafka.producer.bootstrap_servers"] = "localhost:9092"
	m["plugins.streaming.kafka.producer.client_id"] = "pio-sim"
	m["plugins.streaming.kafka.producer.acks"] = "all"
	m["plugins.streaming.kafka.producer.enable_idempotence"] = true

	vpr := viper.New()
	for key, value := range m {
		vpr.SetDefault(key, value)
	}

	return vpr
}

func prepKafkaTopics(opts types.AppOptions) {
	// kafka topic setup
	topicPrefix := cast.ToString(opts.Get(fmt.Sprintf("%s.%s.%s.%s", plugin.PLUGINS_TOML_KEY, plugin.STREAMING_TOML_KEY, kafkaplugin.PLUGIN_NAME, kafkaplugin.TOPIC_PREFIX_PARAM)))
	bootstrapServers := cast.ToString(opts.Get(fmt.Sprintf("%s.%s.%s.%s.%s", plugin.PLUGINS_TOML_KEY, plugin.STREAMING_TOML_KEY, kafkaplugin.PLUGIN_NAME, kafkaplugin.PRODUCER_CONFIG_PARAM, "bootstrap_servers")))
	bootstrapServers = strings.ReplaceAll(bootstrapServers, "_", ".")
	topics := []string{
		string(kafkaservice.BeginBlockReqTopic),
		kafkaservice.BeginBlockResTopic,
		kafkaservice.DeliverTxReqTopic,
		kafkaservice.DeliverTxResTopic,
		kafkaservice.EndBlockReqTopic,
		kafkaservice.EndBlockResTopic,
		kafkaservice.StateChangeTopic,
	}
	deleteTopics(topicPrefix, topics, bootstrapServers)
	createTopics(topicPrefix, topics, bootstrapServers)
}

func createTopics(topicPrefix string, topics []string, bootstrapServers string) {

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers":       bootstrapServers,
		"broker.version.fallback": "0.10.0.0",
		"api.version.fallback.ms": 0,
	})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		tmos.Exit(err.Error())
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDuration, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Printf("time.ParseDuration(60s)")
		tmos.Exit(err.Error())
	}

	var _topics []kafka.TopicSpecification
	for _, s := range topics {
		_topics = append(_topics,
			kafka.TopicSpecification{
				Topic:             fmt.Sprintf("%s-%s", topicPrefix, s),
				NumPartitions:     1,
				ReplicationFactor: 1})
	}
	results, err := adminClient.CreateTopics(ctx, _topics, kafka.SetAdminOperationTimeout(maxDuration))
	if err != nil {
		fmt.Printf("Problem during the topicPrefix creation: %v\n", err)
		tmos.Exit(err.Error())
	}

	// Check for specific topicPrefix errors
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError &&
			result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("Topic creation failed for %s: %v",
				result.Topic, result.Error.String())
			tmos.Exit(err.Error())
		}
	}

	adminClient.Close()
}

func deleteTopics(topicPrefix string, topics []string, bootstrapServers string) {
	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		tmos.Exit(err.Error())
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Delete topics on cluster
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Printf("ParseDuration(60s)")
		tmos.Exit(err.Error())
	}

	var _topics []string
	for _, s := range topics {
		_topics = append(_topics, fmt.Sprintf("%s-%s", topicPrefix, s))
	}

	results, err := a.DeleteTopics(ctx, _topics, kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to delete topics: %v\n", err)
		tmos.Exit(err.Error())
	}

	// Print results
	for _, result := range results {
		fmt.Printf("%s\n", result)
	}

	a.Close()
}

// PrintStats outputs the config and db info.
func PrintStats(config simtypes.Config, db dbm.DB) {
	PrintConfig(config)
	if config.Commit {
		PrintDBInfo(db)
	}
}

// PrintConfig outputs the config.
func PrintConfig(config simtypes.Config) {
	fmt.Println("-vvv-  Config Info  -vvv-")
	cfields := cmdconfig.MakeFieldValueMap(config, true)
	for _, f := range cfields.GetSortedKeys() {
		fmt.Printf("%s: %s\n", f, cfields.GetStringOf(f))
	}
	fmt.Println("-^^^-  Config Info  -^^^-")
}

// PrintDBInfo outputs the db.Stats map.
func PrintDBInfo(db dbm.DB) {
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
