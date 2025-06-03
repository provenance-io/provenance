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
		Use:   "registry [asset_class_id] [nft_id]",
		Short: "Query a registry entry by key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			key := registry.RegistryKey{
				AssetClassId: args[0],
				NftId:        args[1],
			}

			queryClient := registry.NewQueryClient(clientCtx)

			res, err := queryClient.GetRegistry(context.Background(), &registry.QueryGetRegistryRequest{
				Key: &key,
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
		Use:   "has-role [asset_class_id] [nft_id] [role] [address]",
		Short: "Query if an address has a role for a given key",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			key := registry.RegistryKey{
				AssetClassId: args[0],
				NftId:        args[1],
			}

			// convert arg[2] to a registry.RegistryRole enum value
			role, ok := registry.RegistryRole_value[args[2]]
			if !ok {
				return fmt.Errorf("invalid role: %s", args[2])
			}

			queryClient := registry.NewQueryClient(clientCtx)

			// TODO: Parse key and role from args
			res, err := queryClient.HasRole(context.Background(), &registry.QueryHasRoleRequest{
				Key:     &key,
				Role:    registry.RegistryRole(role),
				Address: args[3],
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
