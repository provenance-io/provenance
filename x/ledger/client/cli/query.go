package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/ledger"
)

// GetQueryCmd is the top-level command for attribute CLI queries.
func CmdQuery() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        ledger.ModuleName,
		Aliases:                    []string{"l"},
		Short:                      "Querying commands for the account metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetConfigCmd(),
		GetLedgerEntriesCmd(),
	)

	return queryCmd
}

// GetAttributeParamsCmd returns the command handler for name parameter querying.
func GetConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config <nft_address>",
		Short:   "Query the ledger for the specified nft address",
		Example: fmt.Sprintf(`$ %s query attribute params`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			nftAddress := args[0]
			req := ledger.QueryLedgerConfigRequest{
				NftAddress: nftAddress,
			}

			queryClient := ledger.NewQueryClient(clientCtx)
			l, err := queryClient.Config(context.Background(), &req)
			if err != nil {
				return err
			}

			resp := ledger.QueryLedgerConfigResponse{
				Ledger: l.Ledger,
			}

			return clientCtx.PrintProto(&resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAttributeParamsCmd returns the command handler for name parameter querying.
func GetLedgerEntriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entries <nft_address>",
		Short:   "Query the ledger for the specified nft address",
		Example: fmt.Sprintf(`$ %s query attribute params`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			nftAddress := args[0]
			req := ledger.QueryLedgerRequest{
				NftAddress: nftAddress,
			}

			queryClient := ledger.NewQueryClient(clientCtx)
			l, err := queryClient.Entries(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(l)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
