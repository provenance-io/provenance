package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
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
		ExtendExpirationCmd(),
		InvokeExpirationCmd(),
	)
	return txCmd
}

var txCmdStr = fmt.Sprintf(`%s tx expiration`, version.AppName)

// ExtendExpirationCmd creates a command for extending an expiration
func ExtendExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extend [module-asset-id] [duration n{y|w|d|h}]",
		Aliases: []string{"e"},
		Short:   "Extend expiration metadata for an asset on the provenance blockchain",
		Long: fmt.Sprintf(`Extend expiration metadata for an asset on the provenance blockchain

Example:
$ %s extend pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk duration 1y

module-asset-id   - module asset address
duration          - the duration period for which the module asset will continue to remain on chain.
                    Valid time units are "y", "w", "d", "h"`, txCmdStr),
		Example: fmt.Sprintf(`$ %s extend pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk duration 1y`, txCmdStr),
		Args:    cobra.ExactArgs(2),
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
			duration := args[1]
			msg := types.NewMsgExtendExpirationRequest(moduleAssetID, duration, signers)
			if err = msg.ValidateBasic(); err != nil {
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
		Example: fmt.Sprintf(`$ %s invoke pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk`, txCmdStr),
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
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
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
