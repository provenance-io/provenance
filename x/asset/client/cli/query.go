package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/asset/types"
)

// GetQueryCmd returns the cli query commands for the asset module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the asset module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdListAssets(),
		GetCmdListAssetClasses(),
		GetCmdGetClass(),
	)

	return cmd
}

// GetCmdListAssets returns the command for listing all assets
func GetCmdListAssets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-assets [address]",
		Short: "List all assets owned by an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ListAssets(cmd.Context(), &types.QueryListAssets{
				Address: args[0],
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

// GetCmdGetAsset returns the command for getting one assets
func GetCmdGetClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-class [id]",
		Short: "Gets a specific asset class by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetClass(cmd.Context(), &types.QueryGetClass{Id: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdListAssetClasses returns the command for listing all asset classes
func GetCmdListAssetClasses() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-classes",
		Short: "List all asset classes",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ListAssetClasses(cmd.Context(), &types.QueryListAssetClasses{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
