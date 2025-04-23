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
				NftId:        l.Ledger.NftAddress,
				Status:       strconv.Itoa(int(l.Ledger.StatusTypeId)),
				NextPmtDate:  keeper.EpochDaysToISO8601(l.Ledger.NextPmtDate),
				NextPmtAmt:   strconv.FormatInt(l.Ledger.NextPmtAmt, 10),
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

			nftId := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			getConfig := func(nftId string) *ledger.QueryLedgerConfigResponse {
				req := ledger.QueryLedgerConfigRequest{
					NftAddress: nftId,
				}

				config, err := queryClient.Config(context.Background(), &req)
				if err != nil {
					return nil
				}

				return config
			}

			getEntries := func(nftId string) []*ledger.LedgerEntry {
				req := ledger.QueryLedgerRequest{
					NftAddress: nftId,
				}

				l, err := queryClient.Entries(context.Background(), &req)
				if err != nil {
					return nil
				}

				return l.Entries
			}

			getEntryTypes := func(assetClassId string) map[int32]*ledger.LedgerClassEntryType {
				req := ledger.QueryLedgerClassEntryTypesRequest{
					AssetClassId: assetClassId,
				}

				types, err := queryClient.ClassEntryTypes(context.Background(), &req)
				if err != nil {
					return nil
				}

				return keeper.SliceToMap(types.EntryTypes, func(t *ledger.LedgerClassEntryType) int32 {
					return t.Id
				})
			}

			getBucketTypes := func(assetClassId string) map[int32]*ledger.LedgerClassBucketType {
				req := ledger.QueryLedgerClassBucketTypesRequest{
					AssetClassId: assetClassId,
				}

				types, err := queryClient.ClassBucketTypes(context.Background(), &req)
				if err != nil {
					return nil
				}

				return keeper.SliceToMap(types.BucketTypes, func(t *ledger.LedgerClassBucketType) int32 {
					return t.Id
				})
			}

			config := getConfig(nftId)
			entries := getEntries(nftId)
			entryTypes := getEntryTypes(config.Ledger.AssetClassId)
			bucketTypes := getBucketTypes(config.Ledger.AssetClassId)

			plainTextEntries := make([]*ledger.LedgerEntryPlainText, len(entries))
			for i, entry := range entries {
				appliedAmounts := make([]*ledger.LedgerBucketAmountPlainText, len(entry.AppliedAmounts))
				for j, amount := range entry.AppliedAmounts {
					appliedAmounts[j] = &ledger.LedgerBucketAmountPlainText{
						Bucket:     bucketTypes[amount.BucketTypeId],
						AppliedAmt: strconv.FormatInt(amount.AppliedAmt, 10),
						BalanceAmt: "0",
					}
				}

				plainTextEntries[i] = &ledger.LedgerEntryPlainText{
					CorrelationId:  entry.CorrelationId,
					Sequence:       entry.Sequence,
					Type:           entryTypes[entry.EntryTypeId],
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
