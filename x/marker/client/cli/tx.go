package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cerrs "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channelutils "github.com/cosmos/ibc-go/v6/modules/core/04-channel/client/utils"

	attrcli "github.com/provenance-io/provenance/x/attribute/client/cli"
	"github.com/provenance-io/provenance/x/marker/types"
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
	FlagAllowList              = "allow-list"
	FlagAllowedMsgs            = "allowed-messages"
	FlagPacketTimeoutHeight    = "packet-timeout-height"
	FlagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	FlagAbsoluteTimeouts       = "absolute-timeouts"
	FlagMemo                   = "memo"
	FlagRequiredAttributes     = "required-attributes"
	FlagAllowForceTransfer     = "allow-force-transfer"
	FlagAdd                    = "add"
	FlagRemove                 = "remove"
	FlagGovProposal            = "gov-proposal"
	FlagUsdCents               = "usd-cents"
	FlagVolume                 = "volume"
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
		GetIbcTransferTxCmd(),
		GetCmdAddFinalizeActivateMarker(),
		GetCmdUpdateRequiredAttributes(),
		GetCmdUpdateForcedTransfer(),
		GetCmdSetAccountData(),
		GetCmdUpdateSendDenyListRequest(),
		GetCmdAddNetAssetValues(),
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

			contents, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			var proposal govtypesv1beta1.Content

			switch args[0] {
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
			msg, err := govtypesv1beta1.NewMsgSubmitProposal(proposal, deposit, callerAddr)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %w", err)
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
		Example: fmt.Sprintf(`$ %s tx marker new 1000hotdogcoin --%s=false --%s=false --from=mykey --%s=attr.one,*.attr.two,...`,
			FlagType,
			FlagSupplyFixed,
			FlagAllowGovernanceControl,
			FlagRequiredAttributes,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid coin %s", args[0])
			}
			callerAddr := clientCtx.GetFromAddress()

			flagVals, err := ParseNewMarkerFlags(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddMarkerRequest(
				coin.Denom,
				coin.Amount,
				callerAddr,
				callerAddr,
				flagVals.MarkerType,
				flagVals.SupplyFixed,
				flagVals.AllowGovControl,
				flagVals.AllowForceTransfer,
				flagVals.RequiredAttributes,
				flagVals.UsdCents,
				flagVals.Volume,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	AddNewMarkerFlags(cmd)
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
				return sdkErrors.ErrInvalidCoins.Wrapf("invalid coin %s", args[0])
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
				return sdkErrors.ErrInvalidCoins.Wrapf("invalid coin %s", args[0])
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
				return cerrs.Wrapf(err, "grant for invalid address %s", args[0])
			}
			grant := types.NewAccessGrant(targetAddr, types.AccessListByNames(args[2]))
			if err = grant.Validate(); err != nil {
				return cerrs.Wrapf(err, "invalid access grant permission: %s", args[2])
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
				return cerrs.Wrapf(err, "revoke grant for invalid address %s", args[0])
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
				return sdkErrors.ErrInvalidCoins.Wrapf("invalid coin %s", args[0])
			}
			callerAddr := clientCtx.GetFromAddress()
			recipientAddr := sdk.AccAddress{}
			if len(args) == 3 {
				recipientAddr, err = sdk.AccAddressFromBech32(args[2])
				if err != nil {
					return cerrs.Wrapf(err, "invalid recipient address %s", args[2])
				}
			}
			msg := types.NewMsgWithdrawRequest(callerAddr, recipientAddr, denom, coins)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetNewTransferCmd implements the transfer command for marker funds.
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
				return cerrs.Wrapf(err, "invalid from address %s", args[0])
			}
			to, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return cerrs.Wrapf(err, "invalid recipient address %s", args[1])
			}
			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return sdkErrors.ErrInvalidCoins.Wrapf("invalid coin %s", args[2])
			}
			if len(coins) != 1 {
				return sdkErrors.ErrInvalidCoins.Wrapf("invalid coin %s", args[2])
			}
			msg := types.NewMsgTransferRequest(clientCtx.GetFromAddress(), from, to, coins[0])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetIbcTransferTxCmd returns the command to create a GetIbcTransferTxCmd transaction
