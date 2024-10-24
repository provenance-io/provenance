package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	cmtconfig "github.com/cometbft/cometbft/config"
	cmtcli "github.com/cometbft/cometbft/libs/cli"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
)

// NewRootCmd creates a new root command for provenanced. It is called once in the main function.
// Providing sealConfig = false is only for unit tests that want to run multiple commands.
func NewRootCmd(sealConfig bool) (*cobra.Command, params.EncodingConfig) {
	// tempDir creates a temporary home directory.
	tempDir, err := os.MkdirTemp("", "provenanced")
	if err != nil {
		panic(fmt.Errorf("failed to create temp dir: %w", err))
	}
	defer os.RemoveAll(tempDir)

	// These are added to prevent invalid address caching from tempApp.
	sdk.SetAddrCacheEnabled(false)
	defer sdk.SetAddrCacheEnabled(true)

	app.SetConfig(isTestnetFlagSet(os.Args[1:]), false)

	tempApp := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(tempDir))
	encodingConfig := tempApp.GetEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(app.DefaultNodeHome).
		WithViper(app.EnvPrefix)
	sdk.SetCoinDenomRegex(app.SdkCoinDenomRegex)

	rootCmd := &cobra.Command{
		Use:   "provenanced",
		Short: "Provenance Blockchain App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContext(cmd, initClientCtx); err != nil {
				return err
			}
			if err := config.InterceptConfigsPreRunHandler(cmd); err != nil {
				return err
			}

			// set app context based on initialized EnvTypeFlag
			vpr := server.GetServerContextFromCmd(cmd).Viper
			testnet := vpr.GetBool(config.EnvTypeFlag)
			app.SetConfig(testnet, sealConfig)

			overwriteFlagDefaults(cmd, map[string]string{
				// Override default value for coin-type to match our mainnet or testnet value.
				config.CoinTypeFlag: fmt.Sprint(app.CoinType),
				// Override min gas price(server level config) here since the provenance config would have been set based on flags.
				server.FlagMinGasPrices: pioconfig.GetProvenanceConfig().ProvenanceMinGasPrices,
			})
			return nil
		},
	}

	initRootCmd(rootCmd, encodingConfig, tempApp.BasicModuleManager)
	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        "",
		flags.FlagKeyringBackend: "test",
		config.CoinTypeFlag:      fmt.Sprint(app.CoinTypeMainNet),
	})

	autoCliOpts := tempApp.AutoCliOpts()
	initClientCtx, _ = config.ApplyClientConfigToContext(initClientCtx, config.DefaultClientConfig())
	autoCliOpts.ClientCtx = initClientCtx

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	fixTxWasmInstantiate2Aliases(rootCmd)
	fixQueryWasmBuildAddressFlags(rootCmd)
	return rootCmd, encodingConfig
}

// Execute executes the root command.
func Execute(rootCmd *cobra.Command) error {
	// Create and set a client.Context on the command's Context. During the pre-run
	// of the root command, a default initialized client.Context is provided to
	// seed child command execution with values such as AccountRetriver, Keyring,
	// and a CometBFT RPC. This requires the use of a pointer reference when
	// getting and setting the client.Context. Ideally, we utilize
	// https://github.com/spf13/cobra/pull/1118.
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})
	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())

	rootCmd.PersistentFlags().BoolP(config.EnvTypeFlag, "t", false, "Indicates this command should use the testnet configuration (default: false [mainnet])")
	rootCmd.PersistentFlags().String(flags.FlagLogLevel, zerolog.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
	rootCmd.PersistentFlags().String(flags.FlagLogFormat, cmtconfig.LogFormatPlain, "The logging format (json|plain)")

	// Custom denom flag added to root command
	rootCmd.PersistentFlags().String(config.CustomDenomFlag, "", "Indicates if a custom denom is to be used, and the name of it (default nhash)")
	// Custom msgFee floor price flag added to root command
	rootCmd.PersistentFlags().Int64(config.CustomMsgFeeFloorPriceFlag, 0, "Custom msgfee floor price, optional (default 1905)")

	executor := cmtcli.PrepareBaseCmd(rootCmd, "", app.DefaultNodeHome)
	return executor.ExecuteContext(ctx)
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig, basicManager module.BasicManager) {
	rootCmd.AddCommand(
		InitCmd(basicManager),
		GenesisCmd(encodingConfig.TxConfig, basicManager, app.DefaultNodeHome),
		testnetCmd(basicManager, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		ConfigCmd(),
		AddMetaAddressCmd(),
		snapshot.Cmd(newApp),
		GetPreUpgradeCmd(),
		GetDocGenCmd(),
		GetTreeCmd(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
	)

	fixDebugPubkeyRawTypeFlag(rootCmd)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, createAppAndExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)

	// Disable usage when the start command returns an error.
	startCmd, _, err := rootCmd.Find([]string{"start"})
	if err != nil {
		panic(fmt.Errorf("start command not found: %w", err))
	}
	startCmd.SilenceUsage = true
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		rpc.QueryEventForTxCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")
	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		GetCmdPioSimulateTx(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// AppCreator func(log.Logger, dbm.DB, io.Writer, AppOptions) Application
func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	warnAboutSettings(logger, appOpts)

	var cache storetypes.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	// Create the snapshot dir if not exists
	if _, err = os.Stat(snapshotDir); os.IsNotExist(err) {
		err = os.Mkdir(snapshotDir, 0o755)
		if err != nil {
			panic(err)
		}
	}
	snapshotDB, err := dbm.NewDB("metadata", server.GetAppDBBackend(appOpts), snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	snapshotOptions := snapshottypes.NewSnapshotOptions(
		cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval)),
		cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent)),
	)

	homeDir := cast.ToString(appOpts.Get(flags.FlagHome))
	chainID := cast.ToString(appOpts.Get(flags.FlagChainID))
	if chainID == "" {
		// fallback to genesis chain-id
		reader, err := os.Open(filepath.Join(homeDir, "config", "genesis.json"))
		if err != nil {
			panic(err)
		}
		defer reader.Close()

		chainID, err = genutiltypes.ParseChainIDFromGenesis(reader)
		if err != nil {
			panic(fmt.Errorf("failed to parse chain-id from genesis file: %w", err))
		}
	}

	return app.New(
		logger, db, traceStore, true, appOpts,
		setStoreMetrics(getTelemetryGlobalLabels(logger, appOpts)),
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		baseapp.SetIAVLCacheSize(getIAVLCacheSize(appOpts)),
		baseapp.SetIAVLDisableFastNode(cast.ToBool(appOpts.Get(server.FlagDisableIAVLFastNode))),
		baseapp.SetChainID(chainID),
	)
}

