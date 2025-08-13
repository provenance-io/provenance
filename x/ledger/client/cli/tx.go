package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/provenance-io/provenance/x/ledger/helper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

// CmdTx creates the tx command (and sub-commands) for the exchange module.
func CmdTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        ledger.ModuleName,
		Aliases:                    []string{"l"},
		Short:                      "Transaction commands for the ledger module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreate(),
		CmdAppend(),
		CmdDestroy(),
		CmdCreateLedgerClass(),
		CmdAddLedgerClassStatusType(),
		CmdAddLedgerClassEntryType(),
		CmdAddLedgerClassBucketType(),
		CmdTransferFundsWithSettlement(),
		CmdBulkCreate(),
	)

	return cmd
}

func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <asset_class_id> <nft_id> <ledger_class_id> <status_type_id>",
		Aliases: []string{},
		Short:   "Create a ledger for the nft_address",
		Example: `$ provenanced tx ledger create "asset-class-1" "nft-1" "ledger-class-1" 1 --next-pmt-date "2024-12-31" --next-pmt-amt 1000 --interest-rate 5000000 --maturity-date "2025-12-31" --from mykey
$ provenanced tx ledger create "asset-class-1" "nft-1" "ledger-class-1" 1 --from mykey  # minimal example with required fields only`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			assetClassID := args[0]
			nftID := args[1]
			ledgerClassID := args[2]
			statusTypeID, _ := strconv.ParseInt(args[3], 10, 32)

			// Create the ledger with required fields
			ledgerObj := &ledger.Ledger{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassID,
					NftId:        nftID,
				},
				LedgerClassId: ledgerClassID,
				StatusTypeId:  int32(statusTypeID), //nolint:gosec // Controlled conversion
			}

			// Get optional fields from flags
			nextPmtDateStr, _ := cmd.Flags().GetString("next-pmt-date")
			if nextPmtDateStr != "" {
				nextPmtDate, err := helper.StrToDate(nextPmtDateStr)
				if err != nil {
					return fmt.Errorf("invalid --next-pmt-date: %w", err)
				}
				ledgerObj.NextPmtDate = helper.DaysSinceEpoch(nextPmtDate.UTC())
			}

			nextPmtAmt, _ := cmd.Flags().GetInt64("next-pmt-amt")
			if nextPmtAmt > 0 {
				ledgerObj.NextPmtAmt = nextPmtAmt
			}

			interestRate, _ := cmd.Flags().GetInt32("interest-rate")
			if interestRate > 0 {
				ledgerObj.InterestRate = interestRate
			}

			maturityDateStr, _ := cmd.Flags().GetString("maturity-date")
			if maturityDateStr != "" {
				maturityDate, err := helper.StrToDate(maturityDateStr)
				if err != nil {
					return fmt.Errorf("invalid --maturity-date: %w", err)
				}
				ledgerObj.MaturityDate = helper.DaysSinceEpoch(maturityDate.UTC())
			}

			// Get enum values from flags
			dayCountConvention, _ := cmd.Flags().GetString("day-count-convention")
			if dayCountConvention != "" {
				switch dayCountConvention {
				case "actual-365":
					ledgerObj.InterestDayCountConvention = ledger.DAY_COUNT_CONVENTION_ACTUAL_365
				case "actual-360":
					ledgerObj.InterestDayCountConvention = ledger.DAY_COUNT_CONVENTION_ACTUAL_360
				case "thirty-360":
					ledgerObj.InterestDayCountConvention = ledger.DAY_COUNT_CONVENTION_THIRTY_360
				case "actual-actual":
					ledgerObj.InterestDayCountConvention = ledger.DAY_COUNT_CONVENTION_ACTUAL_ACTUAL
				case "days-365":
					ledgerObj.InterestDayCountConvention = ledger.DAY_COUNT_CONVENTION_DAYS_365
				case "days-360":
					ledgerObj.InterestDayCountConvention = ledger.DAY_COUNT_CONVENTION_DAYS_360
				default:
					return fmt.Errorf("invalid --day-count-convention: %s", dayCountConvention)
				}
			}

			interestAccrualMethod, _ := cmd.Flags().GetString("interest-accrual-method")
			if interestAccrualMethod != "" {
				switch interestAccrualMethod {
				case "simple":
					ledgerObj.InterestAccrualMethod = ledger.INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST
				case "compound":
					ledgerObj.InterestAccrualMethod = ledger.INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST
				case "daily":
					ledgerObj.InterestAccrualMethod = ledger.INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING
				case "monthly":
					ledgerObj.InterestAccrualMethod = ledger.INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING
				case "quarterly":
					ledgerObj.InterestAccrualMethod = ledger.INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING
				case "annual":
					ledgerObj.InterestAccrualMethod = ledger.INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING
				case "continuous":
					ledgerObj.InterestAccrualMethod = ledger.INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING
				default:
					return fmt.Errorf("invalid --interest-accrual-method: %s", interestAccrualMethod)
				}
			}

			paymentFrequency, _ := cmd.Flags().GetString("payment-frequency")
			if paymentFrequency != "" {
				switch paymentFrequency {
				case "daily":
					ledgerObj.PaymentFrequency = ledger.PAYMENT_FREQUENCY_DAILY
				case "weekly":
					ledgerObj.PaymentFrequency = ledger.PAYMENT_FREQUENCY_WEEKLY
				case "monthly":
					ledgerObj.PaymentFrequency = ledger.PAYMENT_FREQUENCY_MONTHLY
				case "quarterly":
					ledgerObj.PaymentFrequency = ledger.PAYMENT_FREQUENCY_QUARTERLY
				case "annually":
					ledgerObj.PaymentFrequency = ledger.PAYMENT_FREQUENCY_ANNUALLY
				default:
					return fmt.Errorf("invalid --payment-frequency: %s", paymentFrequency)
				}
			}

			msg := &ledger.MsgCreateRequest{
				Ledger:    ledgerObj,
				Authority: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add custom flags
	ledgerFlags := pflag.NewFlagSet("ledger", pflag.ContinueOnError)
	ledgerFlags.String("next-pmt-date", "", "Next payment date (YYYY-MM-DD)")
	ledgerFlags.Int64("next-pmt-amt", 0, "Next payment amount")
	ledgerFlags.Int("interest-rate", 0, "Interest rate (10000000 = 10.000000%)")
	ledgerFlags.String("maturity-date", "", "Maturity date (YYYY-MM-DD)")
	ledgerFlags.String("day-count-convention", "", "Day count convention (actual-365, actual-360, thirty-360, actual-actual, days-365, days-360)")
	ledgerFlags.String("interest-accrual-method", "", "Interest accrual method (simple, compound, daily, monthly, quarterly, annual, continuous)")
	ledgerFlags.String("payment-frequency", "", "Payment frequency (daily, weekly, monthly, quarterly, annually)")
	cmd.Flags().AddFlagSet(ledgerFlags)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdDestroy() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destroy <asset_class_id> <nft_id>",
		Aliases: []string{},
		Short:   "Destroy a ledger by asset_class_id and nft_id",
		Example: `$ provenanced tx ledger destroy pb1a2b3c4... --from mykey`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			assetClassID := args[0]
			nftID := args[1]

			msg := &ledger.MsgDestroyRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassID,
					NftId:        nftID,
				},
				Authority: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAppend creates a new ledger entry from a JSON file
