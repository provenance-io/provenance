package cli

import (
	"fmt"
	"strconv"
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
	FlagTotalRewardPool         = "total-reward-pool"
	FlagMaxRewardByAddress      = "max-reward-by-address"
	FlagStartTime               = "start-time"
	FlagClaimPeriods            = "claim-periods"
	FlagClaimPeriodDays         = "claim-period-days"
	FlagExpireDays              = "expire-days"
	FlagQualifyingActions       = "qualifying-actions"
	FlagMaxRolloverClaimPeriods = "max-rollover-periods"
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
		GetCmdEndRewardProgram(),
		GetCmdClaimReward(),
	)

	return txCmd
}

func GetCmdRewardProgramAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-reward-program [title] [description]",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"arp"},
		Short:   "Add a reward program",
		Long:    strings.TrimSpace(`Add a reward program`),
		Example: fmt.Sprintf(`$ %[1]s tx reward add-reward-program "Program Title" "A short description" \
	--total-reward-pool 580nhash \
	--max-reward-by-address 10nhash \
	--start-time '2050-01-15T00:00:00Z' \
	--claim-periods 52 \
	--max-rollover-periods 4 \
	--claim-period-days 7 \
	--expire-days 14 \ 
	--qualifying-actions '{"qualifying_actions":[{"delegate":{"minimum_actions":"0","maximum_actions":"1","minimum_delegation_amount":{"denom":"nhash","amount":"0"},"maximum_delegation_amount":{"denom":"nhash","amount":"100"},"minimum_active_stake_percentile":"0.000000000000000000","maximum_active_stake_percentile":"1.000000000000000000"}}]}'
		`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			coinStr, err := cmd.Flags().GetString(FlagTotalRewardPool)
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
				return fmt.Errorf("unable to parse time (%v) required format is RFC3339 (%v) , %w", startTimeStr, time.RFC3339, err)
			}
			rewardProgramDays, err := cmd.Flags().GetUint64(FlagClaimPeriods)
			if err != nil {
				return err
			}
			claimPeriodDays, err := cmd.Flags().GetUint64(FlagClaimPeriodDays)
			if err != nil {
				return err
			}
			expireDays, err := cmd.Flags().GetUint64(FlagExpireDays)
			if err != nil {
				return err
			}
			contents, err := cmd.Flags().GetString(FlagQualifyingActions)
			if err != nil {
				return err
			}
			maxRolloverPeriods, err := cmd.Flags().GetUint64(FlagMaxRolloverClaimPeriods)
			if err != nil {
				return err
			}
			var actions types.QualifyingActions
			err = clientCtx.Codec.UnmarshalJSON([]byte(contents), &actions)
			if err != nil {
				return err
			}
			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgCreateRewardProgramRequest(
				args[0],
				args[1],
				callerAddr.String(),
				coin,
				maxCoin,
				startTime,
				rewardProgramDays,
				claimPeriodDays,
				maxRolloverPeriods,
				expireDays,
				actions.QualifyingActions,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagTotalRewardPool, "", "coins for reward program")
	cmd.Flags().String(FlagMaxRewardByAddress, "", "max amount of coins a single address can claim in rewards")
	cmd.Flags().String(FlagStartTime, "", "time to start the rewards program, this must be a time in the future or within the first epoch of format YYYY-MM-DDTHH:MM:SSZ00:00 (2012-11-01T22:08:41+07:00)")
	cmd.Flags().Uint64(FlagClaimPeriods, 52, "number of claim periods the reward program runs")
	cmd.Flags().Uint64(FlagClaimPeriodDays, 7, "number of days for a claim period interval")
	cmd.Flags().Uint64(FlagExpireDays, 7, "number of days to expire program after it has ended")
	cmd.Flags().String(FlagQualifyingActions, "", "json representation of qualifying actions")
	cmd.Flags().Uint64(FlagMaxRolloverClaimPeriods, 0, "max number of rollover claim periods")
	return cmd
}

func GetCmdEndRewardProgram() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "end-reward-program [reward-program-id]",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"erp", "end", "er"},
		Short:   "End a reward program",
		Long:    strings.TrimSpace(`End a reward program.  This will trigger the reward program to end after the current claim period`),
		Example: fmt.Sprintf(`$ %[1]s tx reward end-reward-program 1`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			callerAddr := clientCtx.GetFromAddress()
			programID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid argument : %s", args[0])
			}

			msg := types.NewMsgEndRewardProgramRequest(
				uint64(programID),
				callerAddr.String(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetCmdClaimReward() *cobra.Command {
	const all = "all"
	cmd := &cobra.Command{
		Use:     "claim-reward [reward-program-id|all]",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"cr", "claim"},
		Short:   "Claim reward for specified reward program or all programs",
		Long:    strings.TrimSpace(`Claim reward for a specified reward program or all programs.  This will transfer all unclaimed rewards for outstanding claim periods to signers address.`),
		Example: fmt.Sprintf(`$ %[1]s tx reward claim-reward 1 --from mykey
$ %[1]s tx reward claim-reward all --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			arg0 := strings.TrimSpace(args[0])
			if arg0 != all {
				return claimRewardProgramByID(clientCtx, arg0, cmd)
			}

			callerAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgClaimAllRewardsRequest(
				callerAddr.String(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func claimRewardProgramByID(client client.Context, arg string, cmd *cobra.Command) error {
	programID, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid argument arg : %s", arg)
	}
	callerAddr := client.GetFromAddress()

	msg := types.NewMsgClaimRewardsRequest(
		uint64(programID),
		callerAddr.String(),
	)
	return tx.GenerateOrBroadcastTxCLI(client, cmd.Flags(), msg)
}
