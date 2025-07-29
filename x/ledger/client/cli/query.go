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

	chunkimport "github.com/provenance-io/provenance/x/ledger/client/cli/import"
	"github.com/provenance-io/provenance/x/ledger/helper"
	"github.com/provenance-io/provenance/x/ledger/keeper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

// Go structs to replace proto-generated structs from ledger_query.proto

// LedgerPlainText represents a ledger in plain text format
type LedgerPlainText struct {
	// Ledger key
	Key *ledger.LedgerKey `json:"key,omitempty"`
	// Status of the ledger
	Status string `json:"status,omitempty"`
	// Next payment date
	NextPmtDate string `json:"next_pmt_date,omitempty"`
	// Next payment amount
	NextPmtAmt string `json:"next_pmt_amt,omitempty"`
	// Interest rate
	InterestRate string `json:"interest_rate,omitempty"`
	// Maturity date
	MaturityDate string `json:"maturity_date,omitempty"`
	// Day count convention for interest
	InterestDayCountConvention ledger.DayCountConvention `json:"interest_day_count_convention,omitempty"`
	// Interest accrual method for interest
	InterestAccrualMethod ledger.InterestAccrualMethod `json:"interest_accrual_method,omitempty"`
	// Payment frequency
	PaymentFrequency ledger.PaymentFrequency `json:"payment_frequency,omitempty"`
}

// LedgerEntryPlainText represents a ledger entry in plain text format
type LedgerEntryPlainText struct {
	// Correlation ID for tracking ledger entries with external systems (max 50 characters)
	CorrelationId string `json:"correlation_id,omitempty"`
	// Sequence number of the ledger entry (less than 100)
	// This field is used to maintain the correct order of entries when multiple entries
	// share the same effective date. Entries are sorted first by effective date, then by sequence.
	Sequence uint32 `json:"sequence,omitempty"`
	// The type of ledger entry specified by the LedgerClassEntryType.id
	Type *ledger.LedgerClassEntryType `json:"type,omitempty"`
	// Posted date
	PostedDate string `json:"posted_date,omitempty"`
	// Effective date
	EffectiveDate string `json:"effective_date,omitempty"`
	// The total amount of the ledger entry
	TotalAmt string `json:"total_amt,omitempty"`
	// The amounts applied to each bucket
	AppliedAmounts []*LedgerBucketAmountPlainText `json:"applied_amounts,omitempty"`
}

// LedgerBucketAmountPlainText represents bucket amounts in plain text format
type LedgerBucketAmountPlainText struct {
	Bucket     *ledger.LedgerClassBucketType `json:"bucket,omitempty"`
	AppliedAmt string                        `json:"applied_amt,omitempty"`
	BalanceAmt string                        `json:"balance_amt,omitempty"`
}

// QueryLedgerEntryResponsePlainText represents the response for ledger entries query in plain text format
type QueryLedgerEntryResponsePlainText struct {
	Entries []*LedgerEntryPlainText `json:"entries,omitempty"`
}

// ProtoMessage methods to make structs compatible with clientCtx.PrintProto
func (m *LedgerPlainText) ProtoMessage() {}
func (m *LedgerPlainText) Reset()        { *m = LedgerPlainText{} }
func (m *LedgerPlainText) String() string {
	return fmt.Sprintf("LedgerPlainText{Key:%v, Status:%s, NextPmtDate:%s, NextPmtAmt:%s, InterestRate:%s, MaturityDate:%s, InterestDayCountConvention:%v, InterestAccrualMethod:%v, PaymentFrequency:%v}",
		m.Key, m.Status, m.NextPmtDate, m.NextPmtAmt, m.InterestRate, m.MaturityDate, m.InterestDayCountConvention, m.InterestAccrualMethod, m.PaymentFrequency)
}

func (m *LedgerEntryPlainText) ProtoMessage() {}
func (m *LedgerEntryPlainText) Reset()        { *m = LedgerEntryPlainText{} }
func (m *LedgerEntryPlainText) String() string {
	return fmt.Sprintf("LedgerEntryPlainText{CorrelationId:%s, Sequence:%d, Type:%v, PostedDate:%s, EffectiveDate:%s, TotalAmt:%s, AppliedAmounts:%v}",
		m.CorrelationId, m.Sequence, m.Type, m.PostedDate, m.EffectiveDate, m.TotalAmt, m.AppliedAmounts)
}

func (m *LedgerBucketAmountPlainText) ProtoMessage() {}
func (m *LedgerBucketAmountPlainText) Reset()        { *m = LedgerBucketAmountPlainText{} }
func (m *LedgerBucketAmountPlainText) String() string {
	return fmt.Sprintf("LedgerBucketAmountPlainText{Bucket:%v, AppliedAmt:%s, BalanceAmt:%s}",
		m.Bucket, m.AppliedAmt, m.BalanceAmt)
}