func CmdAppend() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "append <asset_class_id> <nft_id> <json_file_path>",
		Aliases: []string{"aj"},
		Short:   "Append ledger entries from a JSON file",
		Example: `$ provenanced tx ledger append asset-class-1 nft-1 entries.json --from mykey
where the json is formatted as follows:
[
	{
		"correlation_id": "entry1",
		"reverses_correlation_id": "",
		"is_void": false,
		"sequence": 1,
		"entry_type_id": 1,
		"posted_date": 19665,
		"effective_date": 19665,
		"total_amt": "80000",
		"applied_amounts": [
			{
				"bucket_type_id": 1,
				"applied_amt": "80000"
			}
		],
		"balance_amounts": [
			{
				"bucket_type_id": 1,
				"balance_amt": "80000"
			}
		]
	}
]
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			assetClassID := args[0]
			nftID := args[1]

			jsonData, err := os.ReadFile(args[2])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var rawEntries []json.RawMessage
			if err := json.Unmarshal(jsonData, &rawEntries); err != nil {
				return fmt.Errorf("failed to unmarshal JSON array: %w", err)
			}

			entries := make([]*ledger.LedgerEntry, 0, len(rawEntries))
			for _, rawEntry := range rawEntries {
				var entry ledger.LedgerEntry
				if err := clientCtx.Codec.UnmarshalJSON(rawEntry, &entry); err != nil {
					return err
				}
				entries = append(entries, &entry)
			}

			msg := &ledger.MsgAppendRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassID,
					NftId:        nftID,
				},
				Entries:   entries,
				Authority: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateLedgerClass creates a new ledger class
func CmdCreateLedgerClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-class <ledger_class_id> <asset_class_id> <denom>",
		Aliases: []string{"cc"},
		Short:   "Create a new ledger class",
		Example: `$ provenanced tx ledger create-class usd pb1a2b3c4... usd pb1a2b3c4... --from mykey`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassID := args[0]
			assetClassID := args[1]
			denom := args[2]

			msg := &ledger.MsgCreateLedgerClassRequest{
				LedgerClass: &ledger.LedgerClass{
					LedgerClassId:     ledgerClassID,
					AssetClassId:      assetClassID,
					Denom:             denom,
					MaintainerAddress: clientCtx.FromAddress.String(),
				},
				Authority: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAddLedgerClassStatusType adds a new status type to a ledger class
func CmdAddLedgerClassStatusType() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-status-type <ledger_class_id> <id> <code> <description>",
		Aliases: []string{"ast"},
		Short:   "Add a new status type to a ledger class",
		Example: `$ provenanced tx ledger add-status-type ledger-class-1 1 IN_REPAYMENT "In Repayment" --from mykey`,
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassID := args[0]
			id, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid <id>: %w", err)
			}
			code := args[2]
			description := args[3]

			msg := &ledger.MsgAddLedgerClassStatusTypeRequest{
				LedgerClassId: ledgerClassID,
				StatusType: &ledger.LedgerClassStatusType{
					Id:          int32(id), //nolint:gosec // Controlled conversion
					Code:        code,
					Description: description,
				},
				Authority: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAddLedgerClassEntryType adds a new entry type to a ledger class
func CmdAddLedgerClassEntryType() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-entry-type <ledger_class_id> <id> <code> <description>",
		Aliases: []string{"aet"},
		Short:   "Add a new entry type to a ledger class",
		Example: `$ provenanced tx ledger add-entry-type ledger-class-1 1 DISBURSEMENT "Disbursement" --from mykey`,
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassID := args[0]
			id, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid <id>: %w", err)
			}
			code := args[2]
			description := args[3]

			msg := &ledger.MsgAddLedgerClassEntryTypeRequest{
				LedgerClassId: ledgerClassID,
				EntryType: &ledger.LedgerClassEntryType{
					Id:          int32(id), //nolint:gosec // Controlled conversion
					Code:        code,
					Description: description,
				},
				Authority: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAddLedgerClassBucketType adds a new bucket type to a ledger class
func CmdAddLedgerClassBucketType() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-bucket-type <ledger_class_id> <id> <code> <description>",
		Aliases: []string{"abt"},
		Short:   "Add a new bucket type to a ledger class",
		Example: `$ provenanced tx ledger add-bucket-type ledger-class-1 1 PRINCIPAL "Principal" --from mykey`,
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ledgerClassID := args[0]
			id, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid <id>: %w", err)
			}
			code := args[2]
			description := args[3]

			msg := &ledger.MsgAddLedgerClassBucketTypeRequest{
				LedgerClassId: ledgerClassID,
				BucketType: &ledger.LedgerClassBucketType{
					Id:          int32(id), //nolint:gosec // Controlled conversion
					Code:        code,
					Description: description,
				},
				Authority: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTransferFundsWithSettlement returns the command for transferring funds with settlement instructions
func CmdTransferFundsWithSettlement() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xfer <fund_transfers_json_file",
		Short: "Submit fund transfers",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jsonData, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var rawTransfers []json.RawMessage
			if err := json.Unmarshal(jsonData, &rawTransfers); err != nil {
				return fmt.Errorf("failed to unmarshal JSON array: %w", err)
			}

			transfers := make([]*ledger.FundTransferWithSettlement, 0, len(rawTransfers))
			for _, rawTransfer := range rawTransfers {
				var transfer ledger.FundTransferWithSettlement
				if err := clientCtx.Codec.UnmarshalJSON(rawTransfer, &transfer); err != nil {
					return err
				}
				transfers = append(transfers, &transfer)
			}

			msg := &ledger.MsgTransferFundsWithSettlementRequest{
				Authority: clientCtx.FromAddress.String(),
				Transfers: transfers,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdBulkCreate returns the command for creating ledgers and entries in bulk
func CmdBulkCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-create <ledger_entries_json_file",
		Short: "Create ledgers and entries in bulk",
		Example: `$ provenanced tx ledger bulk-create data.json --from mykey`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jsonData, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var rawEntries []json.RawMessage
			if err := json.Unmarshal(jsonData, &rawEntries); err != nil {
				return fmt.Errorf("failed to unmarshal JSON array: %w", err)
			}

			ledgerToEntries := make([]*ledger.LedgerToEntries, 0, len(rawEntries))
			for _, rawEntry := range rawEntries {
				var entry ledger.LedgerToEntries
				if err := clientCtx.Codec.UnmarshalJSON(rawEntry, &entry); err != nil {
					return err
				}
				ledgerToEntries = append(ledgerToEntries, &entry)
			}

			msg := &ledger.MsgBulkCreateRequest{
				Authority: clientCtx.FromAddress.String(),
				LedgerToEntries: ledgerToEntries,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}