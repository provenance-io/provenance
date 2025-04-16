package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/ledger"
	"github.com/provenance-io/provenance/x/ledger/keeper"
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
		GetBalancesAsOfCmd(),
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

			// Convert to PlainText
			plainText := ledger.LedgerPlainText{
				NftAddress:   l.Ledger.NftAddress,
				Denom:        l.Ledger.Denom,
				NextPmtDate:  keeper.EpochDaysToISO8601(l.Ledger.NextPmtDate),
				NextPmtAmt:   strconv.FormatInt(l.Ledger.NextPmtAmt, 10),
				Status:       l.Ledger.Status,
				InterestRate: strconv.FormatInt(int64(l.Ledger.InterestRate), 10),
				MaturityDate: keeper.EpochDaysToISO8601(l.Ledger.MaturityDate),
			}

			return clientCtx.PrintProto(&plainText)
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

			plainTextEntries := make([]*ledger.LedgerEntryPlainText, len(l.Entries))
			for i, entry := range l.Entries {
				appliedAmounts := make([]*ledger.LedgerBucketAmountPlainText, len(entry.AppliedAmounts))
				for j, amount := range entry.AppliedAmounts {
					appliedAmounts[j] = &ledger.LedgerBucketAmountPlainText{
						Bucket:     amount.Bucket,
						AppliedAmt: strconv.FormatInt(amount.AppliedAmt, 10),
						BalanceAmt: strconv.FormatInt(amount.BalanceAmt, 10),
					}
				}

				plainTextEntries[i] = &ledger.LedgerEntryPlainText{
					CorrelationId:  entry.CorrelationId,
					Sequence:       entry.Sequence,
					Type:           entry.Type,
					SubType:        entry.SubType,
					PostedDate:     keeper.EpochDaysToISO8601(entry.PostedDate),
					EffectiveDate:  keeper.EpochDaysToISO8601(entry.EffectiveDate),
					TotalAmt:       strconv.FormatInt(entry.TotalAmt, 10),
					AppliedAmounts: appliedAmounts,
				}
			}

			// Convert to a response proto for printing.
			plainTextResponse := ledger.QueryLedgerEntryResponsePlainText{
				Entries: plainTextEntries,
			}

			return clientCtx.PrintProto(&plainTextResponse)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetBalancesAsOfCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances [nft-address] [as-of-date]",
		Short: "Query balances for an NFT as of a specific date",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			nftAddress := args[0]
			asOfDate := args[1]

			// Validate the date format
			_, err = time.Parse("2006-01-02", asOfDate)
			if err != nil {
				return fmt.Errorf("invalid date format. Please use ISO8601 format (e.g., 2024-01-01): %w", err)
			}

			queryClient := ledger.NewQueryClient(clientCtx)
			res, err := queryClient.GetBalancesAsOf(cmd.Context(), &ledger.QueryBalancesAsOfRequest{
				NftAddress: nftAddress,
				AsOfDate:   asOfDate,
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