func (m *QueryLedgerEntryResponsePlainText) ProtoMessage() {}
func (m *QueryLedgerEntryResponsePlainText) Reset()        { *m = QueryLedgerEntryResponsePlainText{} }
func (m *QueryLedgerEntryResponsePlainText) String() string {
	return fmt.Sprintf("QueryLedgerEntryResponsePlainText{Entries:%v}", m.Entries)
}

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
		GetLedgerClassEntryTypesCmd(),
		GetLedgerClassStatusTypesCmd(),
		GetLedgerClassBucketTypesCmd(),
		GetLedgerClassCmd(),
		GetAllSettlementsCmd(),
		GetSettlementsByCorrelationIdCmd(),
		chunkimport.CmdBulkImportStatus(),
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
			req := ledger.QueryLedgerRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
				},
			}

			queryClient := ledger.NewQueryClient(clientCtx)
			l, err := queryClient.LedgerQuery(context.Background(), &req)
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

			getConfig := func(nftId string) *ledger.QueryLedgerResponse {
				req := ledger.QueryLedgerRequest{
					Key: &ledger.LedgerKey{
						AssetClassId: assetClassId,
						NftId:        nftId,
					},
				}

				config, err := queryClient.LedgerQuery(context.Background(), &req)
				if err != nil {
					return nil
				}

				return config
			}

			getEntries := func(nftId string) []*ledger.LedgerEntry {
				req := ledger.QueryLedgerEntriesRequest{
					Key: &ledger.LedgerKey{
						AssetClassId: assetClassId,
						NftId:        nftId,
					},
				}

				l, err := queryClient.EntriesQuery(context.Background(), &req)
				if err != nil {
					return nil
				}

				return l.Entries
			}

			getEntryTypes := func(ledgerClassId string) map[int32]*ledger.LedgerClassEntryType {
				req := ledger.QueryLedgerClassEntryTypesRequest{
					LedgerClassId: ledgerClassId,
				}

				types, err := queryClient.ClassEntryTypesQuery(context.Background(), &req)
				if err != nil {
					return nil
				}

				return keeper.SliceToMap(types.EntryTypes, func(t *ledger.LedgerClassEntryType) int32 {
					return t.Id
				})
			}

			getBucketTypes := func(ledgerClassId string) map[int32]*ledger.LedgerClassBucketType {
				req := ledger.QueryLedgerClassBucketTypesRequest{
					LedgerClassId: ledgerClassId,
				}

				types, err := queryClient.ClassBucketTypesQuery(context.Background(), &req)
				if err != nil {
					return nil
				}

				return keeper.SliceToMap(types.BucketTypes, func(t *ledger.LedgerClassBucketType) int32 {
					return t.Id
				})
			}

			config := getConfig(nftId)

			if config == nil {
				return fmt.Errorf("ledger not found for nft id: %s", nftId)
			}

			entries := getEntries(nftId)
			entryTypes := getEntryTypes(config.Ledger.LedgerClassId)
			bucketTypes := getBucketTypes(config.Ledger.LedgerClassId)

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
						if appliedAmt.Bucket.Id == balanceAmt.BucketTypeId {
							appliedAmt.BalanceAmt = balanceAmt.BalanceAmt.String()
						}
					}
				}

				plainTextEntries[i] = &LedgerEntryPlainText{
					CorrelationId:  entry.CorrelationId,
					Sequence:       entry.Sequence,
					Type:           entryTypes[entry.EntryTypeId],
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
			res, err := queryClient.BalancesAsOfQuery(cmd.Context(), &ledger.QueryBalancesAsOfRequest{
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

// GetLedgerClassEntryTypesCmd returns the command handler for querying ledger class entry types
func GetLedgerClassEntryTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entry-types <ledger_class_id>",
		Short:   "Query the ledger class entry types for the specified ledger class",
		Example: fmt.Sprintf(`$ %s query ledger entry-types pb1a2b3c4...`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassId := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			req := ledger.QueryLedgerClassEntryTypesRequest{
				LedgerClassId: ledgerClassId,
			}

			resp, err := queryClient.ClassEntryTypesQuery(context.Background(), &req)
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
		Example: fmt.Sprintf(`$ %s query ledger status-types pb1a2b3c4...`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassId := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			req := ledger.QueryLedgerClassStatusTypesRequest{
				LedgerClassId: ledgerClassId,
			}

			resp, err := queryClient.ClassStatusTypesQuery(context.Background(), &req)
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
		Example: fmt.Sprintf(`$ %s query ledger bucket-types pb1a2b3c4...`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassId := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			req := ledger.QueryLedgerClassBucketTypesRequest{
				LedgerClassId: ledgerClassId,
			}

			resp, err := queryClient.ClassBucketTypesQuery(context.Background(), &req)
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
		Short:   "Query the ledger class for the specified ledger class",
		Example: fmt.Sprintf(`$ %s query ledger class pb1a2b3c4...`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassId := args[0]
			queryClient := ledger.NewQueryClient(clientCtx)

			req := ledger.QueryLedgerClassRequest{
				LedgerClassId: ledgerClassId,
			}

			resp, err := queryClient.ClassQuery(context.Background(), &req)
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
		Example: fmt.Sprintf(`$ %s query ledger settlements pb1a2b3c4... nft-123`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassId := args[0]
			nftId := args[1]

			queryClient := ledger.NewQueryClient(clientCtx)
			res, err := queryClient.SettlementsQuery(cmd.Context(), &ledger.QuerySettlementsRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
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

func GetSettlementsByCorrelationIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "settlement <asset_class_id> <nft_id> <correlation_id>",
		Short:   "Query settlements for an NFT by correlation ID",
		Example: fmt.Sprintf(`$ %s query ledger settlement pb1a2b3c4... nft-123 correlation-456`, version.AppName),
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			assetClassId := args[0]
			nftId := args[1]
			correlationId := args[2]

			queryClient := ledger.NewQueryClient(clientCtx)
			res, err := queryClient.SettlementsByCorrelationIdQuery(cmd.Context(), &ledger.QuerySettlementsByCorrelationIdRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
				},
				CorrelationId: correlationId,
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
