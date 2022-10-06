package cli

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/expiration/types"
)

// GetQueryCmd is the top-level command for expiration CLI queries
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"exp"},
		Short:                      "Querying commands for the expiration module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetQueryParamsCmd(),
		GetExpirationCmd(),
		GetAllExpirationsCmd(),
		GetAllExpirationsByOwnerCmd(),
		GetAllExpiredExpirationsCmd(),
	)

	return queryCmd
}

var queryCmdStr = fmt.Sprintf(`%s query expiration`, version.AppName)

// GetQueryParamsCmd is the CLI command for expiration parameter querying
func GetQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Aliases: []string{"p"},
		Short:   "Query the current expiration parameters",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(`$ %s params`, queryCmdStr),
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

// GetExpirationCmd is the CLI command for listing a specific expiration
func GetExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get [module-asset-id]",
		Aliases: []string{"g"},
		Short:   "Query expiration for a module asset",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(`$ %s get pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk`, queryCmdStr),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryExpirationRequest{ModuleAssetId: strings.TrimSpace(args[0])}
			res, err := queryClient.Expiration(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAllExpirationsCmd is the CLI command for listing all expirations
func GetAllExpirationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "all",
		Aliases: []string{"a"},
		Short:   "Query all expirations",
		Args:    cobra.NoArgs,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %[1]s all
$ %[1]s all --page=2 --limit=100
                `,
				queryCmdStr,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryAllExpirationsRequest{Pagination: pageReq}
			res, err := queryClient.AllExpirations(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "expirations (all)")

	return cmd
}

// GetAllExpirationsByOwnerCmd is the CLI command for listing all expirations belonging to a particular owner
func GetAllExpirationsByOwnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "owner [address]",
		Aliases: []string{"o"},
		Short:   "Query all expirations by owner",
		Args:    cobra.ExactArgs(1),
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %[1]s owner pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ %[1]s owner pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk --page=2 --limit=100
                `,
				queryCmdStr,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryAllExpirationsByOwnerRequest{Owner: strings.TrimSpace(args[0]), Pagination: pageReq}
			res, err := queryClient.AllExpirationsByOwner(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "expirations (by owner)")

	return cmd
}

// GetAllExpiredExpirationsCmd is the CLI command for listing all expired expirations
func GetAllExpiredExpirationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "expired",
		Aliases: []string{"e"},
		Short:   "Query all expired expirations",
		Args:    cobra.NoArgs,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %[1]s expired
$ %[1]s expired --page=2 --limit=100
                `,
				queryCmdStr,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryAllExpiredExpirationsRequest{Pagination: pageReq}
			res, err := queryClient.AllExpiredExpirations(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "expirations (expired)")

	return cmd
}

// sdk ReadPageRequest expects binary, but we encoded to base64 in our marshaller
func withPageKeyDecoded(flagSet *flag.FlagSet) *flag.FlagSet {
	encoded, err := flagSet.GetString(flags.FlagPageKey)
	if err != nil {
		panic(err.Error())
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err.Error())
	}
	_ = flagSet.Set(flags.FlagPageKey, string(raw))
	return flagSet
}
