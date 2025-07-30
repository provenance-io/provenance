package cli

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

// NewTxCmd returns a root CLI command handler for certain modules
// transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      types.ModuleName + " subcommands.",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdAddFido2Credentials(),
		MsgUpdateParams(),
		GetCmdDeleteCredential(),
		GetCmdRegisterCosmosCredential(),
	)
	return txCmd
}

const (
	FlagSender         = "sender"
	FlagAttestation    = "encodedAttestation" //gosec:G101
	FlagUserIdentifier = "user-identifier"
)

// AddWebAuthnCredentialFlags adds the flags needed when registering a WebAuthn credential.
func AddWebAuthnCredentialFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagSender, "", "the address of the sender of this message")
	cmd.Flags().String(FlagAttestation, "", "the encoded attestation")
	cmd.Flags().String(FlagUserIdentifier, "", "the user identifier")
}

// GetCmdAddFido2Credentials returns a CLI command handler for creating a MsgAddCredentials transaction.
func GetCmdAddFido2Credentials() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-webauthn-credentials",
		Short: "Add WebAuthn credentials to a smart account",
		Long: strings.TrimSpace(`Add WebAuthn credentials to a smart account.
This command allows you to register WebAuthn credentials for a smart account.
You need to provide the sender address, encoded attestation, and user identifier.
This assumes a base account being present already.`),
		Example: strings.TrimSpace(`$ appd tx smartaccounts add-credentials --sender=<sender_address> --encodedAttestation=<encoded_attestation> --user-identifier=<user_identifier>`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			senderAddress := clientCtx.GetFromAddress().String()

			encodedAttestation, err := cmd.Flags().GetString(FlagAttestation)
			if err != nil {
				return err
			}

			userIdentifier, err := cmd.Flags().GetString(FlagUserIdentifier)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterFido2Credential(senderAddress, encodedAttestation, userIdentifier)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	AddWebAuthnCredentialFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// MsgUpdateParams returns a CLI command handler for creating a MsgUpdateParams transaction.
func MsgUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [enable] [max-credential-allowed]",
		Short: "Update the smart account module params via governance proposal",
		Long:  "Submit a governance proposal to update smart account params. Both <enable> and <max-credential-allowed> are required.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			flagSet := cmd.Flags()
			authority := provcli.GetAuthority(flagSet)

			enabled, err := strconv.ParseBool(args[0])
			if err != nil {
				return err
			}

			maxCreds32, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return err
			}

			// gosec is complaining but this should ensure strconv.ParseUint(args[1], 10, 32) no overflow
			msg := &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					Enabled:              enabled,
					MaxCredentialAllowed: uint32(maxCreds32), //nolint:gosec // disable G115
				},
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return provcli.GenerateOrBroadcastTxCLIAsGovProp(cliCtx, flagSet, msg)
		},
	}

	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDeleteCredential returns a CLI command handler for deleting a credential from a smart account.
func GetCmdDeleteCredential() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-credential [credential-number]",
		Short: "Delete a credential from a smart account",
		Long: strings.TrimSpace(`Delete a credential from a smart account.
This command allows you to remove a credential by its number from a smart account.
The sender must be the owner of the smart account.`),
		Args:    cobra.ExactArgs(1),
		Example: strings.TrimSpace(`$ provenanced tx smartaccounts delete-credential 1 --from=<account_address>`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			senderAddress := clientCtx.GetFromAddress().String()

			credentialNumber, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteCredential(senderAddress, credentialNumber)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRegisterCosmosCredential returns a CLI command handler for registering a Cosmos credential.
func GetCmdRegisterCosmosCredential() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-cosmos-credential [base64-pubkey]",
		Short: "Register a Cosmos credential for a smart account",
		Long: strings.TrimSpace(`Register a Cosmos credential for a smart account.
This command allows you to register a secp256k1 public key as a credential.
You need to provide the base64 encoded public key directly as an argument.`),
		Args:    cobra.ExactArgs(1),
		Example: strings.TrimSpace(`$ appd tx smartaccounts register-cosmos-credential <base64-encoded-pubkey> --from=<account_address>`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			senderAddress := clientCtx.GetFromAddress().String()

			// Decode the base64 public key
			pubKeyBytes, err := base64.StdEncoding.DecodeString(args[0])
			if err != nil {
				// Try with RawURLEncoding if standard encoding fails
				pubKeyBytes, err = base64.RawURLEncoding.DecodeString(args[0])
				if err != nil {
					return err
				}
			}

			// Create a secp256k1 public key
			pubKey := &secp256k1.PubKey{
				Key: pubKeyBytes,
			}

			// Wrap the public key in an Any type
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterCosmosCredential(senderAddress, pubKeyAny)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
