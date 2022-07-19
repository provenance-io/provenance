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
		GetCmdRewardProgramAdd(), GetCmdClaimReward(),
	)

	return txCmd
}

func GetCmdRewardProgramAdd() *cobra.Command {
	actionsExampleJSON := "{\"qualifying_actions\":[{\"delegate\":{\"minimum_actions\":\"0\",\"maximum_actions\":\"0\",\"minimum_delegation_amount\":{\"denom\":\"nhash\",\"amount\":\"0\"},\"maximum_delegation_amount\":{\"denom\":\"nhash\",\"amount\":\"100\"},\"minimum_active_stake_percentile\":\"0.000000000000000000\",\"maximum_active_stake_percentile\":\"1.000000000000000000\"}}]}"
	cmd := &cobra.Command{
		Use:     "add-reward-program [title] [description]",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"arp"},
		Short:   "Add a reward program",
		Long:    strings.TrimSpace(`Add a reward program`),
		Example: fmt.Sprintf(`$ %[1]s tx reward add-reward-program \
		"Program Title" "A short description" \
		--total-reward-pool 580nhash \
		--max-reward-by-address 10nhash \
    	--start-time 2022-05-10\
		--claim-periods 52 \
		--max-rollover-periods 4 \
		--claim-period-days 7 \
		--expire-days 14 \ 
		--qualifying-actions %s
		`, version.AppName, actionsExampleJSON),
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
			startTime, err := convertDateToTime(startTimeStr)
			if err != nil {
				return err
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

// convertDateToTime takes in a string of format YYYY-MM-dd and returns a time.Time or parsing error
func convertDateToTime(dateStr string) (time.Time, error) {
	dateParts := strings.Split(dateStr, "-")
	if len(dateParts) != 3 {
		return time.Time{}, fmt.Errorf("error parsing start date must be of format YYYY-MM-dd: %v", dateStr)
	}
	year, err := strconv.Atoi(dateParts[0])
	if err != nil {
		return time.Time{}, err
	}
	month, err := strconv.Atoi(dateParts[1])
	if err != nil {
		return time.Time{}, err
	}
	day, err := strconv.Atoi(dateParts[2])
	if err != nil {
		return time.Time{}, err
	}
	startTime := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return startTime, nil
}

func GetCmdClaimReward() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claim-reward [reward-program-id]",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"cr", "claim"},
		Short:   "Claim reward for specified reward program",
		Long:    strings.TrimSpace(`Claim reward for a specified reward program.  This will transfer all unclaimed rewards for outstanding claim periods to signers address.`),
		Example: fmt.Sprintf(`$ %[1]s tx reward claim-reward 1 --from mykey`, version.AppName),
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

			msg := types.NewMsgClaimRewardRequest(
				uint64(programID),
				callerAddr.String(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
