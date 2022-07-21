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
	const pending = "pending"
	const active = "active"
	const completed = "completed"
	const outstanding = "outstanding"
	cmd := &cobra.Command{
		Use:     "reward-program {reward_program_id|\"all\"|\"pending\"|\"active\"\"completed\"|\"outstanding\"}",
		Aliases: []string{"rp", "rewardprogram"},
		Short:   "Query the current reward programs",
		Long: fmt.Sprintf(`%[1]s reward-program {reward_program_id} - gets the reward program for a given id.
%[1]s reward-program all - gets all the reward programs
%[1]s reward-program active - gets all active the reward programs`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s reward-program 1
%[1]s reward-program all
%[1]s reward-program pending
%[1]s reward-program active
%[1]s reward-program outstanding
%[1]s reward-program completed`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var request types.RewardProgramsRequest
			arg0 := strings.TrimSpace(args[0])
			switch arg0 {
			case all:
				request.QueryType = types.RewardProgramsRequest_ALL
			case pending:
				request.QueryType = types.RewardProgramsRequest_PENDING
			case active:
				request.QueryType = types.RewardProgramsRequest_ACTIVE
			case completed:
				request.QueryType = types.RewardProgramsRequest_FINISHED
			case outstanding:
				request.QueryType = types.RewardProgramsRequest_OUTSTANDING
			default:
				return outputRewardProgramByID(clientCtx, queryClient, arg0)
			}

			var response *types.RewardProgramsResponse
			if response, err = queryClient.RewardPrograms(
				context.Background(),
				&request,
			); err != nil {
				return fmt.Errorf("failed to query reward programs: %s", err.Error())
			}

			return clientCtx.PrintProto(response)
		},
	}

	return cmd
}

func outputRewardProgramByID(client client.Context, queryClient types.QueryClient, arg string) error {
	programID, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid argument arg : %s", arg)
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

	return client.PrintProto(response)
}

func GetClaimPeriodRewardDistributionCmd() *cobra.Command {
	const all = "all"

	cmd := &cobra.Command{
		Use:     "claim-period-reward-distribution {\"all\"} | {reward_program_id} {claim_period_id}",
		Aliases: []string{"cprd", "reward-distribution", "rd", "claim-periods"},
		Short:   "Query the current claim period reward distributions",
		Long: fmt.Sprintf(`%[1]s claim-period-reward-distribution {reward_program_id} {claim_period_id} - gets the reward program for the given reward_program_id and claim_period_id
%[1]s epoch-reward-distribution all - gets all the claim period reward distributions`, cmdStart),
		Args: cobra.RangeArgs(1, 2),
		Example: fmt.Sprintf(`%[1]s claim-period-reward-distribution 1 1
%[1]s epoch-reward-distribution all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputClaimPeriodRewardDistributionAll(cmd)
			}

			if len(args) != 2 {
				return fmt.Errorf("a reward_program_id and an claim_period_id are required")
			}
			arg1 := args[1]

			return outputClaimPeriodRewardDistributionByID(cmd, arg0, arg1)
		},
	}

	return cmd
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
func outputClaimPeriodRewardDistributionByID(cmd *cobra.Command, rewardID, claimPeriodID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	id, err := strconv.Atoi(rewardID)
	if err != nil {
		return err
	}
	claimID, err := strconv.Atoi(claimPeriodID)
	if err != nil {
		return err
	}
	var response *types.ClaimPeriodRewardDistributionByIDResponse
	if response, err = queryClient.ClaimPeriodRewardDistributionsByID(
		context.Background(),
		&types.ClaimPeriodRewardDistributionByIDRequest{
			RewardId:      uint64(id),
			ClaimPeriodId: uint64(claimID),
		},
	); err != nil {
		return fmt.Errorf("failed to query reward claim %d: %s", id, err.Error())
	}

	if response.GetClaimPeriodRewardDistribution() == nil {
		return fmt.Errorf("reward does not exist for reward-id: %s claim-id %s", rewardID, claimPeriodID)
	}

	return clientCtx.PrintProto(response)
}