func GetIbcTransferTxCmd() *cobra.Command {
	// TODO: refactor ibc-transfer usage comments to be provenance specific
	cmd := &cobra.Command{
		Use:   "ibc-transfer [src-port] [src-channel] [sender] [receiver] [amount]",
		Short: "Transfer a restricted marker token through IBC",
		Long: strings.TrimSpace(`Transfer a restricted marker through IBC. Timeouts can be specified
as absolute or relative using the "absolute-timeouts" flag. Timeout height can be set by passing in the height string
in the form {revision}-{height} using the "packet-timeout-height" flag. Relative timeout height is added to the block
height queried from the latest consensus state corresponding to the counterparty channel. Relative timeout timestamp 
is added to the greater value of the local clock time and the block timestamp queried from the latest consensus state 
corresponding to the counterparty channel. Any timeout set to 0 is disabled.`),
		Example: fmt.Sprintf("%s tx marker ibc-transfer [src-port] [src-channel] [receiver] [amount]", version.AppName),
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sourcePort := args[0]
			sourceChannel := args[1]
			sender := args[2]
			receiver := args[3]
			token, err := sdk.ParseCoinNormalized(args[4])
			if err != nil {
				return err
			}

			timeoutHeightStr, err := cmd.Flags().GetString(FlagPacketTimeoutHeight)
			if err != nil {
				return err
			}
			timeoutHeight, err := clienttypes.ParseHeight(timeoutHeightStr)
			if err != nil {
				return err
			}

			timeoutTimestamp, err := cmd.Flags().GetUint64(FlagPacketTimeoutTimestamp)
			if err != nil {
				return err
			}

			absoluteTimeouts, err := cmd.Flags().GetBool(FlagAbsoluteTimeouts)
			if err != nil {
				return err
			}

			memo, err := cmd.Flags().GetString(FlagMemo)
			if err != nil {
				return err
			}

			// if the timeouts are not absolute, retrieve latest block height and block timestamp
			// for the consensus state connected to the destination port/channel
			if !absoluteTimeouts {
				consensusState, height, _, err := channelutils.QueryLatestConsensusState(clientCtx, sourcePort, sourceChannel)
				if err != nil {
					return err
				}

				if !timeoutHeight.IsZero() {
					absoluteHeight := height
					absoluteHeight.RevisionNumber += timeoutHeight.RevisionNumber
					absoluteHeight.RevisionHeight += timeoutHeight.RevisionHeight
					timeoutHeight = absoluteHeight
				}

				if timeoutTimestamp != 0 {
					// use local clock time as reference time if it is later than the
					// consensus state timestamp of the counter party chain, otherwise
					// still use consensus state timestamp as reference
					now := time.Now().UnixNano()
					consensusStateTimestamp := consensusState.GetTimestamp()
					if now > 0 {
						now := uint64(now)
						if now > consensusStateTimestamp {
							timeoutTimestamp = now + timeoutTimestamp
						} else {
							timeoutTimestamp = consensusStateTimestamp + timeoutTimestamp
						}
					} else {
						return errors.New("local clock time is not greater than Jan 1st, 1970 12:00 AM")
					}
				}
			}
			msg := types.NewIbcMsgTransferRequest(
				clientCtx.GetFromAddress().String(),
				sourcePort, sourceChannel,
				token, sender, receiver,
				timeoutHeight, timeoutTimestamp,
				memo,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagPacketTimeoutHeight, "0-1000", "Packet timeout block height. The timeout is disabled when set to 0-0.")
	cmd.Flags().Uint64(FlagPacketTimeoutTimestamp, uint64((time.Duration(10) * time.Minute).Nanoseconds()), "Packet timeout timestamp in nanoseconds from now. Default is 10 minutes. The timeout is disabled when set to 0.")
	cmd.Flags().Bool(FlagAbsoluteTimeouts, false, "Timeout flags are used as absolute timeouts.")
	cmd.Flags().String(FlagMemo, "", "Memo to be sent along with the packet.")
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

			expSec, err := cmd.Flags().GetInt64(FlagExpiration)
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

				allowList, terr := cmd.Flags().GetStringSlice(FlagAllowList)
				if terr != nil {
					return terr
				}

				allowed, terr := bech32toAccAddresses(allowList)
				if terr != nil {
					return terr
				}

				authorization = types.NewMarkerTransferAuthorization(spendLimit, allowed)
			default:
				return fmt.Errorf("invalid authorization type, %s", args[1])
			}

			exp := time.Unix(expSec, 0)
			msg, err := authz.NewMsgGrant(clientCtx.GetFromAddress(), grantee, authorization, &exp)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagTransferLimit, "", "The total amount an account is allowed to transfer on granter's behalf")
	cmd.Flags().StringSlice(FlagAllowList, []string{}, "Allowed addresses grantee is allowed to send restricted coins separated by ,")
	cmd.Flags().Int64(FlagExpiration, time.Now().AddDate(1, 0, 0).Unix(), "The Unix timestamp. Default is one year.")
	return cmd
}

