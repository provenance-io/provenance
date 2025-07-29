package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	chunkimport "github.com/provenance-io/provenance/x/ledger/client/cli/import"
	"github.com/provenance-io/provenance/x/ledger/helper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
	"github.com/spf13/cobra"
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
		CmdBulkImport(),
		chunkimport.CmdChunkedBulkImport(),
	)

	return cmd
}

func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		// TODO: Update optional fields to use flags instead of positional arguments
		Use:     "create <asset_class_id> <nft_id> <ledger_class_id> <status_type_id> [next_pmt_date] [next_pmt_amt] [interest_rate] [maturity_date]",
		Aliases: []string{},
		Short:   "Create a ledger for the nft_address",
		Example: `$ provenanced tx ledger create "asset-class-1" "nft-1" "ledger-class-1" "nhash" "IN_REPAYMENT" "2024-12-31" "1000.00" "0.05" "2025-12-31" --from mykey
$ provenanced tx ledger create "asset-class-1" "nft-1" "ledger-class-1" "nhash" "IN_REPAYMENT" --from mykey  # minimal example with required fields only`,
		Args: cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			assetClassId := args[0]
			nftId := args[1]
			ledgerClassId := args[2]

			statusTypeId, err := strconv.ParseInt(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid <status_type_id>: %v", err)
			}

			// Create the ledger with required fields
			ledgerObj := &ledger.Ledger{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
				},
				LedgerClassId: ledgerClassId,
				StatusTypeId:  int32(statusTypeId),
			}

			// Add optional fields if provided
			if len(args) > 4 {
				nextPmtDate, err := helper.StrToDate(args[4])
				if err != nil {
					return fmt.Errorf("invalid <next_pmt_date>: %v", err)
				}
				ledgerObj.NextPmtDate = helper.DaysSinceEpoch(nextPmtDate.UTC())

				ledgerObj.NextPmtAmt, err = strconv.ParseInt(args[5], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid <next_pmt_amt>: %v", err)
				}

				ledgerObj.StatusTypeId = int32(statusTypeId)

				intRate, err := strconv.ParseInt(args[6], 10, 32)
				if err != nil {
					return fmt.Errorf("invalid <interest_rate>: %v", err)
				}
				ledgerObj.InterestRate = int32(intRate)

				maturityDate, err := helper.StrToDate(args[7])
				if err != nil {
					return fmt.Errorf("invalid <maturity_date>: %v", err)
				}
				ledgerObj.MaturityDate = helper.DaysSinceEpoch(maturityDate.UTC())
			}

			msg := &ledger.MsgCreateRequest{
				Ledger:    ledgerObj,
				Authority: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

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

			assetClassId := args[0]
			nftId := args[1]

			// TODO: remove this check and let the server validate the nft_id
			if nftId == "" {
				return fmt.Errorf("invalid <nft_id>")
			}

			msg := &ledger.MsgDestroyRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
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
		Use:     "append <asset_class_id> <nft_id> <json>",
		Aliases: []string{"aj"},
		Short:   "Append ledger entries from a JSON file",
		Example: `$ provenanced tx ledger append pb1a2b3c4... 0ADE096F-60D8-49CF-8D20-418DABD549B1 'json' --from mykey
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
	},
	{
		"correlation_id": "entry2",
		"reverses_correlation_id": "",
		"is_void": false,
		"sequence": 1,
		"entry_type_id": 2,
		"posted_date": 19700,
		"effective_date": 19700,
		"total_amt": "1000",
		"applied_amounts": [
			{
				"bucket_type_id": 1,
				"applied_amt": "400"
			},
			{
				"bucket_type_id": 2,
				"applied_amt": "600"
			}
		],
		"balance_amounts": [
			{
				"bucket_type_id": 1,
				"balance_amt": "76600"
			},
			{
				"bucket_type_id": 2,
				"balance_amt": "0"
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

			assetClassId := args[0]
			nftId := args[1]

			// TODO: remove this check and let the server validate the nft_id
			if nftId == "" {
				return fmt.Errorf("invalid <nft_id>")
			}

			// TODO read this from a file and update the example
			jsonData := args[2]

			var entries []*ledger.LedgerEntry
			if err := json.Unmarshal([]byte(jsonData), &entries); err != nil {
				return fmt.Errorf("failed to unmarshal JSON: %w", err)
			}

			// Validate entries
			// TODO: move this to the server
			for _, entry := range entries {
				if entry.Sequence >= 100 {
					return fmt.Errorf("sequence must be less than 100")
				}
				if len(entry.CorrelationId) > 50 {
					return fmt.Errorf("correlation_id must be 50 characters or less")
				}
			}

			msg := &ledger.MsgAppendRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
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

			ledgerClassId := args[0]
			assetClassId := args[1]
			denom := args[2]

			msg := &ledger.MsgCreateLedgerClassRequest{
				LedgerClass: &ledger.LedgerClass{
					LedgerClassId:     ledgerClassId,
					AssetClassId:      assetClassId,
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

			ledgerClassId := args[0]
			id, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid <id>: %v", err)
			}
			code := args[2]
			description := args[3]

			msg := &ledger.MsgAddLedgerClassStatusTypeRequest{
				LedgerClassId: ledgerClassId,
				StatusType: &ledger.LedgerClassStatusType{
					Id:          int32(id),
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

			ledgerClassId := args[0]
			id, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid <id>: %v", err)
			}
			code := args[2]
			description := args[3]

			msg := &ledger.MsgAddLedgerClassEntryTypeRequest{
				LedgerClassId: ledgerClassId,
				EntryType: &ledger.LedgerClassEntryType{
					Id:          int32(id),
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

			ledgerClassId := args[0]
			id, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid <id>: %v", err)
			}
			code := args[2]
			description := args[3]

			msg := &ledger.MsgAddLedgerClassBucketTypeRequest{
				LedgerClassId: ledgerClassId,
				BucketType: &ledger.LedgerClassBucketType{
					Id:          int32(id),
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
		Use:   "xfer <transfers-json>",
		Short: "Submit a fund transfer with settlement instructions (ledger module)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// TODO: read this from a file and update the example
			jsonStr := args[0]

			var transfers []ledger.FundTransferWithSettlement
			if err := json.Unmarshal([]byte(jsonStr), &transfers); err != nil {
				return fmt.Errorf("failed to parse transfers JSON: %w", err)
			}

			msg := &ledger.MsgTransferFundsWithSettlementRequest{
				Authority: clientCtx.FromAddress.String(),
				Transfers: make([]*ledger.FundTransferWithSettlement, len(transfers)),
			}
			for i := range transfers {
				msg.Transfers[i] = &transfers[i]
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdBulkImport imports ledger data from a genesis state file
func CmdBulkImport() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bulk-import <genesis_state_file>",
		Aliases: []string{"bi"},
		Short:   "Bulk import ledger data from a genesis state file",
		Example: `$ provenanced tx ledger bulk-import genesis.json --from mykey`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			genesisStateFile := args[0]

			// Read the genesis state file
			genesisStateBytes, err := os.ReadFile(genesisStateFile)
			if err != nil {
				return fmt.Errorf("failed to read genesis state file: %w", err)
			}

			// Parse the genesis state
			var genesisState ledger.GenesisState
			if err := json.Unmarshal(genesisStateBytes, &genesisState); err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			msg := &ledger.MsgBulkImportRequest{
				Authority:    clientCtx.FromAddress.String(),
				GenesisState: &genesisState,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
