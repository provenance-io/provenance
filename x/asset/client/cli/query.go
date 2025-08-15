package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

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
		Use:   "list-assets <address>",
		Short: "List all assets owned by an address",
		Example: fmt.Sprintf(`$ %[1]s query asset list-assets pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ %[1]s query asset list-assets pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk --page=2 --limit=100
`, version.AppName),
		Args: cobra.ExactArgs(1),
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

			res, err := queryClient.ListAssets(cmd.Context(), &types.QueryListAssets{
				Address:    args[0],
				Pagination: pageReq,
			})
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

// GetCmdGetAsset returns the command for getting one assets
func GetCmdGetClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-class [id]",
		Short:   "Gets a specific asset class by id",
		Example: fmt.Sprintf(`$ %s query asset get-class my-asset-class`, version.AppName),
		Args:    cobra.ExactArgs(1),
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
		Example: fmt.Sprintf(`$ %[1]s query asset list-classes
$ %[1]s query asset list-classes --page=2 --limit=50
`, version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.ListAssetClasses(cmd.Context(), &types.QueryListAssetClasses{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "asset classes")
	return cmd
}
