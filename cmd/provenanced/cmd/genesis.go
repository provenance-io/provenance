package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/provenance-io/provenance/x/exchange"
	exchangecli "github.com/provenance-io/provenance/x/exchange/client/cli"
	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
	markercli "github.com/provenance-io/provenance/x/marker/client/cli"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

const (
	flagRestricted = "restrict"

	flagManager  = "manager"
	flagAccess   = "access"
	flagEscrow   = "escrow"
	flagActivate = "activate"
	flagFinalize = "finalize"

	flagDenom = "denom"
)

// genCmdStart is the first part of the command that gets you to one of the ones in here.
var genCmdStart = fmt.Sprintf("%s genesis", version.AppName)

// GenesisCmd creates the genesis command with sub-commands for adding things to the genesis file.
func GenesisCmd(txConfig client.TxConfig, moduleBasics module.BasicManager, defaultNodeHome string) *cobra.Command {
	// This is similar to the SDK's x/genutil/client/cli/commands.go CommandsWithCustomMigrationMap function.
	// This does not have the migrate genesis command, though, and also has our own custom commands.
	cmd := &cobra.Command{
		Use:                        "genesis",
		Short:                      "Application's genesis-related subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	gentxModule := moduleBasics[genutiltypes.ModuleName].(genutil.AppModuleBasic)

	cmd.AddCommand(
		genutilcli.GenTxCmd(moduleBasics, txConfig, banktypes.GenesisBalancesIterator{}, defaultNodeHome, txConfig.SigningContext().ValidatorAddressCodec()),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, defaultNodeHome, gentxModule.GenTxValidator, txConfig.SigningContext().ValidatorAddressCodec()),
		genutilcli.ValidateGenesisCmd(moduleBasics),
		AddGenesisAccountCmd(txConfig, defaultNodeHome),
		AddRootDomainAccountCmd(defaultNodeHome),
		AddGenesisMarkerCmd(defaultNodeHome),
		AddGenesisFlatFeeCmd(defaultNodeHome),
		AddGenesisDefaultMarketCmd(defaultNodeHome),
		AddGenesisCustomMarketCmd(defaultNodeHome),
	)

	return cmd
}

// AddGenesisAccountCmd returns add-account cobra command.
func AddGenesisAccountCmd(txConfig client.TxConfig, defaultNodeHome string) *cobra.Command {
	// Change the "add-genesis-account" command to just "add-account" since it's under the genesis command already.
	// Also make sure the command can be invoked using either.
	cmd := genutilcli.AddGenesisAccountCmd(defaultNodeHome, txConfig.SigningContext().AddressCodec())

	addGenAcctNameOld := "add-genesis-account"
	addGenAcctNameNew := "add-account"

	switch {
	case strings.HasPrefix(cmd.Use, addGenAcctNameOld):
		// It's got the old name, we'll change it to the new name, make sure the
		// new name isn't an alias, and add the old name to the aliases.
		cmd.Use = addGenAcctNameNew + strings.TrimPrefix(cmd.Use, addGenAcctNameOld)
		aliases := make([]string, 0, len(cmd.Aliases)+1)
		for _, alias := range cmd.Aliases {
			if alias != addGenAcctNameNew {
				aliases = append(aliases, alias)
			}
		}
		aliases = append(aliases, addGenAcctNameOld)
		cmd.Aliases = aliases
	case strings.HasPrefix(cmd.Use, addGenAcctNameNew):
		// It's already the new name, just make sure the old name is in the aliases.
		hasOld := false
		for _, alias := range cmd.Aliases {
			if alias == addGenAcctNameOld {
				hasOld = true
				break
			}
		}
		if !hasOld {
			cmd.Aliases = append(cmd.Aliases, addGenAcctNameOld)
		}
	default:
		// It's some other name. Make sure both names are in the aliases
		hasNew, hasOld := false, false
		for _, alias := range cmd.Aliases {
			switch alias {
			case addGenAcctNameNew:
				hasNew = true
			case addGenAcctNameOld:
				hasOld = true
			}
			if hasNew && hasOld {
				break
			}
		}

		if !hasNew {
			cmd.Aliases = append(cmd.Aliases, addGenAcctNameNew)
		}
		if !hasOld {
			cmd.Aliases = append(cmd.Aliases, addGenAcctNameOld)
		}
	}

	return cmd
}

// appStateUpdater is a function that makes modifications to an app-state.
// Use one in conjunction with updateGenesisFileRunE if your command only needs to parse
// some inputs to make additions or changes to app-state,
type appStateUpdater func(clientCtx client.Context, cmd *cobra.Command, args []string, appState map[string]json.RawMessage) error

