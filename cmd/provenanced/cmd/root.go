package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode"

	"github.com/rs/zerolog"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
)

const (
	// EnvTypeFlag is a flag for indicating a testnet
	EnvTypeFlag = "testnet"
	// Flag used to indicate coin type.
	CoinTypeFlag = "coin-type"
)

// ChainID is the id of the running chain
var ChainID string

// NewRootCmd creates a new root command for provenanced. It is called once in the main function.
// Providing sealConfig = false is only for unit tests that want to run multiple commands.
func NewRootCmd(sealConfig bool) (*cobra.Command, params.EncodingConfig) {
	encodingConfig := app.MakeEncodingConfig()
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("PIO")
	sdk.SetCoinDenomRegex(app.SdkCoinDenomRegex)
	rootCmd := &cobra.Command{
		Use:   "provenanced",
		Short: "Provenance Blockchain App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			if cmd.Flags().Changed(flags.FlagHome) {
				rootDir, _ := cmd.Flags().GetString(flags.FlagHome)
				initClientCtx = initClientCtx.WithHomeDir(rootDir)
			}
			if err := client.SetCmdClientContext(cmd, initClientCtx); err != nil {
				return err
			}
			if err := config.InterceptConfigsPreRunHandler(cmd); err != nil {
				return err
			}

			// set app context based on initialized EnvTypeFlag
			vpr := server.GetServerContextFromCmd(cmd).Viper
			testnet := vpr.GetBool(EnvTypeFlag)
			app.SetConfig(testnet, sealConfig)

			// Store the chain id in a global variable.
			ChainID = vpr.GetString(flags.FlagChainID)

			overwriteFlagDefaults(cmd, map[string]string{
				// Override default value for coin-type to match our mainnet or testnet value.
				CoinTypeFlag: fmt.Sprint(app.CoinType),
				// Override min gas price(server level config) here since the provenance config would have been set based on flags.
				server.FlagMinGasPrices: pioconfig.GetProvenanceConfig().ProvenanceMinGasPrices,
			})
			return nil
		},
	}
	initRootCmd(rootCmd, encodingConfig)
	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        ChainID,
		flags.FlagKeyringBackend: "test",
		CoinTypeFlag:             fmt.Sprint(app.CoinTypeMainNet),
	})

	return rootCmd, encodingConfig
}

// Execute executes the root command.
func Execute(rootCmd *cobra.Command) error {
	// Create and set a client.Context on the command's Context. During the pre-run
	// of the root command, a default initialized client.Context is provided to
	// seed child command execution with values such as AccountRetriver, Keyring,
	// and a Tendermint RPC. This requires the use of a pointer reference when
	// getting and setting the client.Context. Ideally, we utilize
	// https://github.com/spf13/cobra/pull/1118.
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})
	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())

	rootCmd.PersistentFlags().BoolP(EnvTypeFlag, "t", false, "Indicates this command should use the testnet configuration (default: false [mainnet])")
	rootCmd.PersistentFlags().String(flags.FlagLogLevel, zerolog.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
	rootCmd.PersistentFlags().String(flags.FlagLogFormat, tmcfg.LogFormatPlain, "The logging format (json|plain)")

	// Custom denom flag added to root command
	rootCmd.PersistentFlags().String(config.CustomDenomFlag, "", "Indicates if a custom denom is to be used, and the name of it (default nhash)")
	// Custom msgFee floor price flag added to root command
	rootCmd.PersistentFlags().Int64(config.CustomMsgFeeFloorPriceFlag, 0, "Custom msgfee floor price, optional (default 1905)")

	executor := tmcli.PrepareBaseCmd(rootCmd, "", app.DefaultNodeHome)
	// Add the --help flag now so that its usage can be capitalized.
	// Otherwise, it gets added by cobra at the last second.
	// And for some reason, running the root_test in an IDE doesn't see it, but
	// running with make test causes a failure.
	rootCmd.InitDefaultHelpFlag()
	capitalizeUses(rootCmd)

	return executor.ExecuteContext(ctx)
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig) {
	rootCmd.AddCommand(
		InitCmd(app.ModuleBasics),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		AddRootDomainAccountCmd(app.DefaultNodeHome),
		AddGenesisMarkerCmd(app.DefaultNodeHome),
		AddGenesisMsgFeeCmd(app.DefaultNodeHome, encodingConfig.InterfaceRegistry),
		AddGenesisCustomFloorPriceDenomCmd(app.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		testnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		ConfigCmd(),
		AddMetaAddressCmd(),
		snapshot.Cmd(newApp),
		GetPreUpgradeCmd(),
	)

	fixDebugPubkeyRawTypeFlag(rootCmd)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, createAppAndExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(app.DefaultNodeHome),
	)

	// Add Rosetta command
	rootCmd.AddCommand(server.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Marshaler))
	rootCmd.AddCommand(pruning.PruningCmd(newApp))

	// Disable usage when the start command returns an error.
	startCmd, _, err := rootCmd.Find([]string{"start"})
	if err != nil {
		panic(fmt.Errorf("start command not found: %w", err))
	}
	startCmd.SilenceUsage = true
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
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	app.ModuleBasics.AddQueryCommands(cmd)
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
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		GetCmdPioSimulateTx(),
		flags.LineBreak,
	)

	app.ModuleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	if ChainID == "pio-mainnet-1" {
		timeoutCommit := cast.ToDuration(appOpts.Get("consensus.timeout_commit"))
		if timeoutCommit < config.DefaultConsensusTimeoutCommit/2 {
			logger.Error(fmt.Sprintf("Your consensus.timeout_commit config value is too low and should be increased to %q (it is currently %q).",
				config.DefaultConsensusTimeoutCommit, timeoutCommit))
		}
	}

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
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

	return app.New(
		logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		app.MakeEncodingConfig(), // Ideally, we would reuse the one created by NewRootCmd.
		appOpts,
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
		baseapp.SetIAVLLazyLoading(cast.ToBool(appOpts.Get(server.FlagIAVLLazyLoading))),
	)
}

func createAppAndExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
) (servertypes.ExportedApp, error) {
	encCfg := app.MakeEncodingConfig() // Ideally, we would reuse the one created by NewRootCmd.
	encCfg.Marshaler = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	var a *app.App

	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	if height != -1 {
		a = app.New(logger, db, traceStore, false, map[int64]bool{}, homePath, cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)), encCfg, appOpts)

		if err := a.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		a = app.New(logger, db, traceStore, true, map[int64]bool{}, homePath, cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)), encCfg, appOpts)
	}

	return a.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
}

// setFlagDefault sets the default value of a flag if it hasn't been changed.
func setFlagDefault(s *pflag.FlagSet, key, val string) {
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

// overwriteFlagDefaults recursively overwrites the default values of flags
// with the provided defaults for this command and all sub-commands.
func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	for key, val := range defaults {
		setFlagDefault(c.Flags(), key, val)
		setFlagDefault(c.PersistentFlags(), key, val)
	}

	for _, sc := range c.Commands() {
		overwriteFlagDefaults(sc, defaults)
	}
}

// capitalizeFirst returns the provided string with the first letter capitalized.
func capitalizeFirst(str string) string {
	if len(str) == 0 {
		return str
	}
	// str[0] only gets the first byte, but we want the first rune (might be
	// multiple bytes, we don't know). So we use the hack where range on a
	// string loops over runes (not bytes), and break after getting the first.
	// Just getting the first, for now, since we commonly won't need the rest.
	var first rune
	for _, r := range str {
		first = r
		break
	}
	if !unicode.IsLower(first) {
		return str
	}
	// Alright. We need to upper-case the first one. We'll need all the runes so we can put it back together correctly.
	runes := make([]rune, 0, len(str))
	for _, r := range str {
		runes = append(runes, r)
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// capitalizeFlagUsage capitalizes the first letter of the flag's usage string.
func capitalizeFlagUsage(f *pflag.Flag) {
	f.Usage = capitalizeFirst(f.Usage)
}

// capitalizeUses recursively capitalizes the first letter of command short
// usage and all flag usages for this command and all sub-commands.
func capitalizeUses(c *cobra.Command) {
	c.Short = capitalizeFirst(c.Short)
	c.Flags().VisitAll(capitalizeFlagUsage)
	c.PersistentFlags().VisitAll(capitalizeFlagUsage)

	for _, sc := range c.Commands() {
		capitalizeUses(sc)
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
