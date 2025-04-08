package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/provenance-io/provenance/x/registry"
)

// CmdTx returns the transaction commands for the registry module
func CmdTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        registry.ModuleName,
		Short:                      "Registry transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdRegisterAddress(),
		CmdUpdateRoles(),
		CmdRemoveAddress(),
	)

	return cmd
}

// CmdRegisterAddress returns the command to register a new address
func CmdRegisterAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-address [address] [roles]",
		Short: "Register a new address with roles",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := registry.MsgRegisterAddress{
				Authority: clientCtx.GetFromAddress().String(),
				Address:   args[0],
				Roles:     map[string]registry.RoleAddresses{},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateRoles returns the command to update roles for an address
func CmdUpdateRoles() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-roles [address] [roles]",
		Short: "Update roles for an address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := registry.MsgUpdateRoles{
				Authority: clientCtx.GetFromAddress().String(),
				Address:   args[0],
				Roles:     map[string]registry.RoleAddresses{},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRemoveAddress returns the command to remove an address
func CmdRemoveAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-address [address]",
		Short: "Remove an address from the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := registry.MsgRemoveAddress{
				Authority: clientCtx.GetFromAddress().String(),
				Address:   args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