// bech32toAccAddresses returns []AccAddress from a list of Bech32 string addresses.
func bech32toAccAddresses(accAddrs []string) ([]sdk.AccAddress, error) {
	addrs := make([]sdk.AccAddress, len(accAddrs))
	for i, addr := range accAddrs {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, err
		}
		addrs[i] = accAddr
	}
	return addrs, nil
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

// GetCmdAddFinalizeActivateMarker implements the add finalize activate marker command
func GetCmdAddFinalizeActivateMarker() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-finalize-activate [coin] [access-grant-string]",
		Aliases: []string{"cfa"},
		Args:    cobra.ExactArgs(2),
		Short:   "Creates, Finalizes and Activates a new marker",
		Long: strings.TrimSpace(`Creates a new marker, finalizes it and put's it ACTIVATED state managed by the from address
with the given supply amount and denomination provided in the coin argument
`),
		Example: fmt.Sprintf(`$ %s tx marker create-finalize-activate 1000hotdogcoin address1,mint,admin;address2,burn --%s=false --%s=false --%s=attr.one,*.attr.two,... --from=mykey`,
			FlagType,
			FlagSupplyFixed,
			FlagAllowGovernanceControl,
			FlagRequiredAttributes,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid coin %s", args[0])
			}
			callerAddr := clientCtx.GetFromAddress()

			flagVals, err := ParseNewMarkerFlags(cmd)
			if err != nil {
				return err
			}

			accessGrants := ParseAccessGrantFromString(args[1])
			if len(accessGrants) == 0 {
				panic("at least one access grant should be present.")
			}

			msg := types.NewMsgAddFinalizeActivateMarkerRequest(
				coin.Denom, coin.Amount, callerAddr, callerAddr, flagVals.MarkerType,
				flagVals.SupplyFixed, flagVals.AllowGovControl,
				flagVals.AllowForceTransfer, flagVals.RequiredAttributes, accessGrants, flagVals.UsdCents, flagVals.Volume,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	AddNewMarkerFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUpdateRequiredAttributes implements the update required attributes command
func GetCmdUpdateRequiredAttributes() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-required-attributes <denom>",
		Aliases: []string{"ura"},
		Args:    cobra.ExactArgs(1),
		Short:   "Update required attributes on an existing restricted marker",
		Long: strings.TrimSpace(`Updates the required attributes of an existing restricted marker.
`),
		Example: fmt.Sprintf(`$ %s tx marker update-required-attributes hotdogcoin --%s=attr.one,*.attr.two,... --%s=attr.one,*.attr.two,...`,
			version.AppName,
			FlagAdd,
			FlagRemove,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			flagSet := cmd.Flags()

			msg := &types.MsgUpdateRequiredAttributesRequest{Denom: args[0]}

			msg.AddRequiredAttributes, err = flagSet.GetStringSlice(FlagAdd)
			if err != nil {
				return fmt.Errorf("incorrect value for %s flag.  Accepted: comma delimited list of attributes Error: %w", FlagAdd, err)
			}

			msg.RemoveRequiredAttributes, err = flagSet.GetStringSlice(FlagRemove)
			if err != nil {
				return fmt.Errorf("incorrect value for %s flag.  Accepted: comma delimited list of attributes Error: %w", FlagRemove, err)
			}

			authSetter := func(authority string) {
				msg.TransferAuthority = authority
			}

			return generateOrBroadcastOptGovProp(clientCtx, flagSet, authSetter, msg)
		},
	}
	cmd.Flags().StringSlice(FlagAdd, []string{}, "comma delimited list of required attributes to be added to restricted marker")
	cmd.Flags().StringSlice(FlagRemove, []string{}, "comma delimited list of required attributes to be removed from restricted marker")
	addOptGovPropFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUpdateForcedTransfer returns a CLI command for updating a marker's allow_force_transfer flag.
func GetCmdUpdateForcedTransfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-forced-transfer <denom> {true|false}",
		Aliases: []string{"uft"},
		Short:   "Submit a governance proposal to update the allow_forced_transfer field on a restricted marker",
		Example: fmt.Sprintf("$ %s tx marker update-forced-transfer hotdogcoin true", version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateForcedTransferRequest(args[0], false, authtypes.NewModuleAddress(govtypes.ModuleName))
			msg.AllowForcedTransfer, err = ParseBoolStrict(args[1])
			if err != nil {
				return err
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, cmd.Flags(), msg)
		},
	}

	govcli.AddGovPropFlagsToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUpdateSendDenyListRequest implements the update deny list command
func GetCmdUpdateSendDenyListRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-deny-list <denom>",
		Aliases: []string{"udl", "deny-list", "deny"},
		Args:    cobra.ExactArgs(1),
		Short:   "Update list of addresses for a restricted marker that are allowed to execute transfers",
		Long: strings.TrimSpace(`Update list of addresses for a restricted marker that are allowed to execute transfers.
`),
		Example: fmt.Sprintf(`$ %s tx marker update-deny-list hotdogcoin --%s=bech32addr1,bech32addrs2,... --%s=bech32addr1,bech32addrs2,...`,
			version.AppName,
			FlagAdd,
			FlagRemove,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			flagSet := cmd.Flags()

			msg := &types.MsgUpdateSendDenyListRequest{Denom: args[0]}

			msg.AddDeniedAddresses, err = flagSet.GetStringSlice(FlagAdd)
			if err != nil {
				return fmt.Errorf("incorrect value for %s flag.  Accepted: comma delimited list of bech32 addresses Error: %w", FlagAdd, err)
			}

			msg.RemoveDeniedAddresses, err = flagSet.GetStringSlice(FlagRemove)
			if err != nil {
				return fmt.Errorf("incorrect value for %s flag.  Accepted: comma delimited list of bech32 addresses Error: %w", FlagRemove, err)
			}

			authSetter := func(authority string) {
				msg.Authority = authority
			}

			return generateOrBroadcastOptGovProp(clientCtx, flagSet, authSetter, msg)
		},
	}
	cmd.Flags().StringSlice(FlagAdd, []string{}, "comma delimited list of bech32 addresses to be added to restricted marker transfer deny list")
	cmd.Flags().StringSlice(FlagRemove, []string{}, "comma delimited list of bech32 addresses to be removed removed from restricted marker deny list")
	addOptGovPropFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSetAccountData returns a CLI command for setting a marker's account data.
func GetCmdSetAccountData() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-data <denom> " + attrcli.AccountDataFlagsUse,
		Aliases: []string{"accountdata", "ad"},
		Short:   "Set a marker's account data",
		Example: fmt.Sprintf(`$ %[1]s tx marker account-data hotdogcoin --%[2]s "This is hotdogcoin's marker data.'"
$ %[1]s tx marker account-data hotdogcoin --%[3]s hotdogcoin-account-data.json
$ %[1]s tx marker account-data hotdogcoin --%[4]s`,
			version.AppName, attrcli.FlagValue, attrcli.FlagFile, attrcli.FlagDelete),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			flagSet := cmd.Flags()

			msg := &types.MsgSetAccountDataRequest{Denom: strings.TrimSpace(args[0])}

			msg.Value, err = attrcli.ReadAccountDataFlags(flagSet)
			if err != nil {
				return err
			}

			setSigner := func(signer string) {
				msg.Signer = signer
			}

			return generateOrBroadcastOptGovProp(clientCtx, flagSet, setSigner, msg)
		},
	}

	attrcli.AddAccountDataFlagsToCmd(cmd)
	addOptGovPropFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddNetAssetValues returns a CLI command for adding/updating marker net asset values.
