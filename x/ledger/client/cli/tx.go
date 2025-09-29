package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"

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
			statusTypeID, err :=
				strconv.ParseInt(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid status_type_id: %w", err)
			}

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
			flagSet := cmd.Flags()

			nextPmtDateStr, err := ReadFlagNextPmtDate(flagSet)
			if err != nil {
				return fmt.Errorf("failed to read --%s: %w", FlagNextPmtDate, err)
			}
			if nextPmtDateStr != "" {
				nextPmtDate, err := helper.ParseYMD(nextPmtDateStr)
				if err != nil {
					return fmt.Errorf("invalid --next-pmt-date: %w", err)
				}
				ledgerObj.NextPmtDate = helper.DaysSinceEpoch(nextPmtDate.UTC())
			}

			nextPmtAmt, err := ReadFlagNextPmtAmt(flagSet)
			if err != nil {
				return fmt.Errorf("invalid --next-pmt-amt: %w", err)
			}
			if nextPmtAmt > 0 {
				ledgerObj.NextPmtAmt = math.NewInt(nextPmtAmt)
			}

			interestRate, err := ReadFlagInterestRate(flagSet)
			if err != nil {
				return fmt.Errorf("invalid --interest-rate: %w", err)
			}
			if interestRate > 0 {
				ledgerObj.InterestRate = interestRate
			}

			maturityDateStr, err := ReadFlagMaturityDate(flagSet)
			if err != nil {
				return fmt.Errorf("failed to read --%s: %w", FlagMaturityDate, err)
			}
			if maturityDateStr != "" {
				maturityDate, err := helper.ParseYMD(maturityDateStr)
				if err != nil {
					return fmt.Errorf("invalid --maturity-date: %w", err)
				}
				ledgerObj.MaturityDate = helper.DaysSinceEpoch(maturityDate.UTC())
			}

			// Get enum values from flags
			dayCountConvention, err := ReadFlagDayCountConvention(flagSet)
			if err != nil {
				return err
			}
			ledgerObj.InterestDayCountConvention = dayCountConvention

			interestAccrualMethod, err := ReadFlagInterestAccrualMethod(flagSet)
			if err != nil {
				return err
			}
			ledgerObj.InterestAccrualMethod = interestAccrualMethod

			paymentFrequency, err := ReadFlagPaymentFrequency(flagSet)
			if err != nil {
				return err
			}
			ledgerObj.PaymentFrequency = paymentFrequency

			msg := &ledger.MsgCreateLedgerRequest{
				Ledger: ledgerObj,
				Signer: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddFlagNextPmtDate(cmd)
	AddFlagNextPmtAmt(cmd)
	AddFlagInterestRate(cmd)
	AddFlagMaturityDate(cmd)
	AddFlagDayCountConvention(cmd)
	AddFlagInterestAccrualMethod(cmd)
	AddFlagPaymentFrequency(cmd)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdDestroy() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destroy <asset_class_id> <nft_id>",
		Aliases: []string{},
		Short:   "Destroy a ledger by asset_class_id and nft_id",
		Example: `$ provenanced tx ledger destroy asset-class-1 nft-1 --from mykey`,
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
				Signer: clientCtx.FromAddress.String(),
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
		Aliases: []string{},
		Short:   "Append ledger entries from a JSON file",
		Example: `$ provenanced tx ledger append asset-class-1 nft-1 entries.json --from mykey
where the json is formatted as follows (array ofLedgerEntry type):
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
		"applied_amounts": [{"bucket_type_id": 1, "applied_amt": "80000"}, ...],	
		"balance_amounts": [{"bucket_type_id": 1, "balance_amt": "80000"}, ...],
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
				Entries: entries,
				Signer:  clientCtx.FromAddress.String(),
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
		Example: `$ provenanced tx ledger create-class ledger-class-1 asset-class-1 usd --from mykey`,
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
				Signer: clientCtx.FromAddress.String(),
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
				Signer: clientCtx.FromAddress.String(),
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
				Signer: clientCtx.FromAddress.String(),
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
				Signer: clientCtx.FromAddress.String(),
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
		Example: `$ provenanced tx ledger xfer transfers.json --from mykey
		where the json is formatted as follows (array ofFundTransferWithSettlement type):
		[
			{
				"key": {"asset_class_id": "asset-class-1", "nft_id": "nft-1"},
				"ledger_entry_correlation_id": "entry1",
				"settlement_instructions": [
					{
						"amount": {"denom": "usd", "amount": "100000"},
						"recipient_address": "tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx",
						"status": "FUNDING_TRANSFER_STATUS_PENDING",
						"memo": "test transfer",
					}
					...
				]
			}
		]
		`,
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
				Signer:    clientCtx.FromAddress.String(),
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
		Use:   "bulk-create <ledger_entries_json_file>",
		Short: "Create ledgers and entries in bulk",
		Example: `$ provenanced tx ledger bulk-create data.json --from mykey
		where the json is formatted as follows (array of LedgerAndEntries type):
		[
			{
				"ledger_key": {"asset_class_id": "asset-class-1", "nft_id": "nft-1"},
				"ledger": {"ledger_class_id": "ledger-class-1", "status_type_id": 1, "next_pmt_date": 19665, ... },
				"entries": [
					{
						"correlation_id": "entry1",
						"reverses_correlation_id": "",
						"is_void": false,
						"sequence": 1,
						"entry_type_id": 1,
						"posted_date": 19665,
						"effective_date": 19665,
						"total_amt": "80000",
						"applied_amounts": [{"bucket_type_id": 1, "applied_amt": "80000"}, ...],	
						"balance_amounts": [{"bucket_type_id": 1, "balance_amt": "80000"}, ...],
					}
					...
				]
			}
		]
		`,
		Args: cobra.ExactArgs(1),
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

			ledgerAndEntries := make([]*ledger.LedgerAndEntries, 0, len(rawEntries))
			for _, rawEntry := range rawEntries {
				var entry ledger.LedgerAndEntries
				if err := clientCtx.Codec.UnmarshalJSON(rawEntry, &entry); err != nil {
					return err
				}
				ledgerAndEntries = append(ledgerAndEntries, &entry)
			}

			msg := &ledger.MsgBulkCreateRequest{
				Signer:           clientCtx.FromAddress.String(),
				LedgerAndEntries: ledgerAndEntries,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
