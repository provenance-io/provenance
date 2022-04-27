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
		GetRewardClaimCmd(),
		GetEpochRewardDistributionCmd(),
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

			return outputRewardProgramById(cmd, arg0)
		},
	}

	return cmd
}

func GetRewardClaimCmd() *cobra.Command {
	const all = "all"

	cmd := &cobra.Command{
		Use:     "reward-claim {address|\"all\"}",
		Aliases: []string{"rc", "rewardclaim"},
		Short:   "Query the current reward claims",
		Long: fmt.Sprintf(`%[1]s reward-claim {address} - gets the reward claim for the address.
%[1]s reward-claim all - gets all the reward claims`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s reward-claim "pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk"
%[1]s reward-claim all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputRewardClaimsAll(cmd)
			}

			return outputRewardClaimById(cmd, arg0)
		},
	}

	return cmd
}

func GetEpochRewardDistributionCmd() *cobra.Command {
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
				return outputEpochRewardDistributionAll(cmd)
			}

			if len(args) != 2 {
				return fmt.Errorf("a reward_program_id and an epoch_id are required")
			}
			arg1 := args[1]

			return outputEpochRewardDistributionById(cmd, arg0, arg1)
		},
	}

	return cmd
}

func GetEligibilityCriteriaCmd() *cobra.Command {
	const all = "all"

	cmd := &cobra.Command{
		Use:     "eligibility-criteria {name|\"all\"}",
		Aliases: []string{"ec", "eligibilitycriteria"},
		Short:   "Query the current eligiblity criteria",
		Long: fmt.Sprintf(`%[1]s eligibility-criteria {name} - gets the eligibility criteria that matches the name.
%[1]s eligibility-criteria all - gets all the eligibility criteria`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s eligibility-criteria "delegate"
%[1]s eligibility-criteria all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputEligibilityCriteriaAll(cmd)
			}

			return outputEligibilityCriteriaByName(cmd, arg0)
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
func outputRewardProgramById(cmd *cobra.Command, arg string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	programId, err := strconv.Atoi(arg)
	if err != nil {
		return err
	}

	var response *types.RewardProgramByIDResponse
	if response, err = queryClient.RewardProgramByID(
		context.Background(),
		&types.RewardProgramByIDRequest{Id: uint64(programId)},
	); err != nil {
		return fmt.Errorf("failed to query reward program %d: %s", programId, err.Error())
	}

	if response.GetRewardProgram() == nil {
		return fmt.Errorf("reward program %d does not exist", programId)
	}

	return clientCtx.PrintProto(response)
}

func outputRewardClaimsAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)

	var response *types.RewardClaimsResponse
	if response, err = queryClient.RewardClaims(
		context.Background(),
		&types.RewardClaimsRequest{},
	); err != nil {
		return fmt.Errorf("failed to query reward programs: %s", err.Error())
	}

	return clientCtx.PrintProto(response)
}

func outputRewardClaimById(cmd *cobra.Command, arg string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)

	var response *types.RewardClaimByIDResponse
	if response, err = queryClient.RewardClaimByID(
		context.Background(),
		&types.RewardClaimByIDRequest{Id: arg},
	); err != nil {
		return fmt.Errorf("failed to query reward claim %s: %s", arg, err.Error())
	}

	if response.GetRewardClaim() == nil {
		return fmt.Errorf("reward claim %s does not exist", arg)
	}

	return clientCtx.PrintProto(response)
}

func outputEpochRewardDistributionAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)

	var response *types.EpochRewardDistributionResponse
	if response, err = queryClient.EpochRewardDistributions(
		context.Background(),
		&types.EpochRewardDistributionRequest{},
	); err != nil {
		return fmt.Errorf("failed to query reward programs: %s", err.Error())
	}

	return clientCtx.PrintProto(response)
}

func outputEpochRewardDistributionById(cmd *cobra.Command, rewardId, epochId string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	id, err := strconv.Atoi(rewardId)
	if err != nil {
		return err
	}

	var response *types.EpochRewardDistributionByIDResponse
	if response, err = queryClient.EpochRewardDistributionsByID(
		context.Background(),
		&types.EpochRewardDistributionByIDRequest{
			RewardId: uint64(id),
			EpochId:  epochId,
		},
	); err != nil {
		return fmt.Errorf("failed to query reward claim %d: %s", id, err.Error())
	}

	if response.GetEpochRewardDistribution() == nil {
		return fmt.Errorf("epoch reward does not exist for reward-id: %s epoch-id %s", rewardId, epochId)
	}

	return clientCtx.PrintProto(response)
}

func outputEligibilityCriteriaAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)

	var response *types.EligibilityCriteriaResponse
	if response, err = queryClient.EligibilityCriteria(
		context.Background(),
		&types.EligibilityCriteriaRequest{},
	); err != nil {
		return fmt.Errorf("failed to query eligibility criteria: %s", err.Error())
	}

	return clientCtx.PrintProto(response)
}

func outputEligibilityCriteriaByName(cmd *cobra.Command, name string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)

	var response *types.EligibilityCriteriaRequestByNameResponse
	if response, err = queryClient.EligibilityCriteriaByName(
		context.Background(),
		&types.EligibilityCriteriaRequestByNameRequest{
			Name: name,
		},
	); err != nil {
		return fmt.Errorf("failed to query eligibility criteria %s: %s", name, err.Error())
	}

	if response.GetEligibilityCriteria() == nil {
		return fmt.Errorf("eligibility criteria does not exist for name: %s", name)
	}

	return clientCtx.PrintProto(response)
}