func GetCmdAddNetAssetValues() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-net-asset-values <denom> " + attrcli.AccountDataFlagsUse,
		Aliases: []string{"add-navs", "anavs"},
		Short:   "Add/updates net asset values for a marker",
		Example: fmt.Sprintf(`$ %[1]s tx marker add-net-asset-values hotdogcoin 1usd,1;2nhash,3`,
			version.AppName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			netAssetValues := ParseNetAssertValueString(args[1])
			denom := strings.TrimSpace(args[0])

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), types.NewMsgAddNetAssetValuesRequest(denom, clientCtx.From, netAssetValues))
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getPeriodReset(duration int64) time.Time {
	return time.Now().Add(getPeriod(duration))
}

func getPeriod(duration int64) time.Duration {
	return time.Duration(duration) * time.Second
}

// ParseAccessGrantFromString splits string (example address1,perm1,perm2...;address2, perm1...) to AccessGrant
func ParseAccessGrantFromString(addressPermissionString string) []types.AccessGrant {
	parts := strings.Split(addressPermissionString, ";")
	grants := make([]types.AccessGrant, 0)
	for _, p := range parts {
		// ignore if someone put in a ; at the end by mistake
		if len(p) == 0 {
			continue
		}
		partsPerAddress := strings.Split(p, ",")
		// if it has an address has to have at least one access associated with it
		if !(len(partsPerAddress) > 1) {
			panic("at least one grant should be provided with address")
		}
		var permissions types.AccessList
		address := sdk.MustAccAddressFromBech32(partsPerAddress[0])
		for _, permission := range partsPerAddress[1:] {
			permissions = append(permissions, types.AccessByName(permission))
		}
		grants = append(grants, *types.NewAccessGrant(address, permissions))
	}
	return grants
}

// ParseNetAssertValueString splits string (example address1,perm1,perm2...;address2, perm1...) to list of NetAssetValue's
func ParseNetAssertValueString(netAssetValuesString string) []types.NetAssetValue {
	navs := strings.Split(netAssetValuesString, ";")
	netAssetValues := make([]types.NetAssetValue, len(navs))
	for i, nav := range navs {
		parts := strings.Split(nav, ",")
		if len(parts) != 2 {
			panic("invalid net asset value, expected coin,volume")
		}
		coin, err := sdk.ParseCoinNormalized(parts[0])
		if err != nil {
			panic(fmt.Sprintf("invalid coin %s", parts[0]))
		}
		volume, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid volume %s", parts[1]))
		}
		netAssetValues[i] = types.NewNetAssetValue(coin, volume)
	}
	return netAssetValues
}

// AddNewMarkerFlags adds the flags needed when defining a new marker.
// The provided values can be retrieved using ParseNewMarkerFlags.
func AddNewMarkerFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagType, "COIN", "a marker type to assign (default is COIN)")
	cmd.Flags().Bool(FlagSupplyFixed, false, "Indicates that the supply is fixed")
	cmd.Flags().Bool(FlagAllowGovernanceControl, false, "Indicates that governance control is allowed")
	cmd.Flags().Bool(FlagAllowForceTransfer, false, "Indicates that force transfer is allowed")
	cmd.Flags().StringSlice(FlagRequiredAttributes, []string{}, "comma delimited list of required attributes needed for a restricted marker to have send authority")
	cmd.Flags().Uint64(FlagUsdCents, 0, "Indicates the net asset value of marker in usd cents, i.e. 1234 = $1.234")
	cmd.Flags().Uint64(FlagVolume, 0, "Indicates the volume of the net asset value")
}

// NewMarkerFlagValues represents the values provided in the flags added by AddNewMarkerFlags.
type NewMarkerFlagValues struct {
	MarkerType         types.MarkerType
	SupplyFixed        bool
	AllowGovControl    bool
	AllowForceTransfer bool
	RequiredAttributes []string
	UsdCents           uint64
	Volume             uint64
}

