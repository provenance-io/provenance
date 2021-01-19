package cli

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// NewTxCmd is the top-level command for attribute CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"am"},
		Short:                      "Transaction commands for the attribute module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		NewAddAccountAttributeCmd(),
		NewDeleteAccountAttributeCmd(),
	)
	return txCmd
}

//  NewAddAccountAttributeCmd creates a command for adding an account attributes.
func NewAddAccountAttributeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [name] [address] [type] [value]",
		Short: "Add an account attribute to the provenance blockchain",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			account, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("account address must be a Bech32 string: %w", err)
			}
			attributeType, err := types.AttributeTypeFromString(strings.TrimSpace(args[2]))
			if err != nil {
				return fmt.Errorf("account attribute type is invalid: %w", err)
			}
			valueString := strings.TrimSpace(args[3])
			var value []byte
			if attributeType == types.AttributeType_Bytes {
				var err error
				if value, err = base64.StdEncoding.DecodeString(valueString); err != nil {
					return err
				}
			} else {
				value = []byte(valueString)
			}

			msg := types.NewMsgAddAttributeRequest(
				account,
				clientCtx.GetFromAddress(),
				name,
				attributeType,
				value,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewDeleteAccountAttributeCmd creates a command for removing account attributes.
func NewDeleteAccountAttributeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [name] [address]",
		Short: "Delete an account attribute from the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			account, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("account address must be a Bech32 string: %w", err)
			}
			msg := types.NewMsgDeleteAttributeRequest(
				account,
				clientCtx.GetFromAddress(),
				args[0],
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
