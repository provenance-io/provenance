package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        ibcratelimit.ModuleName,
		Short:                      "Querying commands for the ibcratelimit module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetParamsCmd(),
	)

	return queryCmd
}

// GetParamsCmd returns the command handler for ibcratelimit parameter querying.
func GetParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Short:   "Query the current ibcratelimit params",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(`$ %s query ibcratelimit params`, version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := ibcratelimit.NewQueryClient(clientCtx)
			res, err := queryClient.Params(context.Background(), &ibcratelimit.ParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
