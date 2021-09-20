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

	"github.com/provenance-io/provenance/x/marker/types"
)

// GetQueryCmd returns the top-level command for marker CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the marker module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		QueryParamsCmd(),
		AllMarkersCmd(),
		AllHoldersCmd(),
		MarkerCmd(),
		MarkerAccessCmd(),
		MarkerEscrowCmd(),
		MarkerSupplyCmd(),
	)
	return queryCmd
}

// QueryParamsCmd returns the command handler for marker parameter querying.
func QueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current marker parameters",
		Args:  cobra.NoArgs,
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query marker params`, version.AppName)),
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

// AllMarkersCmd is the CLI command for listing all marker module registrations.
func AllMarkersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [status, optional]",
		Short: "List all marker registrations on the Provenance Blockchain",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query marker list`, version.AppName)),
		Args: cobra.RangeArgs(0, 1),
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

			var status types.MarkerStatus
			if len(args) > 0 {
				status, err = types.MarkerStatusFromString(args[0])
				if err != nil {
					fmt.Printf("expected one of 'proposed,finalized,active,cancelled,destroyed\n")
					return err
				}
			}

			var response *types.QueryAllMarkersResponse
			if response, err = queryClient.AllMarkers(
				context.Background(),
				&types.QueryAllMarkersRequest{Status: status, Pagination: pageReq},
			); err != nil {
				fmt.Printf("failed to query markers: %s\n", err.Error())
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "markers")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// AllHoldersCmd is the CLI command for listing all marker module registrations.
func AllHoldersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "holding [denom]",
		Aliases: []string{"hold", "holder"},
		Short:   "List all accounts holding the given marker on the Provenance Blockchain",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query marker holding nhash`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			id := strings.ToLower(strings.TrimSpace(args[0]))
			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			var response *types.QueryHoldingResponse
			if response, err = queryClient.Holding(
				context.Background(),
				&types.QueryHoldingRequest{
					Id:         id,
					Pagination: pageReq,
				},
			); err != nil {
				fmt.Printf("failed to query blockchain balances for \"%s\": %v\n", id, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "markers")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// MarkerCmd is the CLI command for querying marker module registrations.
func MarkerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get [address|denom]",
		Short:   "Get marker details",
		Long:    `Note: the address is for the base_account of the denom should you choose to use the address rather than the denom name`,
		Example: fmt.Sprintf(`$ %s query marker get "nhash"`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			id := strings.TrimSpace(args[0])

			var response *types.QueryMarkerResponse
			if response, err = queryClient.Marker(
				context.Background(),
				&types.QueryMarkerRequest{Id: id},
			); err != nil {
				fmt.Printf("failed to query marker \"%s\" details: %v\n", id, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// MarkerAccessCmd is the CLI command for querying marker access list.
func MarkerAccessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grants [address|denom]",
		Short:   "Get access grants defined for marker",
		Example: fmt.Sprintf(`$ %s query marker grants "nhash"`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			id := strings.ToLower(strings.TrimSpace(args[0]))

			var response *types.QueryAccessResponse
			if response, err = queryClient.Access(
				context.Background(),
				&types.QueryAccessRequest{Id: id},
			); err != nil {
				fmt.Printf("failed to query marker \"%s\" for access control list: %v\n", id, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// MarkerEscrowCmd is the CLI command for querying marker module registrations.
func MarkerEscrowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "escrow [address|denom]",
		Short:   "Get coins in escrow by marker",
		Example: fmt.Sprintf(`$ %s query marker escrow "nhash"`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			id := strings.ToLower(strings.TrimSpace(args[0]))

			var response *types.QueryEscrowResponse
			if response, err = queryClient.Escrow(
				context.Background(),
				&types.QueryEscrowRequest{Id: id},
			); err != nil {
				fmt.Printf("failed to query marker \"%s\" for escrow balances: %v\n", id, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// MarkerSupplyCmd is the CLI command for querying marker module registrations.
func MarkerSupplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "supply [address|denom]",
		Short:   "Get total supply for marker",
		Example: fmt.Sprintf(`$ %s query marker supply "nhash"`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			id := strings.ToLower(strings.TrimSpace(args[0]))

			var response *types.QuerySupplyResponse
			if response, err = queryClient.Supply(
				context.Background(),
				&types.QuerySupplyRequest{Id: id},
			); err != nil {
				fmt.Printf("failed to query marker \"%s\" for total supply configuration: %v\n", id, err)
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}
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
