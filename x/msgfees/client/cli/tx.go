package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// Flag names and values
const (
	FlagMinFee    = "additional-fee"
	FlagMsgType   = "msg-type"
	FlagRecipient = "recipient"
	FlagBips      = "bips"
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
		GetUpdateConversionFeeDenomProposal(),
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
		Example: fmt.Sprintf(`$ %[1]s tx msgfees add "adding" "adding MsgWriterRecordRequest fee"  10nhash --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --additional-fee=612nhash --recipient=pb... --bips=5000
$ %[1]s tx msgfees update "updating" "updating MsgWriterRecordRequest fee"  10nhash --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --additional-fee=612000nhash --recipient=pb... --bips=5000
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

			recipient, err := cmd.Flags().GetString(FlagRecipient)
			if err != nil {
				return err
			}

			bips, err := cmd.Flags().GetString(FlagBips)
			if err != nil {
				return err
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

			var proposal govtypesv1beta1.Content

			switch args[0] {
			case "add":
				proposal = &types.AddMsgFeeProposal{
					Title:                args[1],
					Description:          args[2],
					MsgTypeUrl:           msgType,
					AdditionalFee:        addFee,
					Recipient:            recipient,
					RecipientBasisPoints: bips,
				}
			case "update":
				proposal = &types.UpdateMsgFeeProposal{
					Title:                args[1],
					Description:          args[2],
					MsgTypeUrl:           msgType,
					AdditionalFee:        addFee,
					Recipient:            recipient,
					RecipientBasisPoints: bips,
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
			msg, err := govtypesv1beta1.NewMsgSubmitProposal(proposal, deposit, callerAddr)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %w", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagMsgType, "", "proto type url for msg type")
	cmd.Flags().String(FlagMinFee, "", "additional fee for msg based fee")
	cmd.Flags().String(FlagRecipient, "", "optional recipient address for receiving partial fee based on basis points")
	cmd.Flags().String(FlagBips, "", "basis fee points to distribute to recipient")
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
$ %[1]s tx msgfees npum "updating nhash to usd mil" "changes the nhash per mil to 1234nhash" 1234 1000000000nhash
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

func GetUpdateConversionFeeDenomProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "conversion-fee-denom <title> <description> <conversion-fee-denom> <deposit>",
		Aliases: []string{"cfd", "c-f-d"},
		Args:    cobra.ExactArgs(4),
		Short:   "Submit a conversion fee denom update proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a conversion fee denom update proposal along with an initial deposit.
The custom fee denom is the denom that usd will be converted to for fees with usd as denom type.`),
		Example: fmt.Sprintf(`$ %[1]s tx msgfees conversion-fee-denom "updating conversion fee denom" "changes the conversion fee denom to customcoin"  customcoin 1000000000nhash
$ %[1]s tx msgfees cfd "updating conversion fee denom" "changes the conversion fee denom to customcoin"  customcoin 1000000000nhash
`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			title, description, customCoin, depositArg := args[0], args[1], args[2], args[3]
			proposal := types.NewUpdateConversionFeeDenomProposal(title, description, customCoin)
			deposit, err := sdk.ParseCoinsNormalized(depositArg)
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
