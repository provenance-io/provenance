package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/msgfees/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Flag names and values
const (
	FlagFeeRate = "fee-rate"
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
		GetCmdMsgBasedFeesProposal(),
	)

	return txCmd
}

func GetCmdMsgBasedFeesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [add|update|remove] [title] [description] [deposit]",
		Args:  cobra.ExactArgs(4),
		Short: "Submit a msg based fee proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a msg fees proposal along with an initial deposit.
For add, update, and removal of msg based fees amount and min fee and/or rate fee must be set.
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

			feeMsg, err := clientCtx.InterfaceRegistry.Resolve(msgType)
			if err != nil {
				return err
			}

			_, ok := feeMsg.(sdk.Msg)
			if !ok {
				return fmt.Errorf("message type is not a sdk message: %v", msgType)
			}

			anyMsg, err := codectypes.NewAnyWithValue(feeMsg)
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

			var proposal govtypes.Content

			switch args[0] {
			case "add":
				proposal = &types.AddMsgBasedFeeProposal{
					Title:         args[1],
					Description:   args[2],
					Msg:           anyMsg,
					AdditionalFee: addFee,
				}
			case "update":
				proposal = &types.UpdateMsgBasedFeeProposal{
					Title:         args[1],
					Description:   args[2],
					Msg:           anyMsg,
					AdditionalFee: addFee,
				}
			case "remove":
				proposal = &types.RemoveMsgBasedFeeProposal{
					Title:       args[1],
					Description: args[2],
					Msg:         anyMsg,
				}
			default:
				return fmt.Errorf("unknown proposal type %s", args[0])
			}

			deposit, err := sdk.ParseCoinsNormalized(args[3])
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
	cmd.Flags().String(FlagMsgType, "", "proto type url for msg type")
	cmd.Flags().String(FlagMinFee, "", "additional fee for msg based fee")
	return cmd
}