// warnAboutSettings logs warnings about any settings that might cause problems.
func warnAboutSettings(logger log.Logger, appOpts servertypes.AppOptions) {
	defer func() {
		// If this func panics, just move on. It's just supposed to log things anyway.
		_ = recover()
	}()

	chainID := cast.ToString(appOpts.Get(flags.FlagChainID))
	if chainID == "pio-mainnet-1" {
		skipTimeoutCommit := cast.ToBool(appOpts.Get("consensus.skip_timeout_commit"))
		if !skipTimeoutCommit {
			timeoutCommit := cast.ToDuration(appOpts.Get("consensus.timeout_commit"))
			upperLimit := config.DefaultConsensusTimeoutCommit + 2*time.Second
			if timeoutCommit > upperLimit {
				logger.Error(fmt.Sprintf("Your consensus.timeout_commit config value is too high and should be decreased to at most %q. The recommended value is %q. Your current value is %q.",
					upperLimit, config.DefaultConsensusTimeoutCommit, timeoutCommit))
			}
		}
	}
}

func createAppAndExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var a *app.App

	if height != -1 {
		a = app.New(logger, db, traceStore, false, appOpts)

		if err := a.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		a = app.New(logger, db, traceStore, true, appOpts)
	}

	return a.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// fixDebugPubkeyRawTypeFlag removes the -t shorthand option of the --type flag from the debug pubkey-raw command.
// It clashes with our persistent --testnet/-t flag added to the root command.
func fixDebugPubkeyRawTypeFlag(rootCmd *cobra.Command) {
	debugCmd, _, err := rootCmd.Find([]string{"debug", "pubkey-raw"})
	if err != nil || debugCmd == nil {
		// If the command doesn't exist, there's nothing to do.
		return
	}

	// We can't just debugCmd.Flags().Lookup("type").Shorthand = "" here because
	// there's some internal flagset stuff that needs clearing too, that we can't
	// do from here. So we do this the hard way. We clear out the command's flags
	// and re-add them all, but get rid of the shorthand on the type flag first.
	debugFlags := debugCmd.Flags()
	debugCmd.ResetFlags()
	debugFlags.VisitAll(func(f *pflag.Flag) {
		if f.Name == "type" {
			f.Shorthand = ""
		}
		debugCmd.Flags().AddFlag(f)
	})
}

// fixTxWasmInstantiate2Aliases fixes the tx wasm instantiate2 aliases so that they're different
// from the instantiate aliases by adding a 2 to the end (if it doesn't yet have one).
func fixTxWasmInstantiate2Aliases(rootCmd *cobra.Command) {
	cmd, _, err := rootCmd.Find([]string{"tx", "wasm", "instantiate2"})
	if err != nil || cmd == nil {
		// If the command doesn't exist, there's nothing to do.
		return
	}

	// Go through all the aliases and make sure they end with a "2"
	for i, alias := range cmd.Aliases {
		if !strings.HasSuffix(alias, "2") {
			cmd.Aliases[i] = alias + "2"
		}
	}
}

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			if f.Changed {
				return
			}
			f.DefValue = val
			if err := f.Value.Set(val); err != nil {
				panic(err)
			}
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}

