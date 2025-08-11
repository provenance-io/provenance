// Package cli provides command-line interface functionality for the ledger module.
// This file contains query commands for interacting with ledger data, including
// commands to query ledger configurations, entries, balances, and settlements.
// The query commands provide both raw ledger data and human-readable plain text
// representations of ledger information.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/ledger/helper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
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
		GetCmd(),
		GetLedgerEntriesCmd(),
		GetBalancesAsOfCmd(),
		GetLedgerClassEntryTypesCmd(),
		GetLedgerClassStatusTypesCmd(),
		GetLedgerClassBucketTypesCmd(),
		GetLedgerClassCmd(),
		GetAllSettlementsCmd(),
		GetSettlementsByCorrelationIDCmd(),
	)

	return queryCmd
}

// Get a ledger's base information
func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <asset_class_id> <nft_id>",
		Short:   "Query the ledger for the specified asset class and nft id",
		Example: fmt.Sprintf(`$ %s query ledger get class-123 nft-123`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassID := args[0]
			nftID := args[1]
			req := ledger.QueryLedgerRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassID,
					NftId:        nftID,
				},
			}

			queryClient := ledger.NewQueryClient(clientCtx)
			l, err := queryClient.Ledger(context.Background(), &req)
			if err != nil {
				return err
			}

			// Convert to PlainText
			plainText := LedgerPlainText{
				Key:                        req.Key,
				Status:                     strconv.Itoa(int(l.Ledger.StatusTypeId)),
				NextPmtDate:                helper.EpochDaysToISO8601(l.Ledger.NextPmtDate),
				NextPmtAmt:                 strconv.FormatInt(l.Ledger.NextPmtAmt, 10),
				InterestRate:               strconv.FormatInt(int64(l.Ledger.InterestRate), 10),
				MaturityDate:               helper.EpochDaysToISO8601(l.Ledger.MaturityDate),
				InterestDayCountConvention: l.Ledger.InterestDayCountConvention,
				InterestAccrualMethod:      l.Ledger.InterestAccrualMethod,
				PaymentFrequency:           l.Ledger.PaymentFrequency,
			}

			// PRint the struct as json
			json, err := json.MarshalIndent(plainText, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(json))

			return nil
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
		Example: fmt.Sprintf(`$ %s query ledger entries class-123 nft-123`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassID := args[0]
			nftID := args[1]
			queryClient := ledger.NewQueryClient(clientCtx)

			getConfig := func(nftID string) *ledger.QueryLedgerResponse {
				req := ledger.QueryLedgerRequest{
					Key: &ledger.LedgerKey{
						AssetClassId: assetClassID,
						NftId:        nftID,
					},
				}

				config, err := queryClient.Ledger(context.Background(), &req)
				if err != nil {
					return nil
				}

				return config
			}

			getEntries := func(nftID string) []*ledger.LedgerEntry {
				req := ledger.QueryLedgerEntriesRequest{
					Key: &ledger.LedgerKey{
						AssetClassId: assetClassID,
						NftId:        nftID,
					},
				}

				l, err := queryClient.LedgerEntries(context.Background(), &req)
				if err != nil {
					return nil
				}

				return l.Entries
			}

			getEntryTypes := func(ledgerClassID string) map[int32]*ledger.LedgerClassEntryType {
				req := ledger.QueryLedgerClassEntryTypesRequest{
					LedgerClassId: ledgerClassID,
				}

				types, err := queryClient.LedgerClassEntryTypes(context.Background(), &req)
				if err != nil {
					return nil
				}

				return sliceToMap(types.EntryTypes, func(t *ledger.LedgerClassEntryType) int32 {
					return t.Id
				})
			}

			getBucketTypes := func(ledgerClassID string) map[int32]*ledger.LedgerClassBucketType {
				req := ledger.QueryLedgerClassBucketTypesRequest{
					LedgerClassId: ledgerClassID,
				}

				types, err := queryClient.LedgerClassBucketTypes(context.Background(), &req)
				if err != nil {
					return nil
				}

				return sliceToMap(types.BucketTypes, func(t *ledger.LedgerClassBucketType) int32 {
					return t.Id
				})
			}

			config := getConfig(nftID)

			if config == nil {
				return fmt.Errorf("ledger not found for nft id: %s", nftID)
			}

			entries := getEntries(nftID)
			entryTypes := getEntryTypes(config.Ledger.LedgerClassId)
			bucketTypes := getBucketTypes(config.Ledger.LedgerClassId)

			// Check if we successfully retrieved the entry types and bucket types
			if entryTypes == nil {
				return fmt.Errorf("failed to retrieve entry types for ledger class: %s", config.Ledger.LedgerClassId)
			}
			if bucketTypes == nil {
				return fmt.Errorf("failed to retrieve bucket types for ledger class: %s", config.Ledger.LedgerClassId)
			}

			plainTextEntries := make([]*LedgerEntryPlainText, len(entries))
			for i, entry := range entries {
				appliedAmts := make([]*LedgerBucketAmountPlainText, len(entry.AppliedAmounts))
				for j, amt := range entry.AppliedAmounts {
					appliedAmts[j] = &LedgerBucketAmountPlainText{
						Bucket:     bucketTypes[amt.BucketTypeId],
						AppliedAmt: amt.AppliedAmt.String(),
					}
				}

				for _, balanceAmt := range entry.BalanceAmounts {
					for _, appliedAmt := range appliedAmts {
						if appliedAmt.Bucket != nil && appliedAmt.Bucket.Id == balanceAmt.BucketTypeId {
							appliedAmt.BalanceAmt = balanceAmt.BalanceAmt.String()
						}
					}
				}

				entryType := entryTypes[entry.EntryTypeId]
				if entryType == nil {
					return fmt.Errorf("entry type not found for id: %d", entry.EntryTypeId)
				}

				plainTextEntries[i] = &LedgerEntryPlainText{
					CorrelationID:  entry.CorrelationId,
					Sequence:       entry.Sequence,
					Type:           entryType,
					PostedDate:     helper.EpochDaysToISO8601(entry.PostedDate),
					EffectiveDate:  helper.EpochDaysToISO8601(entry.EffectiveDate),
					TotalAmt:       entry.TotalAmt.String(),
					AppliedAmounts: appliedAmts,
				}
			}

			// Convert to a response proto for printing.
			plainTextResponse := QueryLedgerEntryResponsePlainText{
				Entries: plainTextEntries,
			}

			json, err := json.MarshalIndent(plainTextResponse, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(json))

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetBalancesAsOfCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "balances <asset_class_id> <nft_id> <as_of_date>",
		Short:   "Query balances for an NFT as of a specific date",
		Example: fmt.Sprintf(`$ %s query ledger balances class-123 nft-123 2024-01-01`, version.AppName),
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassID := args[0]
			nftID := args[1]
			asOfDate := args[2]

			// Validate the date format
			_, err = time.Parse("2006-01-02", asOfDate)
			if err != nil {
				return fmt.Errorf("invalid date format. Please use ISO8601 format (e.g., 2024-01-01): %w", err)
			}

			queryClient := ledger.NewQueryClient(clientCtx)
			res, err := queryClient.LedgerBalancesAsOf(cmd.Context(), &ledger.QueryLedgerBalancesAsOfRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassID,
					NftId:        nftID,
				},
				AsOfDate: asOfDate,
			})
			if err != nil {
				return err
			}

			json, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(json))

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetLedgerClassEntryTypesCmd returns the command handler for querying ledger class entry types
func GetLedgerClassEntryTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entry-types <ledger_class_id>",
		Short:   "Query the entry types for the specified ledger class",
		Example: fmt.Sprintf(`$ %s query ledger entry-types class-123`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassID := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			req := ledger.QueryLedgerClassEntryTypesRequest{
				LedgerClassId: ledgerClassID,
			}

			resp, err := queryClient.LedgerClassEntryTypes(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetLedgerClassStatusTypesCmd returns the command handler for querying ledger class status types
func GetLedgerClassStatusTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status-types <ledger_class_id>",
		Short:   "Query the ledger class status types for the specified ledger class",
		Example: fmt.Sprintf(`$ %s query ledger status-types class-123`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassID := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			req := ledger.QueryLedgerClassStatusTypesRequest{
				LedgerClassId: ledgerClassID,
			}

			resp, err := queryClient.LedgerClassStatusTypes(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetLedgerClassBucketTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bucket-types <ledger_class_id>",
		Short:   "Query the ledger class bucket types for the specified ledger class",
		Example: fmt.Sprintf(`$ %s query ledger bucket-types class-123`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassID := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			req := ledger.QueryLedgerClassBucketTypesRequest{
				LedgerClassId: ledgerClassID,
			}

			resp, err := queryClient.LedgerClassBucketTypes(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetLedgerClassCmd returns the command handler for querying a ledger class
func GetLedgerClassCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "class <ledger_class_id>",
		Short:   "Query the ledger class for the specified id",
		Example: fmt.Sprintf(`$ %s query ledger class class-123`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassID := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			req := ledger.QueryLedgerClassRequest{
				LedgerClassId: ledgerClassID,
			}

			resp, err := queryClient.LedgerClass(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetAllSettlementsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "settlements <asset_class_id> <nft_id>",
		Short:   "Query all settlements for an NFT",
		Example: fmt.Sprintf(`$ %s query ledger settlements class-123 nft-123`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassID := args[0]
			nftID := args[1]

			queryClient := ledger.NewQueryClient(clientCtx)
			res, err := queryClient.LedgerSettlements(cmd.Context(), &ledger.QueryLedgerSettlementsRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassID,
					NftId:        nftID,
				},
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

func GetSettlementsByCorrelationIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "settlement <asset_class_id> <nft_id> <correlation_id>",
		Short:   "Query settlements for an NFT by correlation ID",
		Example: fmt.Sprintf(`$ %s query ledger settlement class-123 nft-123 correlation-456`, version.AppName),
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassID := args[0]
			nftID := args[1]
			correlationID := args[2]

			queryClient := ledger.NewQueryClient(clientCtx)
			res, err := queryClient.LedgerSettlementsByCorrelationId(cmd.Context(), &ledger.QueryLedgerSettlementsByCorrelationIdRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassID,
					NftId:        nftID,
				},
				CorrelationId: correlationID,
			})
			if err != nil {
				fmt.Println("error", err)
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
