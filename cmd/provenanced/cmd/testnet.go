package cmd

// DONTCOVER

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/version"

	"github.com/spf13/cobra"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/provenance-io/provenance/app"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

const (
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNumValidators     = "v"
	flagOutputDir         = "output-dir"
	flagNodeDaemonHome    = "node-daemon-home"
	flagStartingIPAddress = "starting-ip-address"
)

// get cmd to initialize all files for tendermint testnet and application
func testnetCmd(mbm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a simapp testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.
	`,
		Example: fmt.Sprintf(`%[1]s testnet --v 4 --output-dir ./output --starting-ip-address 192.168.20.2`, version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			outputDir, _ := cmd.Flags().GetString(flagOutputDir)
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			minGasPrices, _ := cmd.Flags().GetString(server.FlagMinGasPrices)
			nodeDirPrefix, _ := cmd.Flags().GetString(flagNodeDirPrefix)
			nodeDaemonHome, _ := cmd.Flags().GetString(flagNodeDaemonHome)
			startingIPAddress, _ := cmd.Flags().GetString(flagStartingIPAddress)
			numValidators, _ := cmd.Flags().GetInt(flagNumValidators)
			algo, _ := cmd.Flags().GetString(flags.FlagKeyAlgorithm)

			return InitTestnet(
				clientCtx, cmd, config, mbm, genBalIterator, outputDir, chainID, minGasPrices,
				nodeDirPrefix, nodeDaemonHome, startingIPAddress, keyringBackend, algo, numValidators,
			)
		},
	}

	cmd.Flags().Int(flagNumValidators, 4, "Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./mytestnet", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(server.FlagMinGasPrices, app.DefaultMinGasPrices, fmt.Sprintf("Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. %s,0.001stake)", app.DefaultMinGasPrices))
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().String(flags.FlagKeyAlgorithm, string(hd.Secp256k1Type), "Key signing algorithm to generate keys for")

	return cmd
}

const nodeDirPerm = 0755

// InitTestnet initializes the testnet
func InitTestnet(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator banktypes.GenesisBalancesIterator,
	outputDir,
	chainID,
	minGasPrices,
	nodeDirPrefix,
	nodeDaemonHome,
	startingIPAddress,
	keyringBackend,
	algoStr string,
	numValidators int,
) error {
	if chainID == "" {
		chainID = "chain-" + tmrand.NewRand().Str(6)
	}

	nodeIDs := make([]string, numValidators)
	valPubKeys := make([]cryptotypes.PubKey, numValidators)

	simappConfig := srvconfig.DefaultConfig()
	simappConfig.MinGasPrices = minGasPrices
	simappConfig.API.Enable = true
	simappConfig.Telemetry.Enabled = true
	simappConfig.Telemetry.PrometheusRetentionTime = 60
	simappConfig.Telemetry.EnableHostnameLabel = false
	simappConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", chainID}}

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genFiles    []string
		genMarkers  []markertypes.MarkerAccount
	)

	inBuf := bufio.NewReader(cmd.InOrStdin())
	// generate private keys, node IDs, and initial transactions
	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")

		nodeConfig.SetRoot(nodeDir)
		nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		nodeConfig.Moniker = nodeDirName

		ip, err := getIP(i, startingIPAddress)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(nodeConfig)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		memo := fmt.Sprintf("%s@%s:26656", nodeIDs[i], ip)
		genFiles = append(genFiles, nodeConfig.GenesisFile())

		kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, nodeDir, inBuf)
		if err != nil {
			return err
		}

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(algoStr, keyringAlgos)
		if err != nil {
			return err
		}

		addr, secret, err := server.GenerateSaveCoinKey(kb, nodeDirName, true, algo)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		info := map[string]string{"secret": secret}

		cliPrint, err := json.Marshal(info)
		if err != nil {
			return err
		}

		// save private key seed words
		if err = writeFile(fmt.Sprintf("%v.json", "key_seed"), nodeDir, cliPrint); err != nil {
			return err
		}

		hashAmt := sdk.NewInt(100_000_000_000 / int64(numValidators))
		convAmt := sdk.NewInt(1_000_000_000)
		nhashAmt := hashAmt.Mul(convAmt)
		coins := sdk.Coins{
			sdk.NewCoin(app.DefaultBondDenom, nhashAmt),
		}

		genBalances = append(genBalances, banktypes.Balance{Address: addr.String(), Coins: coins.Sort()})
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(addr, nil, 0, 0))

		valTokens := sdk.TokensFromConsensusPower(100, app.DefaultPowerReduction)
		createValMsg, _ := stakingtypes.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			valPubKeys[i],
			sdk.NewCoin(app.DefaultBondDenom, valTokens),
			stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
			stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
			sdk.OneInt(),
		)

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		if err = txBuilder.SetMsgs(createValMsg); err != nil {
			return err
		}

		txBuilder.SetMemo(memo)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(chainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(clientCtx.TxConfig)

		if err = tx.Sign(txFactory, nodeDirName, txBuilder, true); err != nil {
			return err
		}

		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
			return err
		}

		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), simappConfig)
	}

	markerAcc := markertypes.NewEmptyMarkerAccount(app.DefaultBondDenom, genAccounts[0].GetAddress().String(),
		[]markertypes.AccessGrant{
			*markertypes.NewAccessGrant(genAccounts[0].GetAddress(), []markertypes.Access{
				markertypes.Access_Admin,
				markertypes.Access_Mint,
				markertypes.Access_Burn,
				markertypes.Access_Withdraw,
			}),
		})

	if err := markerAcc.SetSupply(sdk.NewCoin(app.DefaultBondDenom, sdk.NewInt(100_000_000_000).Mul(sdk.NewInt(1_000_000_000)))); err != nil {
		return err
	}

	if err := markerAcc.SetStatus(markertypes.StatusActive); err != nil {
		return err
	}

	genMarkers = append(genMarkers, *markerAcc)

	if err := initGenFiles(clientCtx, mbm, chainID, genAccounts, genBalances, genMarkers, genFiles, numValidators); err != nil {
		return err
	}

	err := collectGenFiles(
		clientCtx, nodeConfig, chainID, nodeIDs, valPubKeys, numValidators,
		outputDir, nodeDirPrefix, nodeDaemonHome, genBalIterator,
	)
	if err != nil {
		return err
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", numValidators)
	return nil
}

func initGenFiles(
	clientCtx client.Context, mbm module.BasicManager, chainID string,
	genAccounts []authtypes.GenesisAccount, genBalances []banktypes.Balance,
	genMarkers []markertypes.MarkerAccount, genFiles []string, numValidators int,
) error {
	appGenState := mbm.DefaultGenesis(clientCtx.JSONCodec)

	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	clientCtx.JSONCodec.MustUnmarshalJSON(appGenState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = accounts
	appGenState[authtypes.ModuleName] = clientCtx.JSONCodec.MustMarshalJSON(&authGenState)

	// PROVENANCE SPECIFIC CONFIG
	// ----------------------------------------

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	clientCtx.JSONCodec.MustUnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState)

	bankGenState.Balances = banktypes.SanitizeGenesisBalances(genBalances)
	for _, bal := range bankGenState.Balances {
		bankGenState.Supply = bankGenState.Supply.Add(bal.Coins...)
	}

	nhashDenomUnit := banktypes.DenomUnit{
		Exponent: 0,
		Denom:    "nhash",
		Aliases:  []string{"nanohash"},
	}
	hashDenomUnit := banktypes.DenomUnit{
		Exponent: 9,
		Denom:    "hash",
	}
	denomUnits := []*banktypes.DenomUnit{
		&hashDenomUnit,
		&nhashDenomUnit,
	}
	denomMetadata := banktypes.Metadata{
		Description: "The native staking token of the Provenance Blockchain.",
		Display:     "hash",
		Base:        "nhash",
		DenomUnits:  denomUnits,
	}
	bankGenState.DenomMetadata = []banktypes.Metadata{denomMetadata}
	appGenState[banktypes.ModuleName] = clientCtx.JSONCodec.MustMarshalJSON(&bankGenState)

	// Set the staking denom
	var stakeGenState stakingtypes.GenesisState
	clientCtx.JSONCodec.MustUnmarshalJSON(appGenState[stakingtypes.ModuleName], &stakeGenState)
	stakeGenState.Params.BondDenom = app.DefaultBondDenom
	appGenState[stakingtypes.ModuleName] = clientCtx.JSONCodec.MustMarshalJSON(&stakeGenState)

	// Set the crisis denom
	var crisisGenState crisistypes.GenesisState
	clientCtx.JSONCodec.MustUnmarshalJSON(appGenState[crisistypes.ModuleName], &crisisGenState)
	crisisGenState.ConstantFee.Denom = app.DefaultBondDenom
	appGenState[crisistypes.ModuleName] = clientCtx.JSONCodec.MustMarshalJSON(&crisisGenState)

	// Set the gov depost denom
	var govGenState govtypes.GenesisState
	clientCtx.JSONCodec.MustUnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState)
	govGenState.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(app.DefaultBondDenom, sdk.NewInt(10000000)))
	govGenState.VotingParams.VotingPeriod, _ = time.ParseDuration("360s")
	appGenState[govtypes.ModuleName] = clientCtx.JSONCodec.MustMarshalJSON(&govGenState)

	// Set the mint module parameters to stop inflation on the BondDenom.
	var mintGenState minttypes.GenesisState
	clientCtx.JSONCodec.MustUnmarshalJSON(appGenState[minttypes.ModuleName], &mintGenState)
	mintGenState.Params.MintDenom = app.DefaultBondDenom
	mintGenState.Minter.AnnualProvisions = sdk.ZeroDec()
	mintGenState.Minter.Inflation = sdk.ZeroDec()
	mintGenState.Params.InflationMax = sdk.ZeroDec()
	mintGenState.Params.InflationMin = sdk.ZeroDec()
	appGenState[minttypes.ModuleName] = clientCtx.JSONCodec.MustMarshalJSON(&mintGenState)

	// Set the root names
	var nameGenState nametypes.GenesisState
	clientCtx.JSONCodec.MustUnmarshalJSON(appGenState[nametypes.ModuleName], &nameGenState)
	// create the four root example domains
	nameGenState.Bindings = append(nameGenState.Bindings, nametypes.NewNameRecord("pb", genAccounts[0].GetAddress(), true))
	nameGenState.Bindings = append(nameGenState.Bindings, nametypes.NewNameRecord("io", genAccounts[0].GetAddress(), true))
	nameGenState.Bindings = append(nameGenState.Bindings, nametypes.NewNameRecord("pio", genAccounts[0].GetAddress(), false))
	nameGenState.Bindings = append(nameGenState.Bindings, nametypes.NewNameRecord("provenance", genAccounts[0].GetAddress(), false))
	appGenState[nametypes.ModuleName] = clientCtx.JSONCodec.MustMarshalJSON(&nameGenState)

	// set markers
	var markerGenState markertypes.GenesisState
	clientCtx.JSONCodec.MustUnmarshalJSON(appGenState[markertypes.ModuleName], &markerGenState)
	markerGenState.Markers = genMarkers
	appGenState[markertypes.ModuleName] = clientCtx.JSONCodec.MustMarshalJSON(&markerGenState)

	// END OF PROVENANCE SPECIFIC CONFIG
	// --------------------------------------------

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:  chainID,
		AppState: appGenStateJSON,
		ConsensusParams: &tmproto.ConsensusParams{
			Block: tmproto.BlockParams{
				MaxBytes:   200000,
				MaxGas:     60_000_000,
				TimeIotaMs: 1000,
			},
			Evidence: tmproto.EvidenceParams{
				MaxAgeNumBlocks: 302400,
				MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
				MaxBytes:        10000,
			},
			Validator: tmproto.ValidatorParams{
				PubKeyTypes: []string{
					types.ABCIPubKeyTypeEd25519,
				},
			},
		},
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}
	return nil
}

func collectGenFiles(
	clientCtx client.Context, nodeConfig *tmconfig.Config, chainID string,
	nodeIDs []string, valPubKeys []cryptotypes.PubKey, numValidators int,
	outputDir, nodeDirPrefix, nodeDaemonHome string, genBalIterator banktypes.GenesisBalancesIterator,
) error {
	var appState json.RawMessage
	genTime := tmtime.Now()

	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")
		nodeConfig.Moniker = nodeDirName

		nodeConfig.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutiltypes.NewInitConfig(chainID, gentxsDir, nodeID, valPubKey)

		genDoc, err := types.GenesisDocFromFile(nodeConfig.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genutil.GenAppStateFromConfig(clientCtx.JSONCodec, clientCtx.TxConfig, nodeConfig, initCfg, *genDoc, genBalIterator)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := nodeConfig.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name string, dir string, contents []byte) error {
	file := filepath.Join(dir, name)

	err := tmos.EnsureDir(dir, 0755)
	if err != nil {
		return err
	}

	err = tmos.WriteFile(file, contents, 0644)
	if err != nil {
		return err
	}

	return nil
}
