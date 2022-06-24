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
	FlagCoin               = "coin"
	FlagMaxRewardByAddress = "max-reward-by-address"
	FlagStartTime          = "start-time"
	FlagRewardPeriodDays   = "reward-period-days"
	FlagClaimPeriodDays    = "claim-period-days"
	FlagExpireDays         = "expire-days"
	FlagQualifyingActions  = "qualifying-actions"
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
    	--start-time 2022-05-10\
		--reward-period-days 365 \
		--claim-period-days 7 \
		--expire-days 14 \ 
		--qualifying-actions {}
The example command details state: TODO
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
			startTime, err := convertDateToTime(startTimeStr)
			if err != nil {
				return err
			}
			rewardProgramDays, err := cmd.Flags().GetUint64(FlagRewardPeriodDays)
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
			var actions types.QualifyingActions
			clientCtx.Codec.MustUnmarshalJSON([]byte(contents), &actions)

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
				expireDays,
				actions.QualifyingActions,
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
	cmd.Flags().String(FlagRewardPeriodDays, "", "number of days the reward program runs")
	cmd.Flags().String(FlagClaimPeriodDays, "", "number of days for a claim period interval")
	cmd.Flags().String(FlagExpireDays, "", "number of days to expire program after it has ended")
	cmd.Flags().String(FlagQualifyingActions, "", "json representation of qualifying actions")
	return cmd
}

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
