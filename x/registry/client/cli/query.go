package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/provenance-io/provenance/x/registry/types"
)

// CmdQuery returns the cli query commands for the registry module
func CmdQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the registry module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryRegistry(),
		GetCmdQueryRegistries(),
		GetCmdQueryHasRole(),
		GetCmdQueryPendingRoleChange(),
		GetCmdQueryPendingRoleChanges(),
	)

	return cmd
}

// GetCmdQueryRegistry returns the command for querying a registry entry
func GetCmdQueryRegistry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <asset_class_id> <nft_id>",
		Short: "Query a registry entry by key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetRegistry(context.Background(), &types.QueryGetRegistryRequest{
				Key: &types.RegistryKey{
					AssetClassId: args[0],
					NftId:        args[1],
				},
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

// GetCmdQueryRegistries returns the command for querying all registry entries
func GetCmdQueryRegistries() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-all [asset_class_id]",
		Short: "Query all registry entries, optionally by asset class id",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			var assetClassID string
			if len(args) == 1 {
				assetClassID = args[0]
			}

			pageReq, err := client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.GetRegistries(context.Background(), &types.QueryGetRegistriesRequest{
				AssetClassId: assetClassID,
				Pagination:   pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "registries")
	return cmd
}

// GetCmdQueryHasRole returns the command for querying if an address has a role
func GetCmdQueryHasRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "has-role <asset_class_id> <nft_id> <role> <address>",
		Short: "Query if an address has a role for a given key",
		// TODO[danny]: Create Long message that utilizes ValidRolesString() to list the roles available.
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// convert arg[2] to a registry.RegistryRole enum value
			role, err := types.ParseRegistryRole(args[2])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.HasRole(context.Background(), &types.QueryHasRoleRequest{
				Key: &types.RegistryKey{
					AssetClassId: args[0],
					NftId:        args[1],
				},
				Role:    role,
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

// GetCmdQueryPendingRoleChange returns the command for querying a single pending role change by id
func GetCmdQueryPendingRoleChange() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-role-change <id>",
		Short: "Query a pending role change by its id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PendingRoleChange(context.Background(), &types.QueryPendingRoleChangeRequest{
				Id: args[0],
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

// GetCmdQueryPendingRoleChanges returns the command for querying pending role changes
func GetCmdQueryPendingRoleChanges() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-role-changes [asset_class_id] [nft_id]",
		Short: "Query pending role changes, optionally filtered by registry key",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 0 && len(args) != 2 {
				return fmt.Errorf("accepts either 0 args (all pending changes) or 2 args (asset_class_id and nft_id), received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			var key *types.RegistryKey
			if len(args) == 2 {
				key = &types.RegistryKey{
					AssetClassId: args[0],
					NftId:        args[1],
				}
			}

			pageReq, err := client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.PendingRoleChanges(context.Background(), &types.QueryPendingRoleChangesRequest{
				Key:        key,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "pending-role-changes")
	return cmd
}
