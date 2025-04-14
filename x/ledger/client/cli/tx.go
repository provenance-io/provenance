package cli

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/provenance-io/provenance/x/ledger"
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
	)

	return cmd
}

func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <nft_address> <denom> [next_pmt_date] [next_pmt_amt] [status] [interest_rate] [maturity_date]",
		Aliases: []string{},
		Short:   "Create a ledger for the nft_address",
		Example: `$ provenanced tx ledger create pb1a2b3c4... usd 2024-12-31 1000.00 IN_REPAYMENT 0.05 2025-12-31 --from mykey
$ provenanced tx ledger create pb1a2b3c4... usd --from mykey  # minimal example with required fields only`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			nftAddress := args[0]
			denom := args[1]

			// Create the ledger with required fields
			ledgerObj := &ledger.Ledger{
				NftAddress: nftAddress,
				Denom:      denom,
			}

			// Add optional fields if provided
			if len(args) > 2 {
				ledgerObj.NextPmtDate = args[2]
			}
			if len(args) > 3 {
				ledgerObj.NextPmtAmt = args[3]
			}
			if len(args) > 4 {
				ledgerObj.Status = args[4]
			}
			if len(args) > 5 {
				ledgerObj.InterestRate = args[5]
			}
			if len(args) > 6 {
				ledgerObj.MaturityDate = args[6]
			}

			msg := &ledger.MsgCreateRequest{
				Ledger: ledgerObj,
				Owner:  clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAppend creates a new ledger entry for a given nft
func CmdAppend() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "append <nft_address> <correlation_id> <sequence> <type> <posted_date> <effective_date> <amount> <prin_applied_amt> <prin_balance_amt>  <int_applied_amt> <int_balance_amt>  <other_applied_amt> <other_balance_amt>",
		Aliases: []string{},
		Short:   "Append an entry to an existing ledger",
		Example: `$ provenanced tx ledger append pb1a2b3c4... txn123 1 SCHEDULED_PAYMENT 2024-04-15 2024-04-15 1000.00 800.00 9200.00 200.00 400.00 0.00 0.00 --from mykey
$ provenanced tx ledger append pb1a2b3c4... txn124 2 FEE 2024-04-16 2024-04-16 50.00 0.00 9200.00 0.00 400.00 50.00 50.00 --from mykey`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 13 {
				return fmt.Errorf("missing arguments")
			}

			if len(args) > 13 {
				return fmt.Errorf("to many arguments")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			nftAddress := args[0]
			if nftAddress == "" {
				return fmt.Errorf("invalid <nft_address>")
			}

			correlation_id := args[1]
			if correlation_id == "" {
				return fmt.Errorf("invalid <correlation_id>")
			}

			sequence, err := parseUint32(args[2])
			if err != nil {
				return fmt.Errorf("invalid <sequence>: %v", err)
			}

			amt, ok := sdkmath.NewIntFromString(args[6])
			if !ok {
				return fmt.Errorf("invalid <amount>")
			}

			prinAppliedAmt, ok := sdkmath.NewIntFromString(args[7])
			if !ok {
				return fmt.Errorf("invalid <prin_applied_amt>")
			}

			prinBalAmt, ok := sdkmath.NewIntFromString(args[8])
			if !ok {
				return fmt.Errorf("invalid <prin_bal_amt>")
			}

			intAppliedAmt, ok := sdkmath.NewIntFromString(args[9])
			if !ok {
				return fmt.Errorf("invalid <int_applied_amt>")
			}

			intBalAmt, ok := sdkmath.NewIntFromString(args[10])
			if !ok {
				return fmt.Errorf("invalid <int_bal_amt>")
			}

			otherAppliedAmt, ok := sdkmath.NewIntFromString(args[11])
			if !ok {
				return fmt.Errorf("invalid <other_applied_amt>")
			}

			otherBalAmt, ok := sdkmath.NewIntFromString(args[12])
			if !ok {
				return fmt.Errorf("invalid <other_bal_amt>")
			}

			postedDate, err := time.Parse("2006-01-02", args[4])
			if err != nil {
				return fmt.Errorf("invalid <posted_date>: %v", err)
			}

			effectiveDate, err := time.Parse("2006-01-02", args[5])
			if err != nil {
				return fmt.Errorf("invalid <effective_date>: %v", err)
			}

			entryType, ok := ledger.LedgerEntryType_value[args[3]]
			if !ok {
				return fmt.Errorf("invalid <type>")
			}

			m := ledger.MsgAppendRequest{
				NftAddress: nftAddress,
				Entry: &ledger.LedgerEntry{
					CorrelationId:   correlation_id,
					Sequence:        sequence,
					Type:            ledger.LedgerEntryType(entryType),
					PostedDate:      postedDate.Format("2006-01-02"),
					EffectiveDate:   effectiveDate.Format("2006-01-02"),
					Amt:             amt,
					PrinAppliedAmt:  prinAppliedAmt,
					PrinBalAmt:      prinBalAmt,
					IntAppliedAmt:   intAppliedAmt,
					IntBalAmt:       intBalAmt,
					OtherAppliedAmt: otherAppliedAmt,
					OtherBalAmt:     otherBalAmt,
				},
				Owner: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &m)
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