// updateGenesisFile reads the existing genesis file, runs the app-state through the
// provided updater, then saves the updated genesis state over the existing genesis file.
func updateGenesisFile(cmd *cobra.Command, args []string, updater appStateUpdater) error {
	clientCtx := client.GetClientContextFromCmd(cmd)
	serverCtx := server.GetServerContextFromCmd(cmd)
	config := serverCtx.Config
	config.SetRoot(clientCtx.HomeDir)
	genFile := config.GenesisFile()

	// Get existing gen state
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("failed to read genesis file: %w", err)
	}

	err = updater(clientCtx, cmd, args, appState)
	if err != nil {
		return err
	}

	// None of the possible errors that might come after this will be helped by printing usage with them.
	cmd.SilenceUsage = true

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}

	genDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(genDoc, genFile)
}

// updateGenesisFileRunE returns a cobra.Command.RunE function that runs updateGenesisFile using the provided updater.
func updateGenesisFileRunE(updater appStateUpdater) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return updateGenesisFile(cmd, args, updater)
	}
}

// AddRootDomainAccountCmd returns add-root-name cobra command.
func AddRootDomainAccountCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-root-name [address_or_key_name] [root-name]",
		Aliases: []string{"add-genesis-root-name"},
		Short:   "Add a name binding to genesis.json",
		Long: `Add a name binding to genesis.json. The provided account must specify
	the account address or key name and domain-name to bind. If a key name is given,
	the address will be looked up in the local Keybase.  The restricted flag can optionally be
	included to lock or unlock an entry to child names.
	`,
		Example: fmt.Sprintf(`$ %[1]s add-root-name mykey rootname`, genCmdStart),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			addr, parseErr := sdk.AccAddressFromBech32(args[0])
			if parseErr != nil {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				keyringBackend, err := cmd.Flags().GetString(flags.FlagKeyringBackend)
				if err != nil {
					return err
				}

				// attempt to lookup address from Keybase if no address was provided
				kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf, cdc)
				if err != nil {
					return err
				}

				info, err := kb.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keybase: %w, could use use address as bech32 string: %s", err, parseErr.Error())
				}

				addr, err = info.GetAddress()
				if err != nil {
					return fmt.Errorf("failed to keyring get address: %w", err)
				}
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			nameGenState := nametypes.GetGenesisStateFromAppState(cdc, appState)

			for _, nr := range nameGenState.Bindings {
				if nr.Name == strings.ToLower(strings.TrimSpace(args[1])) {
					return fmt.Errorf("cannot add name already exists: %s", args[1])
				}
			}

			isRestricted, err := cmd.Flags().GetBool(flagRestricted)
			if err != nil {
				return err
			}
			nameGenState.Bindings = append(nameGenState.Bindings, nametypes.NewNameRecord(args[1], addr, isRestricted))

			nameGenStateBz, err := cdc.MarshalJSON(nameGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal name genesis state: %w", err)
			}

			appState[nametypes.ModuleName] = nameGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().BoolP(flagRestricted, "r", true, "Whether to make the new name restricted")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// AddGenesisMarkerCmd returns add-marker cobra command.
func AddGenesisMarkerCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-marker [coin] --manager [address_or_key_name] --access [grant][,[grant]] --escrow [coin][, [coin]] --finalize --activate --type [COIN] --required-attributes=attr.one,*.attr.two,...",
		Aliases: []string{"add-genesis-marker"},
		Short:   "Adds a marker to genesis.json",
		Long: `Adds a marker to genesis.json. The provided parameters must specify
the marker supply and denom as a coin.  A managing account may be added as a key name or address. An accessgrant
may be assigned to the manager address. The escrowed list of initial tokens must contain valid denominations. If
a marker account is activated any unassigned marker supply must be provided as escrow. Finalized markers must have
a manager address assigned that can activate the marker after genesis.  Activated markers will have supply invariants
enforced immediately.  An optional type flag can be provided or the default of COIN will be used.
`,
		Example: fmt.Sprintf(`$ %[1]s add-marker 1000000000funcoins --manager validator --access withdraw --escrow 100funcoins --finalize --type COIN --required-attributes=attr.one,*.attr.two,...`, genCmdStart),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			// parse coin that represents marker denom and total supply
			markerCoin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse coins: %w", err)
			}

			var managerAddr sdk.AccAddress
			var accessGrants []markertypes.AccessGrant
			markerStatus := markertypes.StatusProposed

			mgr, err := cmd.Flags().GetString(flagManager)
			if err != nil {
				return err
			}
			if len(mgr) > 0 {
				managerAddr, err = sdk.AccAddressFromBech32(mgr)
				if err != nil {
					inBuf := bufio.NewReader(cmd.InOrStdin())
					keyringBackend, keyFlagError := cmd.Flags().GetString(flags.FlagKeyringBackend)
					if keyFlagError != nil {
						return keyFlagError
					}
					// attempt to lookup address from Keybase if no address was provided
					kb, newKeyringError := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf, cdc)
					if newKeyringError != nil {
						return newKeyringError
					}
					info, keyErr := kb.Key(mgr)
					if keyErr != nil {
						return fmt.Errorf("failed to get address from Keybase: %w", keyErr)
					}

					managerAddr, err = info.GetAddress()
					if err != nil {
						return fmt.Errorf("failed to keyring get address: %w", err)
					}
				}
			}
			acs, err := cmd.Flags().GetString(flagAccess)
			if err != nil {
				return err
			}
			if len(acs) > 0 {
				if managerAddr.Empty() {
					return fmt.Errorf("manager address must be specified to assign access roles")
				}
				accessGrants = append(accessGrants, *markertypes.NewAccessGrant(managerAddr, markertypes.AccessListByNames(acs)))
			}

			isFinalize, err := cmd.Flags().GetBool(flagFinalize)
			if err != nil {
				return err
			}

			isActivate, err := cmd.Flags().GetBool(flagActivate)
			if err != nil {
				return err
			}

			if isFinalize {
				if managerAddr.Empty() {
					return fmt.Errorf("a finalized marker account must have a manager set")
				}
				if isActivate {
					return fmt.Errorf("only one of --finalize, --activate maybe used")
				}
				markerStatus = markertypes.StatusFinalized
			}

			if isActivate {
				if isFinalize {
					return fmt.Errorf("only one of --finalize, --activate maybe used")
				}
				markerStatus = markertypes.StatusActive
			}

			newMarkerFlags, err := markercli.ParseNewMarkerFlags(cmd)
			if err != nil {
				return err
			}

			genAccount := markertypes.NewMarkerAccount(
				authtypes.NewBaseAccount(markertypes.MustGetMarkerAddress(markerCoin.Denom), nil, 0, 0),
				markerCoin,
				managerAddr,
				accessGrants,
				markerStatus,
				newMarkerFlags.MarkerType,
				newMarkerFlags.SupplyFixed,
				newMarkerFlags.AllowGovControl,
				newMarkerFlags.AllowForceTransfer,
				newMarkerFlags.RequiredAttributes,
			)

			if err = genAccount.Validate(); err != nil {
				return fmt.Errorf("failed to validate new genesis account: %w", err)
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			escrow, err := cmd.Flags().GetString(flagEscrow)
			if err != nil {
				return err
			}

			if len(escrow) > 0 {
				escrowCoins, parseCoinErr := sdk.ParseCoinsNormalized(escrow)
				if parseCoinErr != nil {
					return fmt.Errorf("failed to parse escrow coins: %w", parseCoinErr)
				}
				balances := banktypes.Balance{Address: genAccount.Address, Coins: escrowCoins}
				bankGenState := banktypes.GetGenesisStateFromAppState(cdc, appState)
				bankGenState.Balances = append(bankGenState.Balances, balances)
				bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

				bankGenStateBz, bankMarshalError := cdc.MarshalJSON(bankGenState)
				if bankMarshalError != nil {
					return fmt.Errorf("failed to marshal bank genesis state: %w", err)
				}
				appState[banktypes.ModuleName] = bankGenStateBz
			}

			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)
			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			if accs.Contains(genAccount.GetAddress()) {
				return fmt.Errorf("cannot add account at existing address %s", genAccount.Address)
			}

			// Add the new account to the set of genesis accounts and sanitize the
			// accounts afterwards.
			accs = append(accs, genAccount)
			accs = authtypes.SanitizeGenesisAccounts(accs)

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to convert accounts into any's: %w", err)
			}
			authGenState.Accounts = genAccs

			authGenStateBz, err := cdc.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}

			appState[authtypes.ModuleName] = authGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")

	markercli.AddNewMarkerFlags(cmd)
	cmd.Flags().String(flagManager, "", "a key name or address to assign as the token manager")
	cmd.Flags().String(flagAccess, "", "A comma separated list of access to grant to the manager account. [mint,burn,deposit,withdraw,delete,grant]")
	cmd.Flags().String(flagEscrow, "", "A list of coins held by the marker account instance.  Note: Any supply not allocated to other accounts should be assigned here.")
	cmd.Flags().BoolP(flagFinalize, "f", false, "Set the marker status to finalized.  Requires manager to be specified. (recommended)")
	cmd.Flags().BoolP(flagActivate, "a", false, "Set the marker status to active.  Total supply constraint will be enforced as invariant.")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// AddGenesisFlatFeeCmd returns add-msg-fee cobra command.
func AddGenesisFlatFeeCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-msg-fee <msg-url> <cost>",
		Aliases: []string{"add-genesis-msg-fee"},
		Short:   "Add a flat msg fee to genesis.json",
		Long: `Add a flat msg fee to genesis.json. This will create a msg based fee for an sdk msg type.
The command will validate that the msg-url is a valid sdk.msg and that the fee is a valid amount.`,
		Example: fmt.Sprintf(`$ %[1]s add-msg-fee /cosmos.bank.v1beta1.MsgSend 10000000000nhash`, genCmdStart),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			msgType := args[0]
			if msgType[0] != '/' {
				msgType = "/" + msgType
			}

			_, err := cdc.InterfaceRegistry().Resolve(msgType)
			if err != nil {
				return fmt.Errorf("invalid msg type %q: %w", msgType, err)
			}

			// TODO[fees]: Make sure ParseCoinsNormalized works with an empty string.
			cost, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid cost %q: %w", args[1], err)
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("could not read genesis state: %w", err)
			}

			flatFeesGenState, err := flatfeestypes.GetGenesisStateFromAppState(cdc, appState)
			if err != nil {
				return fmt.Errorf("could not get x/flatfees genesis state: %w", err)
			}

			found := false
			for _, msgFee := range flatFeesGenState.MsgFees {
				if strings.EqualFold(msgFee.MsgTypeUrl, msgType) {
					found = true
					msgFee.Cost = cost
					break
				}
			}

			if !found {
				flatFeesGenState.MsgFees = append(flatFeesGenState.MsgFees, flatfeestypes.NewMsgFee(msgType, cost...))
			}

			flatFeesGenStateBz, err := cdc.MarshalJSON(flatFeesGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal x/flatfees genesis state: %w", err)
			}

			appState[msgfeetypes.ModuleName] = flatFeesGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// AddGenesisDefaultMarketCmd returns add-default-market cobra command.
func AddGenesisDefaultMarketCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use: `add-default-market [--denom <denom>]

All base accounts already in the genesis file will be given all permissions on this market.
An error is returned if no such accounts are in the existing genesis file.

If --denom <denom> is not provided, the staking bond denom is used.
If no staking bond denom is defined either, then nhash is used.

This command is equivalent to the following command:
$ ` + genCmdStart + ` add-custom-market \
    --name 'Default <denom> Market' \
    --create-ask 100<denom> --create-bid 100<denom> \
    --seller-flat 500<denom> --buyer-flat 500<denom> \
    --seller-ratios 20<denom>:1<denom> --buyer-ratios 20<denom>:1<denom> \
    --accepting-orders --allow-user-settle \
	--access-grants <all permissions to all known base accounts>

`,
		Aliases:               []string{"add-genesis-default-market"},
		Short:                 "Add a default market to the genesis file",
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		RunE: updateGenesisFileRunE(func(clientCtx client.Context, cmd *cobra.Command, _ []string, appState map[string]json.RawMessage) error {
			// Printing usage with the errors from here won't actually help anyone.
			cmd.SilenceUsage = true

			// Identify the accounts that will get all the permissions.
			var authGen authtypes.GenesisState
			err := clientCtx.Codec.UnmarshalJSON(appState[authtypes.ModuleName], &authGen)
			if err != nil {
				return fmt.Errorf("could not extract auth genesis state: %w", err)
			}
			genAccts, err := authtypes.UnpackAccounts(authGen.Accounts)
			if err != nil {
				return fmt.Errorf("could not unpack genesis acounts: %w", err)
			}
			addrs := make([]string, 0, len(genAccts))
			for _, acct := range genAccts {
				// We specifically only want accounts that are base accounts (i.e. no markers or others).
				// This should also include the validator accounts (that have already been added).
				baseAcct, ok := acct.(*authtypes.BaseAccount)
				if ok {
					addrs = append(addrs, baseAcct.Address)
				}
			}
			if len(addrs) == 0 {
				return errors.New("genesis file must have one or more BaseAccount before a default market can be added")
			}

			// Identify the denom that we'll use for the fees.
			feeDenom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return fmt.Errorf("error reading --%s value: %w", flagDenom, err)
			}
			if len(feeDenom) == 0 {
				var stGen stakingtypes.GenesisState
				err = clientCtx.Codec.UnmarshalJSON(appState[stakingtypes.ModuleName], &stGen)
				if err == nil {
					feeDenom = stGen.Params.BondDenom
				}
			}
			if len(feeDenom) == 0 {
				feeDenom = "nhash"
			}
			if err = sdk.ValidateDenom(feeDenom); err != nil {
				return err
			}

			// Create the market and add it to the app state.
			market := makeDefaultMarket(feeDenom, addrs)
			return addMarketsToAppState(clientCtx, appState, market)
		}),
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flagDenom, "", "The fee denom for the market")
	return cmd
}

