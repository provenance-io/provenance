// Package cmd contains provenance daemon init functionality.
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/provenance-io/provenance/app"
	clientconf "github.com/provenance-io/provenance/cmd/provenanced/config"
	markertypes "github.com/provenance-io/provenance/x/marker/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"

	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"

	"github.com/spf13/cobra"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagRecover defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"
)

// InitCmd generates genesis config.
func InitCmd(
	mbm module.BasicManager,
	defaultNodeHome string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize files for a provenance daemon node",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)
			config.Moniker = args[0]

			return Init(cmd, cdc, mbm, config)
		},
	}
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().BoolP(FlagRecover, "r", false, "interactive key recovery from mnemonic")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	// It looks like we're not paying attention to this --minimum-gas-prices flag, but we are.
	// Viper sees the flag and uses it when creating the app.config file.
	// That file is created (if it doesn't yet exist) as part of the cobra pre-run handler stuff.
	// See cmd/provenanced/config/server.go#interceptConfigs.
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("1905%s", app.DefaultFeeDenom), "Minimum gas prices to accept for transactions")
	return cmd
}

// Init initializes genesis config.
func Init(
	cmd *cobra.Command,
	cdc codec.JSONCodec,
	mbm module.BasicManager,
	config *tmconfig.Config,
) error {
	serverCtx := server.GetServerContextFromCmd(cmd)
	chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	isTestnet := serverCtx.Viper.GetBool(EnvTypeFlag)
	doRecover, _ := cmd.Flags().GetBool(FlagRecover)
	doOverwrite := serverCtx.Viper.GetBool(FlagOverwrite)

	genFile := config.GenesisFile()
	if !doOverwrite && tmos.FileExists(genFile) {
		return fmt.Errorf("genesis file already exists: %v", genFile)
	}

	if len(chainID) == 0 {
		chainID = "provenance-chain-" + tmrand.NewRand().Str(6)
	}

	mustFprintf(cmd.OutOrStdout(), "chain id: %s\n", chainID)

	// Get bip39 mnemonic
	var mnemonic string
	if doRecover {
		inBuf := bufio.NewReader(cmd.InOrStdin())
		var err error
		mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	}

	nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
	if err != nil {
		return err
	}

	mustFprintf(cmd.OutOrStdout(), "node id: %s\n", nodeID)

	if err = initGenFile(cmd, cdc, mbm, isTestnet, chainID, genFile); err != nil {
		return err
	}

	configFile := filepath.Join(config.RootDir, "config", "config.toml")
	tmconfig.WriteConfigFile(configFile, config)
	mustFprintf(cmd.OutOrStdout(), "Tendermint config file updated: %s\n", configFile)

	clientConfigFile := filepath.Join(config.RootDir, "config", "client.toml")
	clientConfig, crerr := clientconf.GetClientConfig(clientConfigFile, client.GetClientContextFromCmd(cmd).Viper)
	if crerr != nil {
		return fmt.Errorf("couldn't get client config: %v", crerr)
	}
	clientConfig.ChainID = chainID
	clientConfig.Node = config.RPC.ListenAddress
	clientconf.WriteConfigToFile(clientConfigFile, clientConfig)
	mustFprintf(cmd.OutOrStdout(), "Client config file updated: %s\n", clientConfigFile)

	return nil
}

