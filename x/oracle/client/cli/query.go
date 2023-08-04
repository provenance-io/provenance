package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/provenance-io/provenance/x/oracle/types"
	"github.com/spf13/cobra"
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
		GetQueryStateCmd(),
		GetQueryOracleAddressCmd(),
	)
	return queryCmd
}

var _ = strconv.Itoa(0)

// GetQueryStateCmd queries for the state of an existing oracle query
func GetQueryStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query-state [sequence]",
		Short:   "Returns the request and response of an ICQ query given the packet sequence",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"qs", "state"},
		Example: fmt.Sprintf(`%[1]s q oracle query-state 1`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			sequence, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid sequence: %w", err)
			}
			params := &types.QueryQueryStateRequest{
				Sequence: sequence,
			}

			res, err := queryClient.QueryState(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
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
