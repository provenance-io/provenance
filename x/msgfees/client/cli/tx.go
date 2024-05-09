package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
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
		Use:     "proposal {add|update|remove}",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"p"},
		Short:   "Submit a msg based fee proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a msg fees proposal along with an initial deposit.
For add, update, and removal of msg fees amount and min fee and/or rate fee must be set.
`),
		Example: fmt.Sprintf(`$ %[1]s tx msgfees add --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --additional-fee=612nhash --recipient=pb... --bips=5000 --deposit 1000000000nhash
$ %[1]s tx msgfees update --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --additional-fee=612000nhash --recipient=pb... --bips=5000 --deposit 1000000000nhash
$ %[1]s tx msgfees remove --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --deposit 1000000000nhash
`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			authority := provcli.GetAuthority(flagSet)
			proposalType := args[0]

			msgType, err := cmd.Flags().GetString(FlagMsgType)
			if err != nil {
				return err
			}

			_, err = clientCtx.InterfaceRegistry.Resolve(msgType)
			if err != nil {
				return err
			}

			recipient, err := flagSet.GetString(FlagRecipient)
			if err != nil {
				return err
			}

			bips, err := flagSet.GetString(FlagBips)
			if err != nil {
				return err
			}

			var addFee sdk.Coin
			if proposalType != "remove" {
				additionalFee, errMinFee := flagSet.GetString(FlagMinFee)
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

			var msg sdk.Msg
			switch args[0] {
			case "add":
				msg = types.NewMsgAddMsgFeeProposalRequest(msgType, addFee, recipient, bips, authority)
			case "update":
				msg = types.NewMsgUpdateMsgFeeProposalRequest(msgType, addFee, recipient, bips, authority)
			case "remove":
				msg = types.NewMsgRemoveMsgFeeProposalRequest(msgType, authority)
			default:
				return fmt.Errorf("unknown proposal type %q", args[0])
			}
			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	cmd.Flags().String(FlagMsgType, "", "proto type url for msg type")
	cmd.Flags().String(FlagMinFee, "", "additional fee for msg based fee")
	cmd.Flags().String(FlagRecipient, "", "optional recipient address for receiving partial fee based on basis points")
	cmd.Flags().String(FlagBips, "", "basis fee points to distribute to recipient")
	return cmd
}

func GetUpdateNhashPerUsdMilProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nhash-per-usd-mil <nhash-per-usd-mil>",
		Aliases: []string{"npum", "n-p-u-m"},
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a nhash per usd mil update proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a nhash per usd mil update proposal along with an initial deposit.
The nhash per usd mil is the number of nhash that will be multiplied by the usd mil amount.  Example: $1.000 usd where 1 mil equals 2000nhash will equate to 1000 * 2000 = 2000000nhash
`),
		Example: fmt.Sprintf(`$ %[1]s tx msgfees nhash-per-usd-mil 1234 --deposit 1000000000nhash
$ %[1]s tx msgfees npum 1234 --deposit 1000000000nhash
`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			flagSet := cmd.Flags()
			authority := provcli.GetAuthority(flagSet)
			if err != nil {
				return err
			}
			nhash := args[0]
			rate, err := strconv.ParseUint(nhash, 0, 64)
			if err != nil {
				return fmt.Errorf("unable to parse nhash value: %s", nhash)
			}
			msg := types.NewMsgUpdateNhashPerUsdMilProposalRequest(rate, authority)
			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}
	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetUpdateConversionFeeDenomProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "conversion-fee-denom <conversion-fee-denom>",
		Aliases: []string{"cfd", "c-f-d"},
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a conversion fee denom update proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a conversion fee denom update proposal along with an initial deposit.
The custom fee denom is the denom that usd will be converted to for fees with usd as denom type.`),
		Example: fmt.Sprintf(`$ %[1]s tx msgfees conversion-fee-denom customcoin --deposit 1000000000nhash
$ %[1]s tx msgfees cfd customcoin --deposit 1000000000nhash
`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			flagSet := cmd.Flags()
			authority := provcli.GetAuthority(flagSet)

			if err != nil {
				return err
			}
			customCoin := args[0]
			msg := types.NewMsgUpdateConversionFeeDenomProposalRequest(customCoin, authority)
			if err != nil {
				return err
			}
			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	return cmd
}
