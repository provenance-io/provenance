package cli

import (
	"fmt"

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
			owner := "tp1c2hkxhv3e56hj5gku6nlyk86r5aty78wd9q47v"
			m := ledger.MsgCreateRequest{
				NftAddress: nftAddress,
				Denom:      denom,
				Owner:      owner,
			}

			fmt.Println(m)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &m)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAppend creates a new ledger entry for a given nft
func CmdAppend() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "append <entry>",
		Aliases: []string{},
		Short:   "Append an entry to an existing ledger",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			return fmt.Errorf("TODO")
		},
	}

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
