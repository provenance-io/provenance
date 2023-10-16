package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/sharding/types"
)

// NewTxCmd is the top-level command for sharding CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the sharding module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdRead(),
	)

	return txCmd
}

// GetCmdRead is a command to test reading
func GetCmdRead() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Short:   "Read from the store",
		Long:    "Specify different read types to test gas costs",
		Args:    cobra.ExactArgs(0),
		Example: fmt.Sprintf(`%[1]s tx sharding read`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRead(
				clientCtx.GetFromAddress().String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdWrite is a command to test reading
func GetCmdWrite() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "write",
		Short:   "Write from the store",
		Long:    "Specify different write types to test gas costs",
		Args:    cobra.ExactArgs(0),
		Example: fmt.Sprintf(`%[1]s tx sharding write`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgWrite(
				clientCtx.GetFromAddress().String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdUpdate is a command to test reading
func GetCmdUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update from the store",
		Long:    "Specify different write types to test gas costs",
		Args:    cobra.ExactArgs(0),
		Example: fmt.Sprintf(`%[1]s tx sharding update`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdate(
				clientCtx.GetFromAddress().String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
