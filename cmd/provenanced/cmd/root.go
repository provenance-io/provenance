package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cmtconfig "github.com/cometbft/cometbft/config"
	cmtcli "github.com/cometbft/cometbft/libs/cli"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
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
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	// rosettacmd "github.com/cosmos/rosetta/cmd" // TODO[1760]: rosetta

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
)

const (
	// EnvTypeFlag is a flag for indicating a testnet
	EnvTypeFlag = "testnet"
	// CoinTypeFlag is a flag for indicating coin type.
	CoinTypeFlag = "coin-type"
)

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
		WithBroadcastMode(flags.BroadcastSync). // TODO[1760]: broadcast: Verify that this is right since BroadcastBlock is gone.
		WithHomeDir(app.DefaultNodeHome).
		WithViper("PIO")
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

			// TODO[1760]: client: Check impact that this has on our custom client config.
			initClientCtx, err = clientconfig.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			// This needs to go after ReadFromClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL. This sign mode
			// is only available if the client is online.
			if !initClientCtx.Offline {
				enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
				txConfigOpts := tx.ConfigOptions{
					EnabledSignModes:           enabledSignModes,
					TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(initClientCtx),
				}
				txConfig, err := tx.NewTxConfigWithOptions(
					initClientCtx.Codec,
					txConfigOpts,
				)
				if err != nil {
					return err
				}

				initClientCtx = initClientCtx.WithTxConfig(txConfig)
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

			overwriteFlagDefaults(cmd, map[string]string{
				// Override default value for coin-type to match our mainnet or testnet value.
				CoinTypeFlag: fmt.Sprint(app.CoinType),
				// Override min gas price(server level config) here since the provenance config would have been set based on flags.
				server.FlagMinGasPrices: pioconfig.GetProvenanceConfig().ProvenanceMinGasPrices,
			})
			return nil
		},
	}
	genAutoCompleteCmd(rootCmd)
	initRootCmd(rootCmd, encodingConfig)
	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        "",
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
	rootCmd.PersistentFlags().String(flags.FlagLogFormat, cmtconfig.LogFormatPlain, "The logging format (json|plain)")

	// Custom denom flag added to root command
	rootCmd.PersistentFlags().String(config.CustomDenomFlag, "", "Indicates if a custom denom is to be used, and the name of it (default nhash)")
	// Custom msgFee floor price flag added to root command
	rootCmd.PersistentFlags().Int64(config.CustomMsgFeeFloorPriceFlag, 0, "Custom msgfee floor price, optional (default 1905)")

	executor := cmtcli.PrepareBaseCmd(rootCmd, "", app.DefaultNodeHome)
	return executor.ExecuteContext(ctx)
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig) {
	rootCmd.AddCommand(
		InitCmd(app.ModuleBasics),
		// genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome), // TODO[1760]: genutil
		// genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome), // TODO[1760]: genutil
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		AddRootDomainAccountCmd(app.DefaultNodeHome),
		AddGenesisMarkerCmd(app.DefaultNodeHome),
		AddGenesisMsgFeeCmd(app.DefaultNodeHome, encodingConfig.InterfaceRegistry),
		AddGenesisCustomFloorPriceDenomCmd(app.DefaultNodeHome),
		AddGenesisDefaultMarketCmd(app.DefaultNodeHome),
		AddGenesisCustomMarketCmd(app.DefaultNodeHome),
		cmtcli.NewCompletionCmd(rootCmd, true),
		testnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		ConfigCmd(),
		AddMetaAddressCmd(),
		snapshot.Cmd(newApp),
		GetPreUpgradeCmd(),
		GetDocGenCmd(),
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

	// Add Rosetta command
	// rootCmd.AddCommand(rosettacmd.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Marshaler)) // TODO[1760]: rosetta
	rootCmd.AddCommand(pruning.Cmd(newApp, app.DefaultNodeHome))
	// Disable usage when the start command returns an error.
	startCmd, _, err := rootCmd.Find([]string{"start"})
	if err != nil {
		panic(fmt.Errorf("start command not found: %w", err))
	}
	startCmd.SilenceUsage = true
}

// genAutoCompleteCmd creates the command for autocomplete.
func genAutoCompleteCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "enable-cli-autocomplete [bash|zsh|fish|powershell]",
		Short: "Generates autocomplete scripts for the provenanced binary",
		Long: `To configure your shell to load completions for each session, add to your profile:

# bash example
echo '. <(provenanced enable-cli-autocomplete bash)' >> ~/.bash_profile
source ~/.bash_profile

# zsh example
echo '. <(provenanced enable-cli-autocomplete zsh)' >> ~/.zshrc
source ~/.zshrc
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}

			return fmt.Errorf("shell %s is not supported", args[0])
		},
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
		// authcmd.GetAccountCmd(), // TODO[1760]: auto-cli: Figure out how to still have this.
		rpc.ValidatorCommand(),
		// rpc.BlockCommand(), // TODO[1760]: auto-cli: Figure out how to still have this.
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

// AppCreator func(log.Logger, dbm.DB, io.Writer, AppOptions) Application
func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	warnAboutSettings(logger, appOpts)

	var cache storetypes.MultiStorePersistentCache

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