// getIAVLCacheSize If app.toml has an entry iavl-cache-size = <value>, default is 781250
func getIAVLCacheSize(options servertypes.AppOptions) int {
	iavlCacheSize := cast.ToInt(options.Get("iavl-cache-size"))
	if iavlCacheSize == 0 {
		// going with the DefaultConfig value(which is 781250),
		// the store default value is DefaultIAVLCacheSize = 500000
		return cast.ToInt(serverconfig.DefaultConfig().IAVLCacheSize)
	}
	return iavlCacheSize
}

// isTestnetFlagSet returns true if the args or env vars say to use testnet values.
// It differs from viper.Get("testnet") in that this will not look into any config files.
// That shouldn't be a problem because it's not in any of the templates. So someone would
// need to have it in a custom.toml file, which I don't think anyone utilizes.
//
// We need this because the SDK requires an address codec in order to create most of their
// cobra commands. Those commands need to be created and added to the root command well
// before we get to the point where we're ready to read in config files. But that address
// codec needs to have the HRP defined, which depends on the --testnet flag.
func isTestnetFlagSet(args []string) bool {
	// Go through the args backwards so that if multiple testnet flags are provided, the last one is used.
	for i := len(args) - 1; i >= 0; i-- {
		arg := args[i]
		if arg == "-t" || arg == "--testnet" {
			return true
		}
		if strings.HasPrefix(arg, "-t=") {
			return cast.ToBool(arg[3:])
		}
		if strings.HasPrefix(arg, "--testnet=") {
			return cast.ToBool(arg[10:])
		}
	}
	ev := os.Getenv("PIO_TESTNET")
	if len(ev) > 0 {
		return cast.ToBool(ev)
	}
	return false
}

// setStoreMetrics returns a baseapp option func that will set the store metrics if enabled.
func setStoreMetrics(globalLabels [][]string, enabled bool) func(*baseapp.BaseApp) {
	return func(bApp *baseapp.BaseApp) {
		if enabled {
			bApp.SetStoreMetrics(metrics.NewMetrics(globalLabels))
		}
	}
}

// These keys are used to get specific telemetry config values.
// The "telemetry" part is the mapstruct name of the Telemetry field in the serverconfig.Config struct.
// The second part is the mapstruct name of the desired field in the Telemetry.Config struct.
const (
	// telEnabledKey is the field name for the serverconfig.Config.Telemetry.Enabled field.
	telEnabledKey = "telemetry.enabled"
	// telGlobalLabelsKey is the field name for the serverconfig.Config.Telemetry.GlobalLabels field.
	telGlobalLabelsKey = "telemetry.global-labels"
)

// getTelemetryGlobalLabels will extract the telemetry.global-labels and telemetry.enabled info from the appOpts.
func getTelemetryGlobalLabels(logger log.Logger, appOpts servertypes.AppOptions) ([][]string, bool) {
	logger = logger.With("function", "getTelemetryGlobalLabels")
	if appOpts == nil {
		logger.Debug("No app options available.")
		return nil, false
	}

	// Most of the time appOpts is just a Viper instance, but there's no guarantee that it is.
	// If it is, we can use the SDK's stuff to identify the telemetry config values.
	// Otherwise, we'll try getting what we need directly from appOpts (which is less certain to be right).
	if vpr, ok := appOpts.(*viper.Viper); ok {
		appCfg, err := serverconfig.ParseConfig(vpr)
		if err == nil {
			logger.Debug("Using Telemetry config from app config.",
				telEnabledKey, appCfg.Telemetry.Enabled,
				telGlobalLabelsKey, appCfg.Telemetry.GlobalLabels)
			return appCfg.Telemetry.GlobalLabels, appCfg.Telemetry.Enabled
		}
		logger.Debug("Error parsing app config to get telemetry config (ignoring).", "error", err)
		// I've no clue what might cause that error, but let's try getting stuff the hard way.
	}

	// There's some unit tests that don't use viper for the appOpts, but this is mostly here in case
	// the SDK decides not to use viper for the appOpts in all situations that get us here.

	enabled := cast.ToBool(appOpts.Get(telEnabledKey))
	var globalLabels [][]string
	if enabled {
		if glRaw := appOpts.Get(telGlobalLabelsKey); glRaw != nil {
			var ok bool
			globalLabels, ok = glRaw.([][]string)
			if !ok {
				logger.Debug(fmt.Sprintf("App option %s is not [][]string, is %T", telGlobalLabelsKey, glRaw),
					"telemetry.global-labels", glRaw)
			}
		}
	}

	logger.Debug("Extracted telemetry setup from app options.", telEnabledKey, enabled, telGlobalLabelsKey, globalLabels)
	return globalLabels, enabled
}

// fixQueryWasmBuildAddressFlags resolves an issue where the Wasm build-address query failed to recognize the node flag.
func fixQueryWasmBuildAddressFlags(rootCmd *cobra.Command) {
	// Find the "build-address" command.
	cmd, _, err := rootCmd.Find([]string{"query", "wasm", "build-address"})
	if err != nil || cmd == nil {
		// If the command doesn't exist, there's nothing to do.
		return
	}

	// Apply query flags to the command.
	flags.AddQueryFlagsToCmd(cmd)
}
