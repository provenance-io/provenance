package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"io/ioutil"

	"github.com/provenance-io/provenance/x/marker/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/spf13/cobra"
)

const (
	FlagType                   = "type"
	FlagSupplyFixed            = "supplyFixed"
	FlagAllowGovernanceControl = "allowGovernanceControl"
	FlagTransferLimit          = "transfer-limit"
	FlagExpiration             = "expiration"
	FlagPeriod                 = "period"
	FlagPeriodLimit            = "period-limit"
	FlagSpendLimit             = "spend-limit"
	FlagAllowedMsgs            = "allowed-messages"
)

// NewTxCmd returns the top-level command for marker CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the marker module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetCmdFinalize(),
		GetCmdActivate(),
		GetCmdCancel(),
		GetCmdDelete(),
		GetCmdMint(),
		GetCmdBurn(),
		GetCmdAddAccess(),
		GetCmdDeleteAccess(),
		GetCmdWithdrawCoins(),
		GetNewTransferCmd(),
		GetCmdAddMarker(),
		GetCmdMarkerProposal(),
		GetCmdGrantAuthorization(),
		GetCmdRevokeAuthorization(),
		GetCmdFeeGrant(),
	)
	return txCmd
}

// GetCmdMarkerProposal returns a cmd for creating/submitting marker governance proposals
func GetCmdMarkerProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [type] [proposal-file] [deposit]",
		Args:  cobra.ExactArgs(3),
		Short: "Submit a marker proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a marker proposal along with an initial deposit.
Proposal title, description, deposit, and marker proposal params must be set in a provided JSON file.

Where proposal.json contains:

{
  "title": "Test Proposal",
  "description": "My awesome proposal",
  "denom": "denomstring"
  // additional properties based on type here
}


Valid Proposal Types (and associated parameters):

- AddMarker
	"amount": 100,
	"manager": "pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk", 
	"status": "active", // [proposed, finalized, active]
	"marker_type": "COIN", // COIN, RESTRICTED
	"access_list": [ {"address":"pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk", "permissions": [1,2,3]} ], 
	"supply_fixed": true, 
	"allow_governance_control": true, 

- IncreaseSupply
	"amount": {"denom":"coin", "amount":"10"}

- DecreaseSupply
	"amount": {"denom":"coin", "amount":"10"}

- SetAdministrator
	"access": [{"address":"pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk", "permissions": [1,2,3]}]

- RemoveAdministrator
	"removed_address": ["pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk"]

- ChangeStatus
	"new_status": "MARKER_STATUS_ACTIVE" // [finalized, active, cancelled, destroyed]

- WithdrawEscrow
	"amount": "100coin"
	"target_address": "pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk"

- SetDenomMetadata
	"metadata": {
		"description": "description text",
		"base": "basedenom",
		"display": "displaydenom",
		"name": "Denom Name",
		"symbol": "DSYMB",
		"denom_units": [
			{"denom":"basedenom","exponent":0,"aliases":[]},
			{"denom":"otherdenomunit","exponent":9,"aliases":[]}
		]
	}
`,
		),
		Example: fmt.Sprintf(`$ %s tx marker proposal AddMarker "path/to/proposal.json" 1000%s --from mykey`, version.AppName, sdk.DefaultBondDenom),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var proposal govtypes.Content

			switch args[0] {
			case types.ProposalTypeAddMarker:
				proposal = &types.AddMarkerProposal{}
			case types.ProposalTypeIncreaseSupply:
				proposal = &types.SupplyIncreaseProposal{}
			case types.ProposalTypeDecreaseSupply:
				proposal = &types.SupplyDecreaseProposal{}
			case types.ProposalTypeSetAdministrator:
				proposal = &types.SetAdministratorProposal{}
			case types.ProposalTypeRemoveAdministrator:
				proposal = &types.RemoveAdministratorProposal{}
			case types.ProposalTypeChangeStatus:
				proposal = &types.ChangeStatusProposal{}
			case types.ProposalTypeWithdrawEscrow:
				proposal = &types.WithdrawEscrowProposal{}
			case types.ProposalTypeSetDenomMetadata:
				proposal = &types.SetDenomMetadataProposal{}
			default:
				return fmt.Errorf("unknown proposal type %s", args[0])
			}
			err = json.Unmarshal(contents, proposal)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(proposal, deposit, callerAddr)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %s", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddMarker implements the create marker command
func GetCmdAddMarker() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "new [coin]",
		Aliases: []string{"n"},
		Args:    cobra.ExactArgs(1),
		Short:   "Create a new marker",
		Long: strings.TrimSpace(`Creates a new marker in the Proposed state managed by the from address