// ParseNewMarkerFlags reads the flags added by AddNewMarkerFlags.
func ParseNewMarkerFlags(cmd *cobra.Command) (*NewMarkerFlagValues, error) {
	rv := &NewMarkerFlagValues{}

	markerType, err := cmd.Flags().GetString(FlagType)
	if err != nil {
		return nil, fmt.Errorf("invalid marker type: %w", err)
	}
	if len(markerType) > 0 {
		rv.MarkerType = types.MarkerType(types.MarkerType_value["MARKER_TYPE_"+strings.ToUpper(markerType)])
		if rv.MarkerType < 1 {
			return nil, fmt.Errorf("invalid marker type: %s; expected COIN|RESTRICTED", markerType)
		}
	} else {
		rv.MarkerType = types.MarkerType_Coin
	}

	rv.SupplyFixed, err = cmd.Flags().GetBool(FlagSupplyFixed)
	if err != nil {
		return nil, fmt.Errorf("incorrect value for %s flag.  Accepted: true,false Error: %w", FlagSupplyFixed, err)
	}

	rv.AllowGovControl, err = cmd.Flags().GetBool(FlagAllowGovernanceControl)
	if err != nil {
		return nil, fmt.Errorf("incorrect value for %s flag.  Accepted: true,false Error: %w", FlagAllowGovernanceControl, err)
	}

	rv.AllowForceTransfer, err = cmd.Flags().GetBool(FlagAllowForceTransfer)
	if err != nil {
		return nil, fmt.Errorf("incorrect value for %s flag.  Accepted: true,false Error: %w", FlagAllowForceTransfer, err)
	}

	rv.RequiredAttributes, err = cmd.Flags().GetStringSlice(FlagRequiredAttributes)
	if err != nil {
		return nil, fmt.Errorf("incorrect value for %s flag.  Accepted: comma delimited list of attributes Error: %w", FlagRequiredAttributes, err)
	}

	rv.UsdCents, err = cmd.Flags().GetUint64(FlagUsdCents)
	if err != nil {
		return nil, fmt.Errorf("incorrect value for %s flag.  Accepted: 0 or greater value Error: %w", FlagUsdCents, err)
	}

	rv.Volume, err = cmd.Flags().GetUint64(FlagVolume)
	if err != nil {
		return nil, fmt.Errorf("incorrect value for %s flag.  Accepted: 0 or greater value Error: %w", FlagVolume, err)
	}

	if rv.UsdCents > 0 && rv.Volume == 0 {
		return nil, fmt.Errorf("incorrect value for %s flag.  Must be positive number if %s flag has been set to positive value", FlagVolume, FlagUsdCents)
	}

	return rv, nil
}

// ParseBoolStrict converts the provided input into a boolean.
// Valid strings are "true" and "false"; case is ignored.
func ParseBoolStrict(input string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean string: %q", input)
}

// addOptGovPropFlags adds the gov prop flags and a flag making them optional.
// See also: generateOrBroadcastOptGovProp
func addOptGovPropFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagGovProposal, false, "submit message as a gov proposal")
	govcli.AddGovPropFlagsToCmd(cmd)
}

// generateOrBroadcastOptGovProp either calls GenerateOrBroadcastTxCLIAsGovProp or GenerateOrBroadcastTxCLI
// depending on the presence of --gov-proposal in the flags.
// The authSetter is used to set the authority/signer of the provided message; if doing a gov prop,
// it's set to the gov module's account address, otherwise it's the --from address.
//
// See also: addOptGovPropFlags.
func generateOrBroadcastOptGovProp(clientCtx client.Context, flagSet *pflag.FlagSet, authSetter func(authority string), msg sdk.Msg) error {
	isGov, err := flagSet.GetBool(FlagGovProposal)
	if err != nil {
		return err
	}
	if isGov {
		authSetter(authtypes.NewModuleAddress(govtypes.ModuleName).String())
	} else {
		authSetter(clientCtx.GetFromAddress().String())
	}

	err = msg.ValidateBasic()
	if err != nil {
		return err
	}

	if isGov {
		return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
	}
	return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
}
