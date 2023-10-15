package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/oracle/types"
)

// GetQueryCmd is the top-level command for oracle CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the oracle module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetQueryOracleAddressCmd(),
	)
	return queryCmd
}

// GetQueryOracleAddressCmd queries for the module's oracle address
func GetQueryOracleAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "address",
		Short:   "Returns the address of the module's oracle",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"a"},
		Example: fmt.Sprintf(`%[1]s q oracle address`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryOracleAddressRequest{}

			res, err := queryClient.OracleAddress(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
