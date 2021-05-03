package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

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
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current marker module parameters:

$ %s query marker params
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

// AllMarkersCmd is the CLI command for listing all marker module registrations.
func AllMarkersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [status, optional]",
		Short: "List all marker registrations on the Provenance Blockchain",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
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

	cmd.Flags().Uint32(flags.FlagPage, 1, "Query a specific page of paginated results")
	cmd.Flags().Uint32(flags.FlagLimit, 200, "Query number of results per page returned")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// AllHoldersCmd is the CLI command for listing all marker module registrations.
func AllHoldersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "holding [denom]",
		Aliases: []string{"hold", "holder"},
		Short:   "List all accounts holding the given marker on the Provenance Blockchain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			id := strings.ToLower(strings.TrimSpace(args[0]))
			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
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

	cmd.Flags().Uint32(flags.FlagPage, 1, "Query a specific page of paginated results")
	cmd.Flags().Uint32(flags.FlagLimit, 200, "Query number of results per page returned")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// MarkerCmd is the CLI command for querying marker module registrations.
func MarkerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [address|denom]",
		Short: "Get marker details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			id := strings.ToLower(strings.TrimSpace(args[0]))

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
		Use:   "grants [address|denom]",
		Short: "Get access grants defined for marker",
		Args:  cobra.ExactArgs(1),
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
		Use:   "escrow [address|denom]",
		Short: "Get coins in escrow by marker",
		Args:  cobra.ExactArgs(1),
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
		Use:   "supply [address|denom]",
		Short: "Get total supply for marker",
		Args:  cobra.ExactArgs(1),
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