func initGenFile(
	cmd *cobra.Command,
	cdc codec.JSONCodec,
	mbm module.BasicManager,
	isTestnet bool,
	chainID, genFile string,
) error {
	minDeposit := int64(1000000000000)  // 1,000,000,000,000
	downtimeJailDurationStr := "86400s" // 1 day
	maxValidators := uint32(20)
	maxGas := int64(60000000) // 60,000,000
	unrestrictedDenomRegex := `[a-zA-Z][a-zA-Z0-9\-\.]{7,64}`
	if isTestnet {
		mustFprintf(cmd.OutOrStdout(), "Using testnet defaults\n")
		minDeposit = 10000000            // 10,000,000
		downtimeJailDurationStr = "600s" // 10 minutes
		maxValidators = 100
		maxGas = -1
		unrestrictedDenomRegex = `[a-zA-Z][a-zA-Z0-9\-\.]{2,64}`
	} else {
		mustFprintf(cmd.OutOrStdout(), "Using mainnet defaults\n")
	}

	downtimeJailDuration, err := time.ParseDuration(downtimeJailDurationStr)
	if err != nil {
		// This is a panic instead of a return because it should only fail if the strings above are defined incorrectly.
		// i.e. It's not an error that should be handled elsewhere. It means you need to fix your shit.
		panic(err)
	}

	appGenState := mbm.DefaultGenesis(cdc)

	// Note: The extra enclosures in here are just a defensive measure to hopefully help prevent
	// things from one section from being used in another when they're copy/pasted around.

	// Set the mint parameters
	{
		moduleName := minttypes.ModuleName
		var mintGenState minttypes.GenesisState
		cdc.MustUnmarshalJSON(appGenState[moduleName], &mintGenState)
		mintGenState.Minter.Inflation = sdk.ZeroDec()
		mintGenState.Minter.AnnualProvisions = sdk.OneDec()
		mintGenState.Params.MintDenom = app.DefaultBondDenom
		mintGenState.Params.InflationMax = sdk.ZeroDec()
		mintGenState.Params.InflationMin = sdk.ZeroDec()
		mintGenState.Params.InflationRateChange = sdk.OneDec()
		mintGenState.Params.GoalBonded = sdk.OneDec()
		mintGenState.Params.BlocksPerYear = 6311520 // (86400 / 5) * 365.25
		appGenState[moduleName] = cdc.MustMarshalJSON(&mintGenState)
	}

	// Set the staking denom
	{
		moduleName := stakingtypes.ModuleName
		var stakeGenState stakingtypes.GenesisState
		cdc.MustUnmarshalJSON(appGenState[moduleName], &stakeGenState)
		stakeGenState.Params.BondDenom = app.DefaultBondDenom
		appGenState[moduleName] = cdc.MustMarshalJSON(&stakeGenState)
	}

	// Set the crisis denom
	{
		moduleName := crisistypes.ModuleName
		var crisisGenState crisistypes.GenesisState
		cdc.MustUnmarshalJSON(appGenState[moduleName], &crisisGenState)
		crisisGenState.ConstantFee.Denom = app.DefaultBondDenom
		appGenState[moduleName] = cdc.MustMarshalJSON(&crisisGenState)
	}

	// Set the gov depost denom
	{
		moduleName := govtypes.ModuleName
		var govGenState govtypes.GenesisState
		cdc.MustUnmarshalJSON(appGenState[moduleName], &govGenState)
		govGenState.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(app.DefaultBondDenom, sdk.NewInt(minDeposit)))
		appGenState[moduleName] = cdc.MustMarshalJSON(&govGenState)
	}

	// Set slashing stuff.
	{
		moduleName := slashingtypes.ModuleName
		var slashingGenState slashingtypes.GenesisState
		cdc.MustUnmarshalJSON(appGenState[moduleName], &slashingGenState)
		slashingGenState.Params.DowntimeJailDuration = downtimeJailDuration
		appGenState[moduleName] = cdc.MustMarshalJSON(&slashingGenState)
	}

	// Set some staking stuff too.
	{
		moduleName := stakingtypes.ModuleName
		var stakingGenState stakingtypes.GenesisState
		cdc.MustUnmarshalJSON(appGenState[moduleName], &stakingGenState)
		stakingGenState.Params.MaxValidators = maxValidators
		appGenState[moduleName] = cdc.MustMarshalJSON(&stakingGenState)
	}

	// Set the marker unrestricted denom regex
	{
		moduleName := markertypes.ModuleName
		var markerGenState markertypes.GenesisState
		cdc.MustUnmarshalJSON(appGenState[moduleName], &markerGenState)
		markerGenState.Params.UnrestrictedDenomRegex = unrestrictedDenomRegex
		appGenState[moduleName] = cdc.MustMarshalJSON(&markerGenState)
	}

	appState, err := json.MarshalIndent(appGenState, "", "")
	if err != nil {
		return err
	}

	genDoc := &types.GenesisDoc{}
	if _, err = os.Stat(genFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		genDoc, err = types.GenesisDocFromFile(genFile)
		if err != nil {
			return errors.Wrap(err, "Failed to read genesis doc from file")
		}
	}

	genDoc.ChainID = chainID
	genDoc.AppState = appState
	genDoc.Validators = nil
	genDoc.ConsensusParams = types.DefaultConsensusParams()
	genDoc.ConsensusParams.Block.MaxGas = maxGas

	if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
		return errors.Wrap(err, "Failed to export gensis file")
	}

	mustFprintf(cmd.OutOrStdout(), "Genesis file created: %s\n", genFile)
	return err
}

// mustFprintf same as fmt.Fprintf but panics on error.
func mustFprintf(w io.Writer, format string, a ...interface{}) int {
	r, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		panic(err)
	}
	return r
}
