package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/msgfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Flag names and values
const (
	FlagMinFee  = "additional-fee"
	FlagMsgType = "msg-type"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"mf", "mfees", "mbf"},
		Short:                      "Transaction commands for the msgfees module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdMsgFeesProposal(),
		GetUpdateNhashPerUsdMilProposal(),
	)

	return txCmd
}

func GetCmdMsgFeesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "proposal {add|update|remove} <title> <description> <deposit>",
		Args:    cobra.ExactArgs(4),
		Aliases: []string{"p"},
		Short:   "Submit a msg based fee proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a msg fees proposal along with an initial deposit.
For add, update, and removal of msg fees amount and min fee and/or rate fee must be set.
`),
		Example: fmt.Sprintf(`$ %[1]s tx msgfees add "adding" "adding MsgWriterRecordRequest fee"  10nhash --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --additional-fee=612nhash
$ %[1]s tx msgfees update "updating" "updating MsgWriterRecordRequest fee"  10nhash --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --additional-fee=612000nhash
$ %[1]s tx msgfees remove "removing" "removing MsgWriterRecordRequest fee" 10nhash --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest
`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalType := args[0]

			msgType, err := cmd.Flags().GetString(FlagMsgType)
			if err != nil {
				return err
			}

			msgFee, err := clientCtx.InterfaceRegistry.Resolve(msgType)
			if err != nil {
				return err
			}

			_, ok := msgFee.(sdk.Msg)
			if !ok {
				return fmt.Errorf("message type is not a sdk message: %q", msgType)
			}

			var addFee sdk.Coin
			if proposalType != "remove" {
				additionalFee, errMinFee := cmd.Flags().GetString(FlagMinFee)
				if errMinFee != nil {
					return err
				}
				if additionalFee != "" {
					addFee, err = sdk.ParseCoinNormalized(additionalFee)
					if err != nil {
						return err
					}
				}
			}

			var proposal govtypes.Content

			switch args[0] {
			case "add":
				proposal = &types.AddMsgFeeProposal{
					Title:         args[1],
					Description:   args[2],
					MsgTypeUrl:    msgType,
					AdditionalFee: addFee,
				}
			case "update":
				proposal = &types.UpdateMsgFeeProposal{
					Title:         args[1],
					Description:   args[2],
					MsgTypeUrl:    msgType,
					AdditionalFee: addFee,
				}
			case "remove":
				proposal = &types.RemoveMsgFeeProposal{
					Title:       args[1],
					Description: args[2],
					MsgTypeUrl:  msgType,
				}
			default:
				return fmt.Errorf("unknown proposal type %q", args[0])
			}

			deposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(proposal, deposit, callerAddr)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %q", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagMsgType, "", "proto type url for msg type")
	cmd.Flags().String(FlagMinFee, "", "additional fee for msg based fee")
	return cmd
}

func GetUpdateNhashPerUsdMilProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nhash-per-usd-mil <title> <description> <nhash-per-usd-mil> <deposit>",
		Aliases: []string{"npum", "n-p-u-m"},
		Args:    cobra.ExactArgs(4),
		Short:   "Submit a nhash per usd mil update proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a nhash per usd mil update proposal along with an initial deposit.
The nhash per usd mil is the number of nhash that will be multiplied by the usd mil amount.  Example: $1.000 usd where 1 mil equals 2000nhash will equate to 1000 * 2000 = 2000000nhash
`),
		Example: fmt.Sprintf(`$ %[1]s tx msgfees nhash-per-usd-mil "updating nhash to usd mil" "changes the nhash per mil to 1234nhash"  1234 1000000000nhash
$ %[1]s tx msgfees npum nhash-per-usd-mil "updating nhash to usd mil" "changes the nhash per mil to 1234nhash"   1234 1000000000nhash
`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			title, description, nhash, depositArg := args[0], args[1], args[2], args[3]
			rate, err := strconv.ParseUint(nhash, 0, 64)
			if err != nil {
				return fmt.Errorf("unable to parse nhash value: %s", nhash)
			}
			proposal := types.NewUpdateNhashPerUsdMilProposal(title, description, rate)
			deposit, err := sdk.ParseCoinsNormalized(depositArg)
			if err != nil {
				return err
			}
			callerAddr := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(proposal, deposit, callerAddr)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %q", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
