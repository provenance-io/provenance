package cli

import (
	"fmt"

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
		Use:     "create <nft_address> <denom> <",
		Aliases: []string{},
		Short:   "Create a ledger for the nft_address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			nftAddress := args[0]
			denom := args[1]

			m := ledger.MsgCreateRequest{
				NftAddress: nftAddress,
				Denom:      denom,
				Owner:      clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &m)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAppend creates a new ledger entry for a given nft
func CmdAppend() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "append <nft_address> <uuid> <type> <posted_date> <effective_date> <amount> <prin_applied_amt> <prin_balance_amt>  <int_applied_amt> <int_balance_amt>  <other_applied_amt> <other_balance_amt>",
		Aliases: []string{},
		Short:   "Append an entry to an existing ledger",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 12 {
				return fmt.Errorf("missing arguments")
			}

			if len(args) > 12 {
				return fmt.Errorf("to many arguments")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amt, ok := sdkmath.NewIntFromString(args[5])
			if !ok {
				return fmt.Errorf("Invalid <amount>: %s", err.Error())
			}

			prinAppliedAmt, ok := sdkmath.NewIntFromString(args[6])
			if !ok {
				return fmt.Errorf("Invalid <prin_applied_amt>: %s", err.Error())
			}

			prinBalAmt, ok := sdkmath.NewIntFromString(args[7])
			if !ok {
				return fmt.Errorf("Invalid <prin_bal_amt>: %s", err.Error())
			}

			intAppliedAmt, ok := sdkmath.NewIntFromString(args[8])
			if !ok {
				return fmt.Errorf("Invalid <int_applied_amt>: %s", err.Error())
			}

			intBalAmt, ok := sdkmath.NewIntFromString(args[9])
			if !ok {
				return fmt.Errorf("Invalid <int_bal_amt>: %s", err.Error())
			}

			otherAppliedAmt, ok := sdkmath.NewIntFromString(args[10])
			if !ok {
				return fmt.Errorf("Invalid <other_applied_amt>: %s", err.Error())
			}

			otherBalAmt, ok := sdkmath.NewIntFromString(args[11])
			if !ok {
				return fmt.Errorf("Invalid <other_bal_amt>: %s", err.Error())
			}

			m := ledger.MsgAppendRequest{
				NftAddress: args[0],
				Entry: &ledger.LedgerEntry{
					Uuid: args[1],
					// Type:            args[2],
					// PostedDate:      args[3],
					// EffectiveDate:   args[4],
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

// name := args[0]
// account := args[1]

// err = types.ValidateAttributeAddress(account)
// if err != nil {
// 	return fmt.Errorf("invalid address: %w", err)
// }
// attributeType, err := types.AttributeTypeFromString(strings.TrimSpace(args[2]))
// if err != nil {
// 	return fmt.Errorf("account attribute type is invalid: %w", err)
// }
// valueString := strings.TrimSpace(args[3])
// value, err := encodeAttributeValue(valueString, attributeType)
// if err != nil {
// 	return fmt.Errorf("error encoding value %s to type %s : %w", valueString, attributeType.String(), err)
// }

// msg := types.NewMsgAddAttributeRequest(
// 	account,
// 	clientCtx.GetFromAddress(),
// 	name,
// 	attributeType,
// 	value,
// )

// if len(args) == 5 {
// 	expireTime, err := time.Parse(time.RFC3339, args[4])
// 	if err != nil {
// 		return fmt.Errorf("unable to parse time %q required format is RFC3339 (%v): %w", args[4], time.RFC3339, err)
// 	}
// 	msg.ExpirationDate = &expireTime
// }

// return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
