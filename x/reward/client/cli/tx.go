package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/reward/types"
)

// Flag names and values
const (
	FlagCoin                = "coin"
	FlagMaxRewardByAddress  = "max-reward-by-address"
	FlagStartTime           = "start-time"
	FlagEpochType           = "epoch-type"
	FlagNumEpochs           = "num-epochs"
	FlagEligibilityCriteria = "eligibility-criteria"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"r"},
		Short:                      "Transaction commands for the reward module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdRewardProgramAdd(),
	)

	return txCmd
}

func GetCmdRewardProgramAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-reward-program [title] [description]",
		Args:    cobra.ExactArgs(3),
		Aliases: []string{"arp"},
		Short:   "Add a reward program",
		Long:    strings.TrimSpace(`Add a reward program`),
		Example: fmt.Sprintf(`$ %[1]s tx reward new 
		--coin 580nhash \
		--max-reward-by-address 10nhash \
    	--epoch-type day \
    	--start-time 2022-05-10\
    	--num-epochs 10 \
    	--eligibility-criteria "{\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}"
		`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			coinStr, err := cmd.Flags().GetString(FlagCoin)
			if err != nil {
				return err
			}
			coin, err := sdk.ParseCoinNormalized(coinStr)
			if err != nil {
				return err
			}
			maxCoinStr, err := cmd.Flags().GetString(FlagMaxRewardByAddress)
			if err != nil {
				return err
			}
			maxCoin, err := sdk.ParseCoinNormalized(maxCoinStr)
			if err != nil {
				return err
			}
			// TODO: figure out the best format to parse with
			// startTimeStr, err := cmd.Flags().GetString(FlagStartTime)
			// if err != nil {
			// 	return err
			// }

			epochTypeKey, err := cmd.Flags().GetString(FlagEpochType)
			if err != nil {
				return err
			}
			epochTypeSeconds := types.EpochTypeToSeconds[epochTypeKey]
			if epochTypeSeconds == 0 {
				return fmt.Errorf("epoch type %s does not exist.", epochTypeKey)
			}

			numEpochs, err := cmd.Flags().GetUint64(FlagNumEpochs)
			if err != nil {
				return err
			}
			if numEpochs < 1 {
				return errors.New("number of epochs must be larger than 0")
			}
			eligibilityCriteriaStr, err := cmd.Flags().GetString(FlagEligibilityCriteria)
			if err != nil {
				return err
			}

			var eligibilityCriteria types.EligibilityCriteria
			err = clientCtx.Codec.UnmarshalJSON([]byte(eligibilityCriteriaStr), &eligibilityCriteria)
			if err != nil {
				return fmt.Errorf("unable to parse eligibility criteria : %s", err)
			}

			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgCreateRewardProgramRequest(
				args[0],
				args[1],
				callerAddr.String(),
				coin,
				maxCoin,
				time.Now(),
				epochTypeKey,
				numEpochs,
				eligibilityCriteria,
			)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %q", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagCoin, "", "coins for reward program")
	cmd.Flags().String(FlagMaxRewardByAddress, "", "max amount of coins a single address can claim in rewards")
	cmd.Flags().String(FlagStartTime, "", "time to start the rewards program, this must be a time in the future or within the first epoch")
	cmd.Flags().String(FlagEpochType, "", "epoch type (day, week, month)")
	cmd.Flags().Uint64(FlagNumEpochs, 0, "number of epochs for the reward program")
	cmd.Flags().String(FlagEligibilityCriteria, "", "json of the eligibility criteria")
	return cmd
}
