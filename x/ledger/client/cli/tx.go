package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/provenance-io/provenance/x/ledger"
	"github.com/provenance-io/provenance/x/ledger/keeper"
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
	)

	return cmd
}

func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <asset_class_id> <nft_id> <ledger_class_id> <denom> [next_pmt_date] [next_pmt_amt] [status_type_id] [interest_rate] [maturity_date]",
		Aliases: []string{},
		Short:   "Create a ledger for the nft_address",
		Example: `$ provenanced tx ledger create pb1a2b3c4... 0ADE096F-60D8-49CF-8D20-418DABD549B1 usd 2024-12-31 1000.00 IN_REPAYMENT 0.05 2025-12-31 --from mykey
$ provenanced tx ledger create pb1a2b3c4... 0ADE096F-60D8-49CF-8D20-418DABD549B1 usd --from mykey  # minimal example with required fields only`,
		Args: cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			assetClassId := args[0]
			nftId := args[1]
			ledgerClassId := args[2]

			// Create the ledger with required fields
			ledgerObj := &ledger.Ledger{
				Key: &ledger.LedgerKey{
					AssetClassId: assetClassId,
					NftId:        nftId,
				},
				LedgerClassId: ledgerClassId,
			}

			// Add optional fields if provided
			if len(args) > 2 {
				nextPmtDate, err := keeper.StrToDate(args[2])
				if err != nil {
					return fmt.Errorf("invalid <next_pmt_date>: %v", err)
				}
				ledgerObj.NextPmtDate = keeper.DaysSinceEpoch(nextPmtDate.UTC())
			}
			if len(args) > 3 {
				ledgerObj.NextPmtAmt, err = strconv.ParseInt(args[3], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid <next_pmt_amt>: %v", err)
				}
			}
			if len(args) > 4 {
				statusTypeId, err := strconv.ParseInt(args[4], 10, 32)
				if err != nil {
					return fmt.Errorf("invalid <status_type_id>: %v", err)
				}

				ledgerObj.StatusTypeId = int32(statusTypeId)
			}
			if len(args) > 5 {
				intRate, err := strconv.ParseInt(args[5], 10, 32)
				if err != nil {
					return fmt.Errorf("invalid <interest_rate>: %v", err)
				}
				ledgerObj.InterestRate = int32(intRate)
			}
			if len(args) > 6 {
				maturityDate, err := keeper.StrToDate(args[6])
				if err != nil {
					return fmt.Errorf("invalid <maturity_date>: %v", err)
				}
				ledgerObj.MaturityDate = keeper.DaysSinceEpoch(maturityDate.UTC())
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

// CmdAppendJson creates a new ledger entry from a JSON file
func CmdAppend() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "append <asset_class_id> <nft_id> <json_file>",
		Aliases: []string{"aj"},
		Short:   "Append ledger entries from a JSON file",
		Example: `$ provenanced tx ledger append pb1a2b3c4... 0ADE096F-60D8-49CF-8D20-418DABD549B1 entries.json --from mykey
		where the json file is formatted as follows:
		[
			{
				"correlation_id": "entry1",
				"reverses_correlation_id": "",
				"is_void": false,
				"sequence": 1,
				"entry_type_id": 1,
				"posted_date": 19700,
				"effective_date": 19700,
				"total_amt": 1000,
				"applied_amounts": [
				{
					"bucket_type_id": 1,
					"applied_amt": 800
				},
				{
					"bucket_type_id": 2,
					"applied_amt": 200
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
			if nftId == "" {
				return fmt.Errorf("invalid <nft_id>")
			}

			jsonFile := args[1]
			jsonData, err := os.ReadFile(jsonFile)
			if err != nil {
				return fmt.Errorf("failed to read JSON file: %w", err)
			}

			var entries []*ledger.LedgerEntry
			if err := json.Unmarshal(jsonData, &entries); err != nil {
				return fmt.Errorf("failed to unmarshal JSON: %w", err)
			}

			// Validate entries
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

// parseUint32 parses a string into a uint32
func parseUint32(s string) (uint32, error) {
	val, err := sdkmath.ParseUint(s)
	if err != nil {
		return 0, err
	}
	if val.GT(sdkmath.NewUint(100)) {
		return 0, fmt.Errorf("sequence must be less than 100")
	}
	return uint32(val.Uint64()), nil
}
