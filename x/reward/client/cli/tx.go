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
	FlagSubPeriodType       = "sub-period-type"
	FlagSubPeriods          = "sub-periods"
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
    	--sub-period-type day \
    	--start-time 2022-05-10\
    	--sub-periods 10 \
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
			startTimeStr, err := cmd.Flags().GetString(FlagStartTime)
			if err != nil || len(startTimeStr) == 0 {
				return err
			}
			startTime, err := time.Parse(time.RFC3339, startTimeStr)
			if err != nil {
				return err
			}
			subPeriodTypeKey, err := cmd.Flags().GetString(FlagSubPeriodType)
			if err != nil {
				return err
			}
			subPeriodTypeSeconds := types.PeriodTypeToSeconds[subPeriodTypeKey]
			if subPeriodTypeSeconds == 0 {
				return fmt.Errorf("sub period type %s does not exist", subPeriodTypeKey)
			}

			subPeriods, err := cmd.Flags().GetUint64(FlagSubPeriods)
			if err != nil {
				return err
			}
			if subPeriods < 1 {
				return errors.New("number of sub periods must be larger than 0")
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
				startTime,
				subPeriodTypeKey,
				subPeriods,
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
	cmd.Flags().String(FlagStartTime, "", "time to start the rewards program, this must be a time in the future or within the first epoch of format YYYY-MM-DDTHH:MM:SSZ00:00 (2012-11-01T22:08:41+07:00)")
	cmd.Flags().String(FlagSubPeriodType, "", "sub period type (day, week, month)")
	cmd.Flags().Uint64(FlagSubPeriods, 0, "number of sub periods for the reward program")
	cmd.Flags().String(FlagEligibilityCriteria, "", "json of the eligibility criteria")
	return cmd
}
