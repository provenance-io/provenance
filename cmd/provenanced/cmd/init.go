// Package cmd contains provenance daemon init functionality.
// nolint
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/provenance-io/provenance/app"

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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"

	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagRecover defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"

	flagClientHome = "home-client"
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", string(sdk.MustSortJSON(out)))

	return err
}

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
			cdc := clientCtx.JSONCodec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.Moniker = args[0]
			config.SetRoot(clientCtx.HomeDir)
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			return Init(cmd, config, cdc, mbm, chainID)
		},
	}
	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	// It looks like we're not paying attention to this --minimum-gas-prices flag, but we are.
	// Viper sees the flag and uses it when creating the app.config file.
	// That file is created (if it doesn't yet exist) as part of the cobra pre-run handler stuff.
	// See cmd/provenanced/config/server.go#interceptConfigs.
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("1905%s", app.DefaultFeeDenom), "Minimum gas prices to accept for transactions")
	cmd.Flags().BoolP(FlagRecover, "r", false, "interactive key recovery from mnemonic")
	return cmd
}

// Init initializes genesis config.
func Init(
	cmd *cobra.Command,
	config *tmconfig.Config,
	cdc codec.JSONCodec,
	mbm module.BasicManager,
	chainID string,
) error {
	if chainID == "" {
		chainID = "provenance-chain-" + tmrand.NewRand().Str(6)
	}
	// Get bip39 mnemonic
	var mnemonic string
	if r, _ := cmd.Flags().GetBool(FlagRecover); r {
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
	genFile := config.GenesisFile()
	if !viper.GetBool(FlagOverwrite) && tmos.FileExists(genFile) {
		return fmt.Errorf("genesis.json file already exists: %v", genFile)
	}
	if err = initGenFile(cdc, mbm, config.Moniker, chainID, nodeID, genFile); err != nil {
		return err
	}
	tmconfig.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	return nil
}

func initGenFile(cdc codec.JSONCodec, mbm module.BasicManager, moniker, chainID, nodeID string, genFile string) error {
	appGenState := mbm.DefaultGenesis(cdc)
	// Set the mint parameters
	var mintGenState minttypes.GenesisState
	//cdc.MustUnmarshalJSON(appGenState[mint.ModuleName], &mintGenState)
	mintGenState.Minter.Inflation = sdk.ZeroDec()
	mintGenState.Minter.AnnualProvisions = sdk.OneDec()
	mintGenState.Params.MintDenom = app.DefaultBondDenom
	mintGenState.Params.InflationMax = sdk.ZeroDec()
	mintGenState.Params.InflationMin = sdk.ZeroDec()
	mintGenState.Params.InflationRateChange = sdk.OneDec()
	mintGenState.Params.GoalBonded = sdk.OneDec()
	mintGenState.Params.BlocksPerYear = 6311520 // (86400 / 5) * 365.25
	appGenState[minttypes.ModuleName] = cdc.MustMarshalJSON(&mintGenState)
	// Set the staking denom
	var stakeGenState stakingtypes.GenesisState
	cdc.MustUnmarshalJSON(appGenState[stakingtypes.ModuleName], &stakeGenState)
	stakeGenState.Params.BondDenom = app.DefaultBondDenom
	appGenState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(&stakeGenState)
	// Set the crisis denom
	var crisisGenState crisistypes.GenesisState
	cdc.MustUnmarshalJSON(appGenState[crisistypes.ModuleName], &crisisGenState)
	crisisGenState.ConstantFee.Denom = app.DefaultBondDenom
	appGenState[crisistypes.ModuleName] = cdc.MustMarshalJSON(&crisisGenState)
	// Set the gov depost denom
	var govGenState govtypes.GenesisState
	cdc.MustUnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState)
	govGenState.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(app.DefaultBondDenom, sdk.NewInt(10000000)))
	appGenState[govtypes.ModuleName] = cdc.MustMarshalJSON(&govGenState)
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

	if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
		return errors.Wrap(err, "Failed to export gensis file")
	}

	toPrint := newPrintInfo(moniker, chainID, nodeID, "", appState)

	return displayInfo(toPrint)
}
