package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/x/smartaccounts/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd is the top-level command for smart account CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the smartaccount module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		QueryAccountByAddressCmd(),
		QueryParamsCmd(),
	)
	return queryCmd
}

// QueryAccountByAddressCmd queries for the address for an account
func QueryAccountByAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "address",
		Short:   "Returns the smart account for the address provided.",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"a"},
		Example: fmt.Sprintf(`%[1]s q smartaccount address`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			address := strings.ToLower(strings.TrimSpace(args[0]))
			smartAccount := &types.AccountQueryRequest{Address: address}

			res, err := queryClient.SmartAccount(context.Background(), smartAccount)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryParamsCmd returns the command handler for marker parameter querying.
func QueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current smart account parameters",
		Args:  cobra.NoArgs,
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query smartaccount params`, version.AppName)),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