with the given supply amount and denomination provided in the coin argument
`),
		Example: fmt.Sprintf(`$ %s tx marker new 1000hotdogcoin --%s=false --%s=false --from=mykey`, FlagType, FlagSupplyFixed, FlagAllowGovernanceControl),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			markerType := ""
			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid coin %s", args[0])
			}
			callerAddr := clientCtx.GetFromAddress()
			markerType, err = cmd.Flags().GetString(FlagType)
			if err != nil {
				return fmt.Errorf("invalid marker type: %w", err)
			}
			typeValue := types.MarkerType_Coin
			if len(markerType) > 0 {
				typeValue = types.MarkerType(types.MarkerType_value["MARKER_TYPE_"+markerType])
				if typeValue < 1 {
					return fmt.Errorf("invalid marker type: %s; expected COIN|RESTRICTED", markerType)
				}
			}
			supplyFixed, err := cmd.Flags().GetBool(FlagSupplyFixed)
			if err != nil {
				return fmt.Errorf("incorrect value for %s flag.  Accepted: true,false Error: %s", FlagSupplyFixed, err)
			}
			allowGovernanceControl, err := cmd.Flags().GetBool(FlagAllowGovernanceControl)
			if err != nil {
				return fmt.Errorf("incorrect value for %s flag.  Accepted: true,false Error: %s", FlagAllowGovernanceControl, err)
			}
			msg := types.NewMsgAddMarkerRequest(coin.Denom, coin.Amount, callerAddr, callerAddr, typeValue, supplyFixed, allowGovernanceControl)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(FlagType, "COIN", "a marker type to assign (default is COIN)")
	cmd.Flags().Bool(FlagSupplyFixed, false, "a true or false value to denote if a supply is fixed (default is false)")
	cmd.Flags().Bool(FlagAllowGovernanceControl, false, "a true or false value to denote if marker is allowed governance control (default is false)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdMint implements the mint additional supply for marker command.
func GetCmdMint() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mint [coin]",
		Aliases: []string{"m"},
		Args:    cobra.ExactArgs(1),
		Short:   "Mint coins against the marker",
		Long: strings.TrimSpace(`Mints coins of the marker's denomination and places them
in the marker's account under escrow.  Caller must possess the mint permission and 
marker must be in the active status.`),
		Example: fmt.Sprintf(`$ %s tx marker mint 1000hotdogcoin --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return sdkErrors.Wrapf(sdkErrors.ErrInvalidCoins, "invalid coin %s", args[0])
			}
			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgMintRequest(callerAddr, coin)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdBurn implements the burn coin supply from marker command.
func GetCmdBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "burn [coin]",
		Aliases: []string{"b"},
		Args:    cobra.ExactArgs(1),
		Short:   "Burn coins from the marker",
		Long: strings.TrimSpace(`Burns the number of coins specified from the marker associated
with the coin's denomination.  Only coins held in the marker's account may be burned.  Caller
must possess the burn permission.  Use the bank send operation to transfer coin into the marker
for burning.  Marker must be in the active status to burn coin.`),
		Example: fmt.Sprintf(`$ %s tx marker burn 1000hotdogcoin --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return sdkErrors.Wrapf(sdkErrors.ErrInvalidCoins, "invalid coin %s", args[0])
			}
			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgBurnRequest(callerAddr, coin)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdFinalize implements the finalize marker command.
func GetCmdFinalize() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "finalize [denom]",
		Aliases: []string{"f"},
		Args:    cobra.ExactArgs(1),
		Short:   "Finalize the marker account",
		Long: strings.TrimSpace(`Finalize a marker identified by the given denomination. Only
the marker manager may finalize a marker.  Once finalized callers who have been assigned
permission may perform mint,burn, or grant operations.  Only the manager may activate the marker.`),
		Example: fmt.Sprintf(`$ %s tx marker finalize hotdogcoin --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgFinalizeRequest(args[0], callerAddr)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdActivate implements the activate marker command.
func GetCmdActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activate [denom]",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		Short:   "Activate the marker account",
		Long: strings.TrimSpace(`Activate a marker identified by the given denomination. Only
the marker manager may activate a marker.  Once activated any total supply less than the
amount in circulation will be minted.  Invariant checks will be enforced.`),
		Example: fmt.Sprintf(`$ %s tx marker activate hotdogcoin --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgActivateRequest(args[0], callerAddr)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCancel implements the cancel marker command.
func GetCmdCancel() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel [denom]",
		Aliases: []string{"c"},
		Args:    cobra.ExactArgs(1),
		Short:   "Cancel the marker account",
		Example: fmt.Sprintf(`$ %s tx marker cancel hotdogcoin --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgCancelRequest(args[0], callerAddr)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDelete implements the destroy marker command.
func GetCmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destroy [denom]",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(1),
		Short:   "Mark the marker for deletion",
		Example: fmt.Sprintf(`$ %s tx marker destroy hotdogcoin --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgDeleteRequest(args[0], callerAddr)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddAccess implements the delegate access to a marker command.
func GetCmdAddAccess() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grant [address] [denom] [permission]",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(3),
		Short:   "Grant access to a marker for the address coins from the marker",
		Long: strings.TrimSpace(`Grant administrative access to a marker.  From Address must have appropriate
existing access.  Permissions are appended to any existing access grant.  Valid permissions
are one of [mint, burn, deposit, withdraw, delete, admin, transfer].`),
		Example: fmt.Sprintf(`$ %s tx marker grant pb1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj coindenom burn --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			targetAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return sdkErrors.Wrapf(err, "grant for invalid address %s", args[0])
			}
			grant := types.NewAccessGrant(targetAddr, types.AccessListByNames(args[2]))
			if err = grant.Validate(); err != nil {
				return sdkErrors.Wrapf(err, "invalid access grant permission: %s", args[2])
			}
			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgAddAccessRequest(args[1], callerAddr, *grant)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDeleteAccess implements the revoke administrative access for a marker command.
func GetCmdDeleteAccess() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke [address] [denom]",
		Aliases: []string{"r"},
		Args:    cobra.ExactArgs(2),
		Short:   "Revoke all access to a marker for the address",
		Long: strings.TrimSpace(`Revoke all administrative access to a marker for given access.
From Address must have appropriate existing access.`),
		Example: fmt.Sprintf(`$ %s tx marker revoke pb1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj coindenom --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			targetAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return sdkErrors.Wrapf(err, "revoke grant for invalid address %s", args[0])
			}
			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewDeleteAccessRequest(args[1], callerAddr, targetAddr)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdWithdrawCoins implements the withdraw coins from escrow command.
func GetCmdWithdrawCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw [marker-denom] [coins] [(optional) recipient address]",
		Aliases: []string{"w"},
		Args:    cobra.RangeArgs(2, 3),
		Short:   "Withdraw coins from the marker.",
		Long: "Withdraw coins from the marker escrow account.  Must be called by a user with the appropriate permissions. " +
			"If the recipient is not provided then the withdrawn amount is deposited in the caller's account.",
		Example: fmt.Sprintf(`$ %s tx marker withdraw coindenom 100coindenom pb1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			denom := args[0]
			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return sdkErrors.Wrapf(sdkErrors.ErrInvalidCoins, "invalid coin %s", args[0])
			}
			callerAddr := clientCtx.GetFromAddress()
			recipientAddr := sdk.AccAddress{}
			if len(args) == 3 {
				recipientAddr, err = sdk.AccAddressFromBech32(args[2])
				if err != nil {
					return sdkErrors.Wrapf(err, "invalid recipient address %s", args[2])
				}
			}
			msg := types.NewMsgWithdrawRequest(callerAddr, recipientAddr, denom, coins)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// Transfer handles a message to send coins from one account to another
func GetNewTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer [from] [to] [coins]",
		Aliases: []string{"t"},
		Short:   "Transfer coins from one account to another",
		Example: fmt.Sprintf(`$ %s tx marker transfer tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx tp1z6403t8z42fpl760zguuf2pc24g5gq96sez0k4 100coindenom --from mykey`, version.AppName),
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			from, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return sdkErrors.Wrapf(err, "invalid from address %s", args[0])
			}
			to, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return sdkErrors.Wrapf(err, "invalid recipient address %s", args[1])
			}
			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return sdkErrors.Wrapf(sdkErrors.ErrInvalidCoins, "invalid coin %s", args[2])
			}
			if len(coins) != 1 {
				return sdkErrors.Wrapf(sdkErrors.ErrInvalidCoins, "invalid coin %s", args[2])
			}
			msg := types.NewMsgTransferRequest(clientCtx.GetFromAddress(), from, to, coins[0])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetCmdGrantAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grant-authz [grantee] [authorization_type]",
		Aliases: []string{"ga"},
		Args:    cobra.ExactArgs(2),
		Short:   "Grant authorization to an address",
		Long:    strings.TrimSpace(`grant authorization to an address to execute an authorization type [transfer]`),
		Example: fmt.Sprintf(`$ %s tx marker grant-authz tp1skjw.. transfer --transfer-limit=1000nhash`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			exp, err := cmd.Flags().GetInt64(FlagExpiration)
			if err != nil {
				return err
			}

			var authorization authz.Authorization
			switch args[1] {
			case "transfer":
				limit, terr := cmd.Flags().GetString(FlagTransferLimit)
				if terr != nil {
					return terr
				}

				spendLimit, terr := sdk.ParseCoinsNormalized(limit)
				if terr != nil {
					return terr
				}

				if !spendLimit.IsAllPositive() {
					return fmt.Errorf("transfer-limit should be greater than zero")
				}

				authorization = types.NewMarkerTransferAuthorization(spendLimit)
			default:
				return fmt.Errorf("invalid authorization type, %s", args[1])
			}

			msg, err := authz.NewMsgGrant(clientCtx.GetFromAddress(), grantee, authorization, time.Unix(exp, 0))
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagTransferLimit, "", "The total amount an account is allowed to tranfer on granter's behalf")
	cmd.Flags().Int64(FlagExpiration, time.Now().AddDate(1, 0, 0).Unix(), "The Unix timestamp. Default is one year.")
	return cmd
}

func GetCmdRevokeAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke-authz [grantee] [authorization_type]",
		Short:   "Revoke authorization to an address",
		Aliases: []string{"ra"},
		Args:    cobra.ExactArgs(2),
		Long:    strings.TrimSpace(`revoke authorization to a grantee address for authorization type [transfer]`),
		Example: fmt.Sprintf(`$ %s tx marker revoke-authz tp1skjw.. transfer`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			var action string
			switch args[1] {
			case "transfer":
				action = types.MarkerTransferAuthorization{}.MsgTypeURL()
			default:
				return fmt.Errorf("invalid action type, %s", args[1])
			}

			msg := authz.NewMsgRevoke(clientCtx.GetFromAddress(), grantee, action)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdFeeGrant returns a CLI command handler for creating a MsgGrantAllowance transaction.
func GetCmdFeeGrant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feegrant [denom] [administrator_key_or_address] [grantee]",
		Short: "Grant Fee allowance to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Grant authorization to pay fees from your address. Note, the'--from' flag is
				ignored as it is implied from [administrator].

Examples:
%s tx %s feegrant markerdenom pb1edlyu... pb1psh7r... --spend-limit 100stake --expiration 2022-01-30T15:04:05Z or
%s tx %s feegrant markerdenom pb1edlyu... pb1psh7r... --spend-limit 100stake --period 3600 --period-limit 10stake --expiration 36000 or
%s tx %s feegrant markerdenom pb1edlyu... pb1psh7r... --spend-limit 100stake --expiration 2022-01-30T15:04:05Z 
	--allowed-messages "/cosmos.gov.v1beta1.MsgSubmitProposal,/cosmos.gov.v1beta1.MsgVote"
				`, version.AppName, types.ModuleName, version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			denom := args[0]
			err := cmd.Flags().Set(flags.FlagFrom, args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			administrator := clientCtx.GetFromAddress()
			sl, err := cmd.Flags().GetString(FlagSpendLimit)
			if err != nil {
				return err
			}

			// if `FlagSpendLimit` isn't set, limit will be nil
			limit, err := sdk.ParseCoinsNormalized(sl)
			if err != nil {
				return err
			}

			exp, err := cmd.Flags().GetString(FlagExpiration)
			if err != nil {
				return err
			}

			basic := feegrant.BasicAllowance{
				SpendLimit: limit,
			}

			var expiresAtTime time.Time
			if exp != "" {
				expiresAtTime, err = time.Parse(time.RFC3339, exp)
				if err != nil {
					return err
				}
				basic.Expiration = &expiresAtTime
			}

			var allowance feegrant.FeeAllowanceI
			allowance = &basic

			periodClock, err := cmd.Flags().GetInt64(FlagPeriod)
			if err != nil {
				return err
			}

			periodLimitVal, err := cmd.Flags().GetString(FlagPeriodLimit)
			if err != nil {
				return err
			}

			// Check any of period or periodLimit flags set, If set consider it as periodic fee allowance.
			if periodClock > 0 || periodLimitVal != "" {
				var periodLimit sdk.Coins
				periodLimit, err = sdk.ParseCoinsNormalized(periodLimitVal)
				if err != nil {
					return err
				}

				if periodClock <= 0 {
					return fmt.Errorf("period clock was not set")
				}

				if periodLimit == nil {
					return fmt.Errorf("period limit was not set")
				}

				periodReset := getPeriodReset(periodClock)
				if exp != "" && periodReset.Sub(expiresAtTime) > 0 {
					return fmt.Errorf("period (%d) cannot reset after expiration (%v)", periodClock, exp)
				}

				periodic := feegrant.PeriodicAllowance{
					Basic:            basic,
					Period:           getPeriod(periodClock),
					PeriodReset:      getPeriodReset(periodClock),
					PeriodSpendLimit: periodLimit,
					PeriodCanSpend:   periodLimit,
				}

				allowance = &periodic
			}

			allowedMsgs, err := cmd.Flags().GetStringSlice(FlagAllowedMsgs)
			if err != nil {
				return err
			}

			if len(allowedMsgs) > 0 {
				allowance, err = feegrant.NewAllowedMsgAllowance(allowance, allowedMsgs)
				if err != nil {
					return err
				}
			}

			msg, err := types.NewMsgGrantAllowance(denom, administrator, grantee, allowance)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().StringSlice(FlagAllowedMsgs, []string{}, "Set of allowed messages for fee allowance")
	cmd.Flags().String(FlagExpiration, "", "The RFC 3339 timestamp after which the grant expires for the user")
	cmd.Flags().String(FlagSpendLimit, "", "Spend limit specifies the max limit can be used, if not mentioned there is no limit")
	cmd.Flags().Int64(FlagPeriod, 0, "period specifies the time duration in which period_spend_limit coins can be spent before that allowance is reset")
	cmd.Flags().String(FlagPeriodLimit, "", "period limit specifies the maximum number of coins that can be spent in the period")

	return cmd
}

func getPeriodReset(duration int64) time.Time {
	return time.Now().Add(getPeriod(duration))
}

func getPeriod(duration int64) time.Duration {
	return time.Duration(duration) * time.Second
}
