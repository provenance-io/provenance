package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/version"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

const (
	flagVestingStart = "vesting-start-time"
	flagVestingEnd   = "vesting-end-time"
	flagVestingAmt   = "vesting-amount"

	flagRestricted = "restrict"

	flagManager  = "manager"
	flagAccess   = "access"
	flagEscrow   = "escrow"
	flagActivate = "activate"
	flagFinalize = "finalize"
	flagType     = "type"
)

// AddGenesisAccountCmd returns add-genesis-account cobra Command.
func AddGenesisAccountCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add a genesis account to genesis.json",
		Long: `Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
`,
		Example: fmt.Sprintf(`$ %[1]s add-genesis-account mykey 10000000hash`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONCodec
			cdc := depCdc.(codec.Codec)

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
				kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf)
				if err != nil {
					return err
				}

				info, err := kb.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keybase: %w, could not use address as bech32 string: %s", err, parseErr.Error())
				}

				addr = info.GetAddress()
			}

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse coins: %w", err)
			}

			vestingStart, err := cmd.Flags().GetInt64(flagVestingStart)
			if err != nil {
				return err
			}
			vestingEnd, err := cmd.Flags().GetInt64(flagVestingEnd)
			if err != nil {
				return err
			}
			vestingAmtStr, err := cmd.Flags().GetString(flagVestingAmt)
			if err != nil {
				return err
			}

			vestingAmt, err := sdk.ParseCoinsNormalized(vestingAmtStr)
			if err != nil {
				return fmt.Errorf("failed to parse vesting amount: %w", err)
			}

			// create concrete account type based on input parameters
			var genAccount authtypes.GenesisAccount

			balances := banktypes.Balance{Address: addr.String(), Coins: coins.Sort()}
			baseAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)

			if !vestingAmt.IsZero() {
				baseVestingAccount := authvesting.NewBaseVestingAccount(baseAccount, vestingAmt.Sort(), vestingEnd)

				if (balances.Coins.IsZero() && !baseVestingAccount.OriginalVesting.IsZero()) ||
					baseVestingAccount.OriginalVesting.IsAnyGT(balances.Coins) {
					return errors.New("vesting amount cannot be greater than total amount")
				}

				switch {
				case vestingStart != 0 && vestingEnd != 0:
					genAccount = authvesting.NewContinuousVestingAccountRaw(baseVestingAccount, vestingStart)

				case vestingEnd != 0:
					genAccount = authvesting.NewDelayedVestingAccountRaw(baseVestingAccount)

				default:
					return errors.New("invalid vesting parameters; must supply start and end time or end time")
				}
			} else {
				genAccount = baseAccount
			}

			if err = genAccount.Validate(); err != nil {
				return fmt.Errorf("failed to validate new genesis account: %w", err)
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			if accs.Contains(addr) {
				return fmt.Errorf("cannot add account at existing address %s", addr)
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

			bankGenState := banktypes.GetGenesisStateFromAppState(depCdc, appState)
			bankGenState.Balances = append(bankGenState.Balances, balances)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

			bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}

			appState[banktypes.ModuleName] = bankGenStateBz

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
	cmd.Flags().String(flagVestingAmt, "", "amount of coins for vesting accounts")
	cmd.Flags().Int64(flagVestingStart, 0, "schedule start time (unix epoch) for vesting accounts")
	cmd.Flags().Int64(flagVestingEnd, 0, "schedule end time (unix epoch) for vesting accounts")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// AddRootDomainAccountCmd returns add-genesis-root-name cobra command.
func AddRootDomainAccountCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-root-name [address_or_key_name] [root-name]",
		Short: "Add a name binding to genesis.json",
		Long: `Add a name binding to genesis.json. The provided account must specify
	the account address or key name and domain-name to bind. If a key name is given,
	the address will be looked up in the local Keybase.  The restricted flag can optionally be
	included to lock or unlock an entry to child names.
	`,
		Example: fmt.Sprintf(`$ %[1]s add-genesis-root-name mykey rootname`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONCodec
			cdc := depCdc.(codec.Codec)

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
				kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf)
				if err != nil {
					return err
				}

				info, err := kb.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keybase: %w, could use use address as bech32 string: %s", err, parseErr.Error())
				}

				addr = info.GetAddress()
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

