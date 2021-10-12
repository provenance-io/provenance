package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/provenance-io/provenance/x/msgfees/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Flag names and values
const (
	FlagFeeRate = "fee-rate"
	FlagMinFee  = "min-fee"
	FlagMsgType = "msg-type"
	FlagAmount  = "amount"
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
		Use:   "proposal [type] [title] [description] [deposit]",
		Args:  cobra.ExactArgs(4),
		Short: "Submit a marker proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a msg fees proposal along with an initial deposit.
Proposal title, description, deposit, and marker proposal params must be set in a provided JSON file.

`,
		),
		Example: "TODO",
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
				return fmt.Errorf("Message type is not a sdk message: %v", msgType)
			}

			anyMsg, err := codectypes.NewAnyWithValue(feeMsg)
			if err != nil {
				return err
			}

			var amount sdk.Coin
			var minFee sdk.Coin
			var feeRate sdk.Dec
			if proposalType != "remove" {

				amountArg, err := cmd.Flags().GetString(FlagAmount)
				if err != nil || amountArg == "" {
					return errors.New("amount must be set")
				}
				amount, err = sdk.ParseCoinNormalized(amountArg)
				if err != nil {
					return err
				}

				minFeeArg, err := cmd.Flags().GetString(FlagMinFee)
				if err != nil {
					return err
				}
				if minFeeArg != "" {
					minFee, err = sdk.ParseCoinNormalized(minFeeArg)
					if err != nil {
						return err
					}
				}

				feeRateArg, err := cmd.Flags().GetString(FlagFeeRate)
				if err != nil {
					return err
				}
				if feeRateArg != "" {
					decInt, err := strconv.ParseInt(feeRateArg, 10, 64)
					if err != nil {
						return err
					}
					feeRate = sdk.NewDec(decInt)
					if err != nil {
						return err
					}
				} else {
					feeRate = sdk.ZeroDec()
				}
			}

			var proposal govtypes.Content

			switch args[0] {
			case "add":
				proposal = &types.AddMsgBasedFeeProposal{
					Title:       args[1],
					Description: args[2],
					Msg:         anyMsg,
					Amount:      amount,
					FeeRate:     feeRate,
					MinFee:      minFee,
				}
			case "update":
				proposal = &types.UpdateMsgBasedFeeProposal{
					Title:       args[1],
					Description: args[2],
					Msg:         anyMsg,
					Amount:      amount,
					FeeRate:     feeRate,
					MinFee:      minFee,
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
	cmd.Flags().String(FlagAmount, "", "amount for msg based fee")
	cmd.Flags().String(FlagMinFee, "", "min fee rate for msg based fee")
	cmd.Flags().String(FlagFeeRate, "", "fee rate for msg based fee")
	return cmd
}
