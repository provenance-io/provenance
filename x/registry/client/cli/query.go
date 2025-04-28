package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/provenance-io/provenance/x/registry"
	"github.com/spf13/cobra"
)

// CmdQuery returns the cli query commands for the registry module
func CmdQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        registry.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", registry.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryRegistry(),
		GetCmdQueryHasRole(),
	)

	return cmd
}

// GetCmdQueryRegistry returns the command for querying a registry entry
func GetCmdQueryRegistry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry [key]",
		Short: "Query a registry entry by key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := registry.NewQueryClient(clientCtx)

			// TODO: Parse key from args[0]
			res, err := queryClient.GetRegistry(context.Background(), &registry.QueryGetRegistryRequest{
				Key: nil, // Need to parse key from args[0]
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryHasRole returns the command for querying if an address has a role
func GetCmdQueryHasRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "has-role [key] [address] [role]",
		Short: "Query if an address has a role for a given key",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := registry.NewQueryClient(clientCtx)

			// TODO: Parse key and role from args
			res, err := queryClient.HasRole(context.Background(), &registry.QueryHasRoleRequest{
				Key:     nil, // Need to parse key from args[0]
				Address: args[1],
				Role:    0, // Need to parse role from args[2]
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
