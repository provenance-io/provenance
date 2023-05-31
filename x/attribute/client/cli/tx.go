package cli

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

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
		NewSetAccountDataCmd(),
	)
	return txCmd
}

// NewAddAccountAttributeCmd creates a command for adding an account attributes.
func NewAddAccountAttributeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name> <address> <type> <value> [expire-time]",
		Aliases: []string{"a"},
		Short:   "Add an account attribute to the provenance blockchain",
		Long: fmt.Sprintf(`Note: the attribute name must have already been created through the name module.  
Refer to %s tx name bind --help for more information on how to do this.`, version.AppName),
		Args: cobra.RangeArgs(4, 5),
		Example: fmt.Sprintf(`$ %s tx attribute add "attr1.pb" tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx "string" "test value"
		$ %s tx attribute add "attr1.pb" tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx "string" "test value" 2050-01-15T00:00:00Z`, version.AppName, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			account := args[1]

			err = types.ValidateAttributeAddress(account)
			if err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}
			attributeType, err := types.AttributeTypeFromString(strings.TrimSpace(args[2]))
			if err != nil {
				return fmt.Errorf("account attribute type is invalid: %w", err)
			}
			valueString := strings.TrimSpace(args[3])
			value, err := encodeAttributeValue(valueString, attributeType)
			if err != nil {
				return fmt.Errorf("error encoding value %s to type %s : %w", valueString, attributeType.String(), err)
			}

			msg := types.NewMsgAddAttributeRequest(
				account,
				clientCtx.GetFromAddress(),
				name,
				attributeType,
				value,
			)

			if len(args) == 5 {
				expireTime, err := time.Parse(time.RFC3339, args[4])
				if err != nil {
					return fmt.Errorf("unable to parse time %q required format is RFC3339 (%v): %w", args[4], time.RFC3339, err)
				}
				msg.ExpirationDate = &expireTime
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUpdateAccountAttributeCmd creates a command for adding an account attributes.
func NewUpdateAccountAttributeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <name> <address> <original-type> <original-value> <update-type> <update-value>",
		Aliases: []string{"u"},
		Short:   "Update an account attribute on the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx attribute update "attr1.pb" tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx "string" "test value" "int" 100`, version.AppName),
		Args:    cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			account := args[1]

			err = types.ValidateAttributeAddress(account)
			if err != nil {
				return fmt.Errorf("invalid account address: %w", err)
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
				return fmt.Errorf("error encoding value %s to type %s : %w", origValArg, origAttributeType.String(), err)
			}
			updateValArg := strings.TrimSpace(args[5])
			updateValue, err := encodeAttributeValue(updateValArg, updateAttributeType)
			if err != nil {
				return fmt.Errorf("error encoding value %s to type %s : %w", updateValArg, updateAttributeType.String(), err)
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
	if attrType == types.AttributeType_Bytes || attrType == types.AttributeType_Proto {
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
		Use:     "delete-distinct <name> <address> <type> <value>",
		Aliases: []string{"dd"},
		Short:   "Delete an account attribute with specific name and value the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx attribute delete-distinct "attr1.pb" tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx "string" "test value"`, version.AppName),
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			err = types.ValidateAttributeAddress(args[1])
			if err != nil {
				return fmt.Errorf("invalid attribute address: %w", err)
			}
			attributeType, err := types.AttributeTypeFromString(strings.TrimSpace(args[2]))
			if err != nil {
				return fmt.Errorf("account attribute type is invalid: %w", err)
			}
			deleteValue, err := encodeAttributeValue(strings.TrimSpace(args[3]), attributeType)
			if err != nil {
				return fmt.Errorf("error encoding value %s to type %s : %w", deleteValue, attributeType.String(), err)
			}
			msg := types.NewMsgDeleteDistinctAttributeRequest(args[1], clientCtx.GetFromAddress(), args[0], deleteValue)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewDeleteAccountAttributeCmd creates a command for removing account attributes.
func NewDeleteAccountAttributeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name> <address>",
		Aliases: []string{"d"},
		Short:   "Delete an account attribute from the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx attribute delete "attr1.pb" tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			err = types.ValidateAttributeAddress(args[1])
			if err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}
			msg := types.NewMsgDeleteAttributeRequest(
				args[1],
				clientCtx.GetFromAddress(),
				args[0],
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUpdateAccountAttributeExpirationCmd creates a command for updating account attributes expirations
func NewUpdateAccountAttributeExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-expiration <name> <address> <value> [expiration-date]",
		Aliases: []string{"ue"},
		Short:   "Updates an attribute's expiration date on the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx attribute update-expiration "attr1.pb" tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx "attribute value" 2050-01-15T00:00:00Z
$ %s tx attribute update-expiration "attr1.pb" tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx "attribute value"`, version.AppName, version.AppName),
		Args: cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			err = types.ValidateAttributeAddress(args[1])
			if err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}
			var expireTime *time.Time
			if len(args) == 4 {
				parsedTime, err := time.Parse(time.RFC3339, args[3])
				if err != nil {
					return fmt.Errorf("unable to parse time %q required format is RFC3339 (%v): %w", args[3], time.RFC3339, err)
				}
				expireTime = &parsedTime
			}
			msg := types.NewMsgUpdateAttributeExpirationRequest(
				args[1],
				args[0],
				args[2],
				expireTime,
				clientCtx.GetFromAddress(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewSetAccountDataCmd creates a command for setting account data.
func NewSetAccountDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-data " + AccountDataFlagsUse,
		Aliases: []string{"accountdata", "ad"},
		Short:   "Set an account's data to either the value provided or the contents of the file provided",
		Example: fmt.Sprintf(`$ %[1]s tx attribute account-data --%s "This is some account data."
$ %[1]s tx attribute account-data --%s account-data.json
$ %[1]s tx attribute account-data --%s
`,
			version.AppName, FlagValue, FlagFile, FlagDelete),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSetAccountDataRequest{
				Account: clientCtx.GetFromAddress().String(),
			}

			msg.Value, err = ReadAccountDataFlags(cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	AddAccountDataFlagsToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