// makeDefaultMarket creates the default market that uses the provided fee denom
// and gives all permissions to each of the provided addrs.
func makeDefaultMarket(feeDenom string, addrs []string) exchange.Market {
	market := exchange.Market{
		MarketDetails:        exchange.MarketDetails{Name: "Default Market"},
		AcceptingOrders:      true,
		AllowUserSettlement:  true,
		AcceptingCommitments: true,
	}

	if len(feeDenom) > 0 {
		creationFee := sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 100))
		settlementFlat := sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 500))
		settlementRatio := []exchange.FeeRatio{{Price: sdk.NewInt64Coin(feeDenom, 20), Fee: sdk.NewInt64Coin(feeDenom, 1)}}
		market.MarketDetails.Name = fmt.Sprintf("Default %s Market", feeDenom)
		market.FeeCreateAskFlat = creationFee
		market.FeeCreateBidFlat = creationFee
		market.FeeSellerSettlementFlat = settlementFlat
		market.FeeSellerSettlementRatios = settlementRatio
		market.FeeBuyerSettlementFlat = settlementFlat
		market.FeeBuyerSettlementRatios = settlementRatio
		market.FeeCreateCommitmentFlat = creationFee
		market.CommitmentSettlementBips = 50
		market.IntermediaryDenom = feeDenom
	}

	for _, addr := range addrs {
		market.AccessGrants = append(market.AccessGrants,
			exchange.AccessGrant{Address: addr, Permissions: exchange.AllPermissions()})
	}
	return market
}

// AddGenesisCustomMarketCmd returns add-custom-market cobra command.
func AddGenesisCustomMarketCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-custom-market",
		Aliases: []string{"add-genesis-custom-market"},
		Short:   "Add a market to the genesis file",
		RunE: updateGenesisFileRunE(func(clientCtx client.Context, cmd *cobra.Command, args []string, appState map[string]json.RawMessage) error {
			msg, err := exchangecli.MakeMsgGovCreateMarket(clientCtx, cmd.Flags(), args)
			if err != nil {
				return err
			}

			// Now that we've read all the flags and stuff, no need to show usage with any errors anymore in here.
			cmd.SilenceUsage = true
			return addMarketsToAppState(clientCtx, appState, msg.Market)
		}),
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	exchangecli.SetupCmdTxGovCreateMarket(cmd)
	exchangecli.AddUseDetails(cmd, "If no <market id> is provided, the next available one will be used.")
	return cmd
}

// addMarketsToAppState adds the given markets to the app state in the exchange module.
// If a provided market's MarketId is 0, the next available one will be identified and used.
func addMarketsToAppState(clientCtx client.Context, appState map[string]json.RawMessage, markets ...exchange.Market) error {
	cdc := clientCtx.Codec
	var exGenState exchange.GenesisState
	if len(appState[exchange.ModuleName]) > 0 {
		if err := cdc.UnmarshalJSON(appState[exchange.ModuleName], &exGenState); err != nil {
			return fmt.Errorf("could not extract exchange genesis state: %w", err)
		}
	}

	for _, market := range markets {
		if err := market.Validate(); err != nil {
			return err
		}

		if market.MarketId == 0 {
			market.MarketId = getNextAvailableMarketID(exGenState)
			if exGenState.LastMarketId < market.MarketId {
				exGenState.LastMarketId = market.MarketId
			}
		}

		exGenState.Markets = append(exGenState.Markets, market)
	}

	exGenStateBz, err := cdc.MarshalJSON(&exGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal exchange genesis state: %w", err)
	}
	appState[exchange.ModuleName] = exGenStateBz
	return nil
}

// getNextAvailableMarketID returns the next available market id given all the markets in the provided genesis state.
func getNextAvailableMarketID(exGenState exchange.GenesisState) uint32 {
	if len(exGenState.Markets) == 0 {
		return 1
	}

	marketIDsMap := make(map[uint32]bool, len(exGenState.Markets))
	for _, market := range exGenState.Markets {
		marketIDsMap[market.MarketId] = true
	}

	rv := uint32(1)
	for marketIDsMap[rv] {
		rv++
	}
	return rv
}
