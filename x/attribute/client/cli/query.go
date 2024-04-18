package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// GetQueryCmd is the top-level command for attribute CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"am"},
		Short:                      "Querying commands for the account metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetAttributeParamsCmd(),
		GetAccountAttributeCmd(),
		ListAccountAttributesCmd(),
		ScanAccountAttributesCmd(),
		GetAttributeAccountsCmd(),
		GetAccountDataCmd(),
	)

	return queryCmd
}

// GetAttributeParamsCmd returns the command handler for name parameter querying.
func GetAttributeParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Short:   "Query the current name parameters",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(`$ %s query attribute params`, version.AppName),
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

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAccountAttributeCmd gets all account attributes by name.
func GetAccountAttributeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <address> <name>",
		Short: "Get account attributes by name",
		Example: strings.TrimSpace(
			fmt.Sprintf(`
				$ %[1]s query attribute get pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk attrib.name
				$ %[1]s query attribute get pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk attrib.name --page=2 --limit=100
				`,
				version.AppName,
			)),
		Args: cobra.ExactArgs(2),
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

			address := strings.ToLower(strings.TrimSpace(args[0]))
			name := strings.ToLower(strings.TrimSpace(args[1]))

			var response *types.QueryAttributeResponse
			if response, err = queryClient.Attribute(
				context.Background(),
				&types.QueryAttributeRequest{Account: address, Name: name, Pagination: pageReq},
			); err != nil {
				fmt.Printf("failed to query account \"%s\" attributes for name \"%s\": %v\n", address, name, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "get")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// ListAccountAttributesCmd gets all account attributes.
func ListAccountAttributesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <address>",
		Short: "Get all account attributes",
		Example: strings.TrimSpace(
			fmt.Sprintf(`
				$ %[1]s query attribute list pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
				$ %[1]s query attribute list pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk --page=2 --limit=100
				`,
				version.AppName,
			)),
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

			address := strings.ToLower(strings.TrimSpace(args[0]))
			var response *types.QueryAttributesResponse
			if response, err = queryClient.Attributes(
				context.Background(),
				&types.QueryAttributesRequest{Account: address, Pagination: pageReq},
			); err != nil {
				fmt.Printf("failed to query account \"%s\" attributes: %v\n", address, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "list")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// ScanAccountAttributesCmd gets account attributes by name suffix.
func ScanAccountAttributesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan <address> <suffix>",
		Short: "Scan account attributes by name suffix",
		Example: strings.TrimSpace(
			fmt.Sprintf(`
				$ %[1]s query attribute scan pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk name.suffix
				$ %[1]s query attribute scan pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk name.suffix --page=2 --limit=100
				`,
				version.AppName,
			)),
		Args: cobra.ExactArgs(2),
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
			address := strings.ToLower(strings.TrimSpace(args[0]))
			suffix := strings.ToLower(strings.TrimSpace(args[1]))

			var response *types.QueryScanResponse
			if response, err = queryClient.Scan(
				context.Background(),
				&types.QueryScanRequest{Account: address, Suffix: suffix, Pagination: pageReq},
			); err != nil {
				fmt.Printf("failed to query account \"%s\" attributes for suffix \"%s\": %v\n", address, suffix, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "scan")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAttributeAccountsCmd gets account addresses with attribute name
func GetAttributeAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts <name>",
		Short: "List account addresses that have attributes with name",
		Example: strings.TrimSpace(
			fmt.Sprintf(`
				$ %[1]s query attribute accounts example.provenance.io 
				$ %[1]s query attribute accounts example.provenance.io --page=2 --limit=100
				`,
				version.AppName,
			)),
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
			attributeName := strings.ToLower(strings.TrimSpace(args[0]))

			var response *types.QueryAttributeAccountsResponse
			if response, err = queryClient.AttributeAccounts(
				context.Background(),
				&types.QueryAttributeAccountsRequest{AttributeName: attributeName, Pagination: pageReq},
			); err != nil {
				fmt.Printf("failed to query attribute name \"%s\" : %v\n", attributeName, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "accounts")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAccountDataCmd gets data for an account
func GetAccountDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-data <addr>",
		Short:   "Look up account data",
		Aliases: []string{"accountdata", "ad"},
		Example: fmt.Sprintf(`$ %[1]s query attribute account-data pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAccountDataRequest{Account: strings.TrimSpace(args[0])}

			response, err := queryClient.AccountData(context.Background(), req)
			if err != nil {
				return fmt.Errorf("failed to query account data for %q: %w", req.Account, err)
			}

			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
