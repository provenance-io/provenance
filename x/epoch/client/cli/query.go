package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/epoch/types"
)

// GetQueryCmd returns the top-level command for epoch CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the epoch module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		EpochInfosCmd(),
		CurrentEpochCmd(),
	)
	return queryCmd
}

// EpochInfosCmd is the CLI command for listing all epoch info.
func EpochInfosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List all the epoch info on the Provenance Blockchain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var response *types.QueryEpochInfosResponse
			if response, err = queryClient.EpochInfos(
				context.Background(),
				&types.QueryEpochInfosRequest{},
			); err != nil {
				fmt.Printf("failed to query epoch infos: %s\n", err.Error())
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CurrentEpochCmd is the CLI command for retrieving the current epoch
func CurrentEpochCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "current [identifier]",
		Aliases: []string{"c"},
		Short:   "Show current epoch info",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var response *types.QueryCurrentEpochResponse
			if response, err = queryClient.CurrentEpoch(
				context.Background(),
				&types.QueryCurrentEpochRequest{Identifier: args[0]},
			); err != nil {
				fmt.Printf("failed to query current epoch: %s\n", err.Error())
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
