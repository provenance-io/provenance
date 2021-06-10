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
		NewUpdateAccountAttributeCmd(),
		NewDeleteDistinctAccountAttributeCmd(),
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
			value, err := encodeAttributeValue(valueString, attributeType)
			if err != nil {
				return fmt.Errorf("error encoding value %s to type %s : %v", valueString, attributeType.String(), err)

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

//  NewUpdateAccountAttributeCmd creates a command for adding an account attributes.
func NewUpdateAccountAttributeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [name] [address] [original-type] [original-value] [update-type] [update-value]",
		Short: "Update an account attribute on the provenance blockchain",
		Args:  cobra.ExactArgs(6),
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
			origAttributeType, err := types.AttributeTypeFromString(strings.TrimSpace(args[2]))
			if err != nil {
				return fmt.Errorf("account attribute type is invalid: %w", err)
			}
			updateAttributeType, err := types.AttributeTypeFromString(strings.TrimSpace(args[4]))
			if err != nil {
				return fmt.Errorf("account attribute type is invalid: %w", err)
			}
			origValArg := strings.TrimSpace(args[3])
			origValue, err := encodeAttributeValue(origValArg, origAttributeType)
			if err != nil {
				return fmt.Errorf("error encoding value %s to type %s : %v", origValArg, origAttributeType.String(), err)
			}
			updateValArg := strings.TrimSpace(args[5])
			updateValue, err := encodeAttributeValue(updateValArg, updateAttributeType)
			if err != nil {
				return fmt.Errorf("error encoding value %s to type %s : %v", updateValArg, updateAttributeType.String(), err)
			}

			msg := types.NewMsgUpdateAttributeRequest(
				account,
				clientCtx.GetFromAddress(),
				name,
				origValue,
				updateValue,
				origAttributeType,
				updateAttributeType,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func encodeAttributeValue(value string, attrType types.AttributeType) ([]byte, error) {
	var encodedValue []byte
	if attrType == types.AttributeType_Bytes {
		var err error
		if encodedValue, err = base64.StdEncoding.DecodeString(value); err != nil {
			return nil, err
		}
	} else {
		encodedValue = []byte(value)
	}
	return encodedValue, nil
}

// NewDeleteDistinctAccountAttributeCmd creates a command for removing account attributes with specific name value.
func NewDeleteDistinctAccountAttributeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [name] [address] [type] [value]",
		Short: "Delete an account attribute with specific name and value the provenance blockchain",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			account, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("account address must be a Bech32 string: %w", err)
			}
			attributeType, err := types.AttributeTypeFromString(strings.TrimSpace(args[2]))
			if err != nil {
				return fmt.Errorf("account attribute type is invalid: %w", err)
			}
			deleteValue, err := encodeAttributeValue(strings.TrimSpace(args[3]), attributeType)
			if err != nil {
				return fmt.Errorf("error encoding value %s to type %s : %v", deleteValue, attributeType.String(), err)
			}
			msg := types.NewMsgDeleteDistinctAttributeRequest(account, clientCtx.GetFromAddress(), args[0], deleteValue)
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
