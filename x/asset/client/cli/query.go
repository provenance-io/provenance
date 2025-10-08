package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/asset/types"
)

// GetQueryCmd returns the cli query commands for the asset module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the asset module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdAsset(),
		GetCmdAssets(),
		GetCmdAssetClass(),
		GetCmdAssetClasses(),
	)

	return cmd
}

// GetCmdAsset returns the command for getting an asset.
func GetCmdAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "asset <class-id> <asset-id>",
		Short:   "Get an asset by class id and asset id",
		Example: fmt.Sprintf(`$ %s query asset asset my-asset-class my-asset`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Asset(cmd.Context(), &types.QueryAssetRequest{ClassId: args[0], Id: args[1]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdAssets returns the command for listing all assets.
func GetCmdAssets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets [<class-id>|<address>|<class-id> <address>]",
		Short: "List all assets optionally filtered by class-id and/or address",
		Example: fmt.Sprintf(`$ %[1]s query asset assets my-asset-class pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ %[1]s query asset assets my-asset-class --page=2 --limit=100
`, version.AppName),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := &types.QueryAssetsRequest{}
			req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			switch len(args) {
			case 1:
				// Determine if the single arg is an address or a class-id.
				if _, err = sdk.AccAddressFromBech32(args[0]); err == nil {
					req.Owner = args[0]
				} else {
					req.ClassId = args[0]
				}
			case 2:
				req.ClassId = args[0]
				req.Owner = args[1]
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Assets(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "assets")
	return cmd
}

// GetCmdAssetClass returns the command for getting an asset class.
func GetCmdAssetClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "class <id>",
		Short:   "Get an asset class by id",
		Example: fmt.Sprintf(`$ %s query asset class my-asset-class`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AssetClass(cmd.Context(), &types.QueryAssetClassRequest{Id: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdAssetClasses returns the command for listing all asset classes.
func GetCmdAssetClasses() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "classes",
		Short: "List all asset classes",
		Example: fmt.Sprintf(`$ %[1]s query asset classes
$ %[1]s query asset classes --page=2 --limit=50
`, version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := &types.QueryAssetClassesRequest{}
			req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AssetClasses(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "classes")
	return cmd
}
