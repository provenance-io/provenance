package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// GetQueryCmd returns the top-level command for msgfees CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the msgfees module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		AllMsgFeesCmd(),
		ListParamsCmd(),
	)
	return queryCmd
}

// AllMsgFeesCmd is the CLI command for listing all msg fees.
func AllMsgFeesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List all the msg fees on the Provenance Blockchain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			var response *types.QueryAllMsgFeesResponse
			if response, err = queryClient.QueryAllMsgFees(
				context.Background(),
				&types.QueryAllMsgFeesRequest{Pagination: pageReq},
			); err != nil {
				fmt.Printf("failed to query msg fees: %s\n", err.Error())
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "msgfees")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ListParamsCmd is the CLI command for listing all params.
func ListParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Aliases: []string{"p"},
		Short:   "List the msg fees params on the Provenance Blockchain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			if err != nil {
				return err
			}

			var response *types.QueryParamsResponse
			if response, err = queryClient.Params(
				context.Background(),
				&types.QueryParamsRequest{},
			); err != nil {
				fmt.Printf("failed to query msg fees params: %s\n", err.Error())
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
