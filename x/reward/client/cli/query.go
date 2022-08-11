package cli

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

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
		GetRewardsByAddressCmd(),
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

			var request types.QueryRewardProgramsRequest
			arg0 := strings.TrimSpace(args[0])
			switch arg0 {
			case all:
				request.QueryType = types.QueryRewardProgramsRequest_QUERY_TYPE_ALL
			case pending:
				request.QueryType = types.QueryRewardProgramsRequest_QUERY_TYPE_PENDING
			case active:
				request.QueryType = types.QueryRewardProgramsRequest_QUERY_TYPE_ACTIVE
			case completed:
				request.QueryType = types.QueryRewardProgramsRequest_QUERY_TYPE_FINISHED
			case outstanding:
				request.QueryType = types.QueryRewardProgramsRequest_QUERY_TYPE_OUTSTANDING
			default:
				return outputRewardProgramByID(clientCtx, queryClient, arg0)
			}

			var response *types.QueryRewardProgramsResponse
			if response, err = queryClient.RewardPrograms(
				context.Background(),
				&request,
			); err != nil {
				return fmt.Errorf("failed to query reward programs: %w", err)
			}

			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func outputRewardProgramByID(client client.Context, queryClient types.QueryClient, arg string) error {
	programID, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid argument arg : %s", arg)
	}

	var response *types.QueryRewardProgramByIDResponse
	if response, err = queryClient.RewardProgramByID(
		context.Background(),
		&types.QueryRewardProgramByIDRequest{Id: uint64(programID)},
	); err != nil {
		return fmt.Errorf("failed to query reward program %d: %w", programID, err)
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
%[1]s reward-distribution all - gets all the claim period reward distributions`, cmdStart),
		Args: cobra.RangeArgs(1, 2),
		Example: fmt.Sprintf(`%[1]s claim-period-reward-distribution 1 1
%[1]s reward-distribution all`, cmdStart),
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
	flags.AddPaginationFlagsToCmd(cmd, "all")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetRewardsByAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reward-by-address {address} {\"all\"|\"unclaimable\"|\"claimable\"|\"claimed\"|\"expired\"}",
		Aliases: []string{"rpa", "reward-per-address"},
		Short:   "Query all the reward distributions for an address",
		Long:    fmt.Sprintf(`%[1]s reward-by-address {address} {query-type} - gets the reward amount for the given address based on the filter values`, cmdStart),
		Args:    cobra.ExactArgs(2),
		Example: fmt.Sprintf(`%[1]s reward-by-address pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk all
%[1]s reward-by-address pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk unclaimable
%[1]s reward-by-address pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk claimable
%[1]s reward-by-address pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk claimed
%[1]s reward-by-address pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk expired`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			arg1 := args[1]

			return queryRewardDistributionByAddress(cmd, arg0, arg1)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// Query for all ClaimPeriodRewardDistributions
func outputClaimPeriodRewardDistributionAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)

	var response *types.QueryClaimPeriodRewardDistributionsResponse
	if response, err = queryClient.ClaimPeriodRewardDistributions(
		context.Background(),
		&types.QueryClaimPeriodRewardDistributionsRequest{Pagination: pageReq},
	); err != nil {
		return fmt.Errorf("failed to query reward programs: %w", err)
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
	var response *types.QueryClaimPeriodRewardDistributionsByIDResponse
	if response, err = queryClient.ClaimPeriodRewardDistributionsByID(
		context.Background(),
		&types.QueryClaimPeriodRewardDistributionsByIDRequest{
			RewardId:      uint64(id),
			ClaimPeriodId: uint64(claimID),
		},
	); err != nil {
		return fmt.Errorf("failed to query reward claim %d: %w", id, err)
	}

	if response.GetClaimPeriodRewardDistribution() == nil {
		return fmt.Errorf("reward does not exist for reward-id: %s claim-id %s", rewardID, claimPeriodID)
	}

	return clientCtx.PrintProto(response)
}

// sdk ReadPageRequest expects binary, but we encoded to base64 in our marshaller
func withPageKeyDecoded(flagSet *flag.FlagSet) *flag.FlagSet {
	encoded, err := flagSet.GetString(flags.FlagPageKey)
	if err != nil {
		panic(err.Error())
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err.Error())
	}
	_ = flagSet.Set(flags.FlagPageKey, string(raw))
	return flagSet
}

// Query for RewardAccountByAddress depending on claim status
func queryRewardDistributionByAddress(cmd *cobra.Command, address string, queryType string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	var claimStatus types.RewardAccountState_ClaimStatus
	switch queryType {
	case "all":
		claimStatus = types.RewardAccountState_CLAIM_STATUS_UNSPECIFIED
	case "unclaimable":
		claimStatus = types.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE
	case "claimable":
		claimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMABLE
	case "claimed":
		claimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMED
	case "expired":
		claimStatus = types.RewardAccountState_CLAIM_STATUS_EXPIRED
	default:
		return fmt.Errorf("failed to query reward distributions. %s is not a valid query param", queryType)
	}

	queryClient := types.NewQueryClient(clientCtx)

	var response *types.QueryRewardDistributionsByAddressResponse
	if response, err = queryClient.RewardDistributionsByAddress(
		context.Background(),
		&types.QueryRewardDistributionsByAddressRequest{
			Address:     address,
			ClaimStatus: claimStatus,
		},
	); err != nil {
		return fmt.Errorf("failed to query reward distributions: %w", err)
	}

	return clientCtx.PrintProto(response)
}
