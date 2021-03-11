package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetQueryCmd returns the top-level command for marker CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetMetadataParamsCmd(),
		GetMetadataScopeCmd(),
		GetOSLocatorParamsCmd(),
		GetOSLocatorCmd(),
		GetOSLocatorByURICmd(),
	)
	return queryCmd
}

// GetMetadataParamsCmd returns the command handler for metadata parameter querying.
func GetMetadataParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current metadata parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current metadata module parameters:
Example:
$ %s query metadata params
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataScopeCmd returns the command handler for metadata scope querying.
func GetMetadataScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scope [id]",
		Short: "Query the current metadata for scope",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current metadata module parameters:
Example:
$ %s query metadata scope 123e4567-e89b-12d3-a456-426614174000
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			scopeUUID := strings.ToLower(strings.TrimSpace(args[0]))

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Scope(context.Background(), &types.ScopeRequest{ScopeUuid: scopeUUID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Scope)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetOSLocatorParamsCmd returns the command handler for metadata locator parameter querying.
func GetOSLocatorParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locator params",
		Short: "Query the current os locator parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current name module parameters:

$ %s query name params
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.OSLocatorParams(context.Background(), &types.OSLocatorQueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetOSLocatorCmd returns the command handler for querying oslocator by address .
func GetOSLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locator [owner]",
		Short: "Query the OS locator uri for the given owner",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the OS locator records for the name provided:
Example:
$ %s query metadata locator foocorp
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			owner := strings.ToLower(strings.TrimSpace(args[0]))

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.OSLocator(context.Background(), &types.OSLocatorRequest{Owner: owner})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Locator)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetOSLocatorByURICmd returns the command handler for querying oslocator by uri.
func GetOSLocatorByURICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locator [uri]",
		Short: "Query the OS locator uri for the given owner",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the OS locator records for the uri provided:
Example:
$ %s query metadata locator foocorp
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			uri := strings.ToLower(strings.TrimSpace(args[0]))
			pageReq, err := client.ReadPageRequest(cmd.Flags())

			queryClient := types.NewQueryClient(clientCtx)
			var response *types.OSLocatorResponses

			response, err = queryClient.OSLocatorByURI(context.Background(), &types.OSLocatorByURIRequest{Uri: uri, Pagination: pageReq})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetOSLocatorByURICmd returns the command handler for querying oslocator by uri.
func GetOSLocatorByScopeUUID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locator [scope_uuid]",
		Short: "Query the OS locator uri for the given scope owners",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the OS locator records for the scope uuid provided:
Example:
$ %s query metadata locator foocorp
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			scopeUUID := strings.ToLower(strings.TrimSpace(args[0]))

			queryClient := types.NewQueryClient(clientCtx)
			var response *types.OSLocatorScopeResponse

			response, err = queryClient.OSLocatorByScopeUUID(context.Background(), &types.ScopeRequest{ScopeUuid: scopeUUID})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
