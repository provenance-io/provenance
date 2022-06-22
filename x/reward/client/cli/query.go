package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/reward/types"
)

var cmdStart = fmt.Sprintf("%s query reward", version.AppName)

func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the rewards module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetRewardProgramCmd(),
		GetClaimPeriodRewardDistributionCmd(),
	)
	return queryCmd
}

func GetRewardProgramCmd() *cobra.Command {
	const all = "all"
	const active = "active"

	cmd := &cobra.Command{
		Use:     "reward-program {reward_program_id|\"all\"|\"active\"}",
		Aliases: []string{"rp", "rewardprogram"},
		Short:   "Query the current reward programs",
		Long: fmt.Sprintf(`%[1]s reward-program {reward_program_id} - gets the reward program for a given id.
%[1]s reward-program all - gets all the reward programs
%[1]s reward-program active - gets all active the reward programs`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s reward-program 1
%[1]s reward-program all
%[1]s reward-program active`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputRewardProgramsAll(cmd)
			} else if arg0 == active {
				return outputRewardProgramsActive(cmd)
			}

			return outputRewardProgramByID(cmd, arg0)
		},
	}

	return cmd
}

func GetClaimPeriodRewardDistributionCmd() *cobra.Command {
	const all = "all"

	cmd := &cobra.Command{
		Use:     "epoch-reward-distribution {\"all\"} | {reward_program_id} {epoch_id}",
		Aliases: []string{"erd", "reward-distribution", "rd"},
		Short:   "Query the current epoch reward distributions",
		Long: fmt.Sprintf(`%[1]s epoch-reward-distribution {reward_program_id} {epoch_id} - gets the reward program for the given reward_program_id and epoch id
%[1]s epoch-reward-distribution all - gets all the reward programs`, cmdStart),
		Args: cobra.RangeArgs(1, 2),
		Example: fmt.Sprintf(`%[1]s epoch-reward-distribution 1 "day"
%[1]s epoch-reward-distribution all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputClaimPeriodRewardDistributionAll(cmd)
			}

			if len(args) != 2 {
				return fmt.Errorf("a reward_program_id and an epoch_id are required")
			}
			arg1 := args[1]

			return outputClaimPeriodRewardDistributionByID(cmd, arg0, arg1)
		},
	}

	return cmd
}

// Query for all Reward Programs
func outputRewardProgramsAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)

	var response *types.RewardProgramsResponse
	if response, err = queryClient.RewardPrograms(
		context.Background(),
		&types.RewardProgramsRequest{},
	); err != nil {
		return fmt.Errorf("failed to query reward programs: %s", err.Error())
	}

	return clientCtx.PrintProto(response)
}

// Query for all active Reward Programs
func outputRewardProgramsActive(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)

	var response *types.ActiveRewardProgramsResponse
	if response, err = queryClient.ActiveRewardPrograms(
		context.Background(),
		&types.ActiveRewardProgramsRequest{},
	); err != nil {
		return fmt.Errorf("failed to query active reward programs: %s", err.Error())
	}
	return clientCtx.PrintProto(response)
}

// Query for a RewardProgram by Id
func outputRewardProgramByID(cmd *cobra.Command, arg string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	programID, err := strconv.Atoi(arg)
	if err != nil {
		return err
	}

	var response *types.RewardProgramByIDResponse
	if response, err = queryClient.RewardProgramByID(
		context.Background(),
		&types.RewardProgramByIDRequest{Id: uint64(programID)},
	); err != nil {
		return fmt.Errorf("failed to query reward program %d: %s", programID, err.Error())
	}

	if response.GetRewardProgram() == nil {
		return fmt.Errorf("reward program %d does not exist", programID)
	}

	return clientCtx.PrintProto(response)
}

// Query for all ClaimPeriodRewardDistributions
func outputClaimPeriodRewardDistributionAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)

	var response *types.ClaimPeriodRewardDistributionResponse
	if response, err = queryClient.ClaimPeriodRewardDistributions(
		context.Background(),
		&types.ClaimPeriodRewardDistributionRequest{},
	); err != nil {
		return fmt.Errorf("failed to query reward programs: %s", err.Error())
	}

	return clientCtx.PrintProto(response)
}

// Query for a ClaimPeriodRewardDistribution by Id
func outputClaimPeriodRewardDistributionByID(cmd *cobra.Command, rewardID, claimPeriodId string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	id, err := strconv.Atoi(rewardID)
	if err != nil {
		return err
	}
	claimId, err := strconv.Atoi(claimPeriodId)
	if err != nil {
		return err
	}
	var response *types.ClaimPeriodRewardDistributionByIDResponse
	if response, err = queryClient.ClaimPeriodRewardDistributionsByID(
		context.Background(),
		&types.ClaimPeriodRewardDistributionByIDRequest{
			RewardId:      uint64(id),
			ClaimPeriodId: uint64(claimId),
		},
	); err != nil {
		return fmt.Errorf("failed to query reward claim %d: %s", id, err.Error())
	}

	if response.GetClaimPeriodRewardDistribution() == nil {
		return fmt.Errorf("reward does not exist for reward-id: %s claim-id %s", rewardID, claimPeriodId)
	}

	return clientCtx.PrintProto(response)
}
