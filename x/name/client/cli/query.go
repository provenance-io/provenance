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
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/name/types"
)

// GetQueryCmd is the top-level command for name CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the name module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		QueryParamsCmd(),
		ResolveNameCommand(),
		ReverseLookupCommand(),
	)

	return queryCmd
}

// QueryParamsCmd returns the command handler for name parameter querying.
func QueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Short:   "Query the current name parameters",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(`$ %s query name params`, version.AppName),
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

// ResolveNameCommand returns the command handler for resolving the address for a given name.
func ResolveNameCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resolve [name]",
		Short:   "Resolve the address for a name",
		Example: fmt.Sprintf(`$ %s query name resolve attrib.name`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			name := strings.ToLower(strings.TrimSpace(args[0]))

			var response *types.QueryResolveResponse
			if response, err = queryClient.Resolve(
				context.Background(),
				&types.QueryResolveRequest{Name: name},
			); err != nil {
				fmt.Printf("failed to query name \"%s\" for address: %v\n", name, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// ReverseLookupCommand returns the command handler for finding all names that point to an address.
func ReverseLookupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup [address]",
		Short: "Reverse lookup of all names bound to a given address",
		Example: fmt.Sprintf(`$ %[1]s query name lookup pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ %[1]s query name lookup pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk --page=2 --limit=100
`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}

			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("account address must be a Bech32 string: %w", err)
			}

			var response *types.QueryReverseLookupResponse
			if response, err = queryClient.ReverseLookup(
				context.Background(),
				&types.QueryReverseLookupRequest{Address: address.String(), Pagination: pageReq},
			); err != nil {
				fmt.Printf("failed to query reverse lookup against \"%s\": %v\n", address, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "get")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// sdk ReadPageRequest expects binary but we encoded to base64 in our marshaller
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