// AddGenesisMarkerCmd configures a marker account and adds it to the list of genesis accounts
func AddGenesisMarkerCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-marker [coin] --manager [address_or_key_name] --access [grant][,[grant]] --escrow [coin][, [coin]] --finalize --activate --type [COIN]",
		Short: "Adds a marker to genesis.json",
		Long: `Adds a marker to genesis.json. The provided parameters must specify
the marker supply and denom as a coin.  A managing account may be added as a key name or address. An accessgrant
may be assigned to the manager address. The escrowed list of initial tokens must contain valid denominations. If
a marker account is activated any unassigned marker supply must be provided as escrow. Finalized markers must have
a manager address assigned that can activate the marker after genesis.  Activated markers will have supply invariants
enforced immediately.  An optional type flag can be provided or the default of COIN will be used.
`,
		Example: fmt.Sprintf(`$ %[1]s add-genesis-marker 1000000000funcoins --manager validator --access withdraw --escrow 100funcoins --finalize --type COIN`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONCodec
			cdc := depCdc.(codec.Codec)

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
			markerFlagType := ""

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
					kb, newKeyringError := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf)
					if newKeyringError != nil {
						return newKeyringError
					}
					info, keyErr := kb.Key(mgr)
					if keyErr != nil {
						return fmt.Errorf("failed to get address from Keybase: %w", keyErr)
					}

					managerAddr = info.GetAddress()
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

			markerFlagType, err = cmd.Flags().GetString(flagType)
			if err != nil {
				return err
			}

			if len(markerFlagType) == 0 {
				markerFlagType = "COIN"
			}

			markerType := markertypes.MarkerType_value["MARKER_TYPE_"+strings.ToUpper(markerFlagType)]
			if markerType == int32(markertypes.MarkerType_Unknown) {
				panic(fmt.Sprintf("unknown marker type %s", markerFlagType))
			}

			genAccount := markertypes.NewMarkerAccount(
				authtypes.NewBaseAccount(markertypes.MustGetMarkerAddress(markerCoin.Denom), nil, 0, 0),
				markerCoin,
				managerAddr,
				accessGrants,
				markerStatus,
				markertypes.MarkerType(markerType))

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
				bankGenState := banktypes.GetGenesisStateFromAppState(depCdc, appState)
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

	cmd.Flags().String(flagType, "", "a marker type to assign (default is COIN)")
	cmd.Flags().String(flagManager, "", "a key name or address to assign as the token manager")
	cmd.Flags().String(flagAccess, "", "A comma separated list of access to grant to the manager account. [mint,burn,deposit,withdraw,delete,grant]")
	cmd.Flags().String(flagEscrow, "", "A list of coins held by the marker account instance.  Note: Any supply not allocated to other accounts should be assigned here.")
	cmd.Flags().BoolP(flagFinalize, "f", false, "Set the marker status to finalized.  Requires manager to be specified. (recommended)")
	cmd.Flags().BoolP(flagActivate, "a", false, "Set the marker status to active.  Total supply constraint will be enforced as invariant.")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// AddGenesisMsgFeeCmd returns add-genesis-msg-fee cobra command.
func AddGenesisMsgFeeCmd(defaultNodeHome string, interfaceRegistry types.InterfaceRegistry) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-msg-fee [msg-url] [additional-fee]",
		Short: "Add a msg fee to genesis.json",
		Long: `Add a msg fee to to genesis.json. This will create a msg based fee for an sdk msg type.  The command will validate
		that the msg-url is a valid sdk.msg and that the fee is a valid amount and coin.
	`,
		Example: fmt.Sprintf(`$ %[1]s add-genesis-msg-fee /cosmos.bank.v1beta1.MsgSend 10000000000nhash`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONCodec
			cdc := depCdc.(codec.Codec)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			msgType := args[0]

			if msgType[0] != '/' {
				msgType = "/" + msgType
			}

			if err := checkMsgTypeValid(interfaceRegistry, msgType); err != nil {
				return err
			}

			additionalFee, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse coin: %w", err)
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			msgFeesGenState := msgfeetypes.GetGenesisStateFromAppState(cdc, appState)

			found := false
			for _, mf := range msgFeesGenState.MsgFees {
				if strings.EqualFold(mf.MsgTypeUrl, msgType) {
					found = true
					mf.AdditionalFee = additionalFee
				}
			}

			if !found {
				msgFeesGenState.MsgFees = append(msgFeesGenState.MsgFees, msgfeetypes.NewMsgFee(msgType, additionalFee))
			}

			msgFeesGenStateBz, err := cdc.MarshalJSON(&msgFeesGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal msgfees genesis state: %w", err)
			}

			appState[msgfeetypes.ModuleName] = msgFeesGenStateBz

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

func checkMsgTypeValid(registry types.InterfaceRegistry, msgTypeURL string) error {
	msg, err := registry.Resolve(msgTypeURL)
	if err != nil {
		return err
	}

	_, ok := msg.(sdk.Msg)
	if !ok {
		return fmt.Errorf("message type is not a sdk message: %v", msgTypeURL)
	}
	return err
}
