package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/reward/types"
)

func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the rewards module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		QueryRewardProgramsCmd(),
		RewardProgramByIdCmd(),
	)
	return queryCmd
}

func QueryRewardProgramsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List all the reward programs on the Provenance Blockchain. Optionally the '-a' flag can be set to list only active reward programs.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			// We need our response

			isActive, err := cmd.Flags().GetBool("active")
			if err != nil {
				return err
			}

			if isActive {
				var response *types.ActiveRewardProgramsResponse
				if response, err = queryClient.ActiveRewardPrograms(
					context.Background(),
					&types.ActiveRewardProgramsRequest{},
				); err != nil {
					fmt.Printf("failed to query active reward programs: %s\n", err.Error())
					return nil
				}
				return clientCtx.PrintProto(response)
			} else {
				var response *types.RewardProgramsResponse
				if response, err = queryClient.RewardPrograms(
					context.Background(),
					&types.RewardProgramsRequest{},
				); err != nil {
					fmt.Printf("failed to query reward programs: %s\n", err.Error())
					return nil
				}
				return clientCtx.PrintProto(response)
			}
		},
	}

	//flags.AddPaginationFlagsToCmd(cmd, "markers")
	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Bool("active", false, "Queries for only the active reward programs.")

	return cmd
}

func RewardProgramByIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "id",
		Aliases: []string{},
		Short:   "Output the reward program for the specified id",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			programId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			// We need our response
			var response *types.RewardProgramByIDResponse
			if response, err = queryClient.RewardProgramByID(
				context.Background(),
				&types.RewardProgramByIDRequest{Id: uint64(programId)},
			); err != nil {
				fmt.Printf("failed to query reward programs: %s\n", err.Error())
				return nil
			}

			if response.GetRewardProgram() == nil {
				fmt.Printf("reward program %d does not exist.\n", programId)
				return nil
			}

			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
