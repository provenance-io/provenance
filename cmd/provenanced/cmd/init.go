// Package cmd contains provenance daemon init functionality.
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"

	"github.com/provenance-io/provenance/app"
	provconfig "github.com/provenance-io/provenance/cmd/provenanced/config"
	markertypes "github.com/provenance-io/provenance/x/marker/types"

	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagRecover defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"
)

// InitCmd Creates a command for generating genesis and config files.
func InitCmd(mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize files for a provenance daemon node",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args[0]) == 0 {
				return fmt.Errorf("no moniker provided")
			}
			return Init(cmd, mbm, args[0])
		},
	}
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().BoolP(FlagRecover, "r", false, "interactive key recovery from mnemonic")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	return cmd
}

// Init initializes genesis and config files.
func Init(
	cmd *cobra.Command,
	mbm module.BasicManager,
	moniker string,
) error {
	chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	isTestnet, _ := cmd.Flags().GetBool(EnvTypeFlag)
	doRecover, _ := cmd.Flags().GetBool(FlagRecover)
	doOverwrite, _ := cmd.Flags().GetBool(FlagOverwrite)

	if err := provconfig.EnsureConfigDir(cmd); err != nil {
		return err
	}

	// Get the current configs. This will just be all defaults if the files don't exist.
	appConfig, err := provconfig.ExtractAppConfig(cmd)
	if err != nil {
		return err
	}
	tmConfig, err := provconfig.ExtractTmConfig(cmd)
	if err != nil {
		return err
	}
	clientConfig, err := provconfig.ExtractClientConfig(cmd)
	if err != nil {
		return err
	}

	// Stop now if the genesis file already exists and an overwrite wasn't requested.
	genFile := tmConfig.GenesisFile()
	if !doOverwrite && tmos.FileExists(genFile) {
		return fmt.Errorf("genesis file already exists: %v", genFile)
	}

	// Set a few things in the configs.
	if len(appConfig.MinGasPrices) == 0 {
		appConfig.MinGasPrices = app.DefaultMinGasPrices
	}
	tmConfig.Moniker = moniker
	if len(chainID) == 0 {
		chainID = "provenance-chain-" + tmrand.NewRand().Str(6)
		cmd.Printf("chain id: %s\n", chainID)
	}
	clientConfig.ChainID = chainID

	// Gather the bip39 mnemonic if a recover was requested.
	var mnemonic string
	if doRecover {
		inBuf := bufio.NewReader(cmd.InOrStdin())
		mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	}

	// Create and write the node validator files (node_key.json and priv_validator_key.json).
	nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(tmConfig, mnemonic)
	if err != nil {
		return err
	}
	cmd.Printf("node id: %s\n", nodeID)
	clientConfig.Node = tmConfig.RPC.ListenAddress

	// Create and write the genenis file.
	if err = createAndExportGenesisFile(cmd, client.GetClientContextFromCmd(cmd).Codec, mbm, isTestnet, chainID, genFile); err != nil {
		return err
	}
	// Save the configs.
	provconfig.SaveConfigs(cmd, appConfig, tmConfig, clientConfig, true)

	return nil
}

// createAndExportGenesisFile creates and writes the genesis file.
func createAndExportGenesisFile(
	cmd *cobra.Command,
	cdc codec.JSONCodec,
	mbm module.BasicManager,
	isTestnet bool,
	chainID, genFile string,
) error {
	minDeposit := int64(1000000000000)  // 1,000,000,000,000
	downtimeJailDurationStr := "86400s" // 1 day
	maxGas := int64(60000000)           // 60,000,000
	if isTestnet {
		cmd.Printf("Using testnet defaults\n")
		minDeposit = 10000000            // 10,000,000
		downtimeJailDurationStr = "600s" // 10 minutes
	} else {
		cmd.Printf("Using mainnet defaults\n")
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

	// Set the gov deposit denom
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
		stakingGenState.Params.MaxValidators = 100
		appGenState[moduleName] = cdc.MustMarshalJSON(&stakingGenState)
	}

	// Set the marker unrestricted denom regex.
	// This is different from DefaultUnrestrictedDenomRegex.
	// That variable isn't updated because then the default test denom/bond ("stake")
	// doesn't pass and all sorts of tests needs fixing.
	{
		moduleName := markertypes.ModuleName
		var markerGenState markertypes.GenesisState
		cdc.MustUnmarshalJSON(appGenState[moduleName], &markerGenState)
		markerGenState.Params.UnrestrictedDenomRegex = `[a-zA-Z][a-zA-Z0-9\-\.]{7,83}`
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
	if genDoc.ConsensusParams == nil {
		genDoc.ConsensusParams = types.DefaultConsensusParams()
	}
	genDoc.ConsensusParams.Block.MaxGas = maxGas

	if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
		return errors.Wrap(err, "Failed to export gensis file")
	}

	cmd.Printf("Genesis file created: %s\n", genFile)
	return err
}
