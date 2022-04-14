package cli

import (
	"context"
	"fmt"

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
		AllRewardProgramsCmd(),
	)
	return queryCmd
}

func AllRewardProgramsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List all the reward programs on the Provenance Blockchain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			// What is ReadPageRequest

			// We need our response
			var response *types.RewardProgramsResponse
			if response, err = queryClient.RewardPrograms(
				context.Background(),
				&types.RewardProgramsRequest{},
			); err != nil {
				fmt.Printf("failed to query reward programs: %s\n", err.Error())
				return nil
			}

			return clientCtx.PrintProto(response)
		},
	}

	//flags.AddPaginationFlagsToCmd(cmd, "markers")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
