package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/expiration/types"
)

const (
	FlagSigners = "signers"
)

// NewTxCmd is the top-level command for expiration CLI transactions
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"exp"},
		Short:                      "Transaction commands for the expiration module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		AddExpirationCmd(),
		ExtendExpirationCmd(),
		DeleteExpirationCmd(),
		InvokeExpirationCmd(),
	)
	return txCmd
}

// AddExpirationCmd creates a command for adding an expiration
func AddExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [path/to/expiration.json]",
		Aliases: []string{"a"},
		Short:   "Create expiration metadata for an asset on the provenance blockchain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add module asset expiration on the provenance blockchain.
They should be defined in a JSON file.

Example:
$ %s tx expiration add path/to/expiration.json

Where expiration.json contains:

{
  "module_asset_id": "cosmos1...",
  "owner": "cosmos1...",
  "block_height": 1000000,
  "deposit": "10000nhash",
  "message": { // a proto-JSON-encoded sdk.Msg
  	"@type": "/provenance.metadata.v1.MsgDeleteScopeRequest",
  	"scope_id": "scope1...",
  	"signers": ["cosmos1..."]
  }
}
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			moduleAssetID, owner, height, deposit, msgs, err := parseAddExtendRequest(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			expiration := types.Expiration{
				ModuleAssetId: moduleAssetID,
				Owner:         owner,
				BlockHeight:   height,
				Deposit:       deposit[0],
				Message:       msgs,
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddExpirationRequest(expiration, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// ExtendExpirationCmd creates a command for extending an expiration
func ExtendExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extend [path/to/expiration.json]",
		Aliases: []string{"e"},
		Short:   "Extend/update expiration metadata for an asset on the provenance blockchain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Extend module asset expiration on the provenance blockchain.
They should be defined in a JSON file.

Example:
$ %s tx expiration extend path/to/expiration.json

Where expiration.json contains:

{
  "module_asset_id": "cosmos1...",
  "owner": "cosmos1...",
  "block_height": 1000000,
  "deposit": "10000nhash",
  "messages": { // a proto-JSON-encoded sdk.Msg
    "@type": "/provenance.metadata.v1.MsgDeleteScopeRequest",
    "scope_id": "scope1...",
  	"signers": ["cosmos1..."]
  }
}
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			moduleAssetID, owner, height, deposit, msgs, err := parseAddExtendRequest(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			expiration := types.Expiration{
				ModuleAssetId: moduleAssetID,
				Owner:         owner,
				BlockHeight:   height,
				Deposit:       deposit[0],
				Message:       msgs,
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			msg := types.NewMsgExtendExpirationRequest(expiration, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// DeleteExpirationCmd creates a command for deleting an expiration
func DeleteExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [module-asset-id]",
		Aliases: []string{"d"},
		Short:   "Delete expiration metadata for an asset on the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx expiration delete pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			moduleAssetID := args[0]
			msg := types.NewMsgDeleteExpirationRequest(moduleAssetID, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// InvokeExpirationCmd creates a command for invoking expiration logic on an asset
func InvokeExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "invoke [module-asset-id]",
		Aliases: []string{"i"},
		Short:   "Invoke expiration logic for an asset on the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx expiration invoke pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			moduleAssetID := args[0]
			msg := types.NewMsgInvokeExpirationRequest(moduleAssetID, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

type expiration struct {
	ModuleAssetID string          `json:"module_asset_id"`
	Owner         string          `json:"owner"`
	BlockHeight   int64           `json:"block_height"`
	Deposit       string          `json:"deposit"`
	Message       json.RawMessage `json:"message,omitempty"`
}

func parseAddExtendRequest(
	cdc codec.Codec,
	path string,
) (string, string, int64, sdk.Coins, *types2.Any, error) {
	var expiration expiration

	contents, err := os.ReadFile(path)
	if err != nil {
		return "", "", -1, nil, nil, err
	}

	if err := json.Unmarshal(contents, &expiration); err != nil {
		return "", "", -1, nil, nil, err
	}

	deposit, err := sdk.ParseCoinsNormalized(expiration.Deposit)
	if err != nil {
		return "", "", -1, nil, nil, err
	}

	var msg sdk.Msg
	if err := cdc.UnmarshalInterfaceJSON(expiration.Message, &msg); err != nil {
		return "", "", -1, nil, nil, err
	}
	if err := msg.ValidateBasic(); err != nil {
		return "", "", -1, nil, nil, err
	}

	anyMsg, err := types2.NewAnyWithValue(msg)
	if err != nil {
		return "", "", -1, nil, nil, err
	}

	return expiration.ModuleAssetID,
		expiration.Owner,
		expiration.BlockHeight,
		deposit,
		anyMsg, err
}

func addSignerFlagCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagSigners, "", "comma delimited list of bech32 addresses")
}

// parseSigners checks signers flag for signers, else uses the from address
func parseSigners(cmd *cobra.Command, client *client.Context) ([]string, error) {
	flagSet := cmd.Flags()
	if flagSet.Changed(FlagSigners) {
		signerList, _ := flagSet.GetString(FlagSigners)
		signers := strings.Split(signerList, ",")
		for _, signer := range signers {
			_, err := sdk.AccAddressFromBech32(signer)
			if err != nil {
				fmt.Printf("signer address must be a Bech32 string: %v", err)
				return nil, err
			}
		}
		return signers, nil
	}
	return []string{client.GetFromAddress().String()}, nil
}
