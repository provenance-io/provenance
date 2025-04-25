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
		Use:     "config <asset_class_id> <nft_id>",
		Short:   "Query the ledger for the specified nft id",
		Example: fmt.Sprintf(`$ %s query attribute params`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassId := args[0]
			nftId := args[1]
			req := ledger.QueryLedgerConfigRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
				},
			}

			queryClient := ledger.NewQueryClient(clientCtx)
			l, err := queryClient.Config(context.Background(), &req)
			if err != nil {
				return err
			}

			// Convert to PlainText
			plainText := ledger.LedgerPlainText{
				Key:          req.Key,
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
		Use:     "entries <asset_class_id> <nft_id>",
		Short:   "Query the ledger for the specified nft address",
		Example: fmt.Sprintf(`$ %s query attribute params`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassId := args[0]
			nftId := args[1]
			queryClient := ledger.NewQueryClient(clientCtx)

			getConfig := func(nftId string) *ledger.QueryLedgerConfigResponse {
				req := ledger.QueryLedgerConfigRequest{
					Key: &ledger.LedgerKey{
						AssetClassId: assetClassId,
						NftId:        nftId,
					},
				}

				config, err := queryClient.Config(context.Background(), &req)
				if err != nil {
					return nil
				}

				return config
			}

			getEntries := func(nftId string) []*ledger.LedgerEntry {
				req := ledger.QueryLedgerRequest{
					Key: &ledger.LedgerKey{
						AssetClassId: assetClassId,
						NftId:        nftId,
					},
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
			entryTypes := getEntryTypes(config.Ledger.LedgerClassId)
			bucketTypes := getBucketTypes(config.Ledger.LedgerClassId)

			plainTextEntries := make([]*ledger.LedgerEntryPlainText, len(entries))
			for i, entry := range entries {
				appliedAmounts := make([]*ledger.LedgerBucketAmountPlainText, len(entry.AppliedAmounts))
				for j, amount := range entry.AppliedAmounts {
					appliedAmounts[j] = &ledger.LedgerBucketAmountPlainText{
						Bucket:     bucketTypes[amount.BucketTypeId],
						AppliedAmt: amount.AppliedAmt.String(),
						BalanceAmt: "0",
					}
				}

				plainTextEntries[i] = &ledger.LedgerEntryPlainText{
					CorrelationId:  entry.CorrelationId,
					Sequence:       entry.Sequence,
					Type:           entryTypes[entry.EntryTypeId],
					PostedDate:     keeper.EpochDaysToISO8601(entry.PostedDate),
					EffectiveDate:  keeper.EpochDaysToISO8601(entry.EffectiveDate),
					TotalAmt:       entry.TotalAmt.String(),
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
		Use:   "balances <asset_class_id> <nft_id> <as_of_date>",
		Short: "Query balances for an NFT as of a specific date",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassId := args[0]
			nftId := args[1]
			asOfDate := args[2]

			// Validate the date format
			_, err = time.Parse("2006-01-02", asOfDate)
			if err != nil {
				return fmt.Errorf("invalid date format. Please use ISO8601 format (e.g., 2024-01-01): %w", err)
			}

			queryClient := ledger.NewQueryClient(clientCtx)
			res, err := queryClient.GetBalancesAsOf(cmd.Context(), &ledger.QueryBalancesAsOfRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
				},
				AsOfDate: asOfDate,
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
