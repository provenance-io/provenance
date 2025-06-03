package cli

import (
	"fmt"

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
		CmdRegisterNFT(),
		CmdGrantRole(),
		CmdRevokeRole(),
		CmdUnregisterNFT(),
	)

	return cmd
}

// CmdRegisterNFT returns the command to register a new NFT
func CmdRegisterNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-nft [asset_class_id] [nft_id]",
		Short: "Register a new NFT with roles",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse key from args
			key := registry.RegistryKey{
				AssetClassId: args[0],
				NftId:        args[1],
			}

			// TODO: Parse key and roles from args
			msg := registry.MsgRegisterNFT{
				Authority: clientCtx.GetFromAddress().String(),
				Key:       &key,
				Roles:     []registry.RolesEntry{},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdGrantRole returns the command to grant a role
func CmdGrantRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant-role [asset_class_id] [nft_id] [role] [addresses]",
		Short: "Grant a role to addresses",
		Args:  cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			key := registry.RegistryKey{
				AssetClassId: args[0],
				NftId:        args[1],
			}

			// convert arg[2] to a registry.RegistryRole enum value
			role, ok := registry.RegistryRole_value[args[2]]
			if !ok {
				return fmt.Errorf("invalid role: %s", args[2])
			}

			// TODO: Parse key, role and addresses from args
			msg := registry.MsgGrantRole{
				Authority: clientCtx.GetFromAddress().String(),
				Key:       &key,
				Role:      registry.RegistryRole(role),
				Addresses: args[3:],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRevokeRole returns the command to revoke a role
func CmdRevokeRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-role [asset_class_id] [nft_id] [role] [addresses]",
		Short: "Revoke a role from addresses",
		Args:  cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			key := registry.RegistryKey{
				AssetClassId: args[0],
				NftId:        args[1],
			}

			// convert arg[2] to a registry.RegistryRole enum value
			role, ok := registry.RegistryRole_value[args[2]]
			if !ok {
				return fmt.Errorf("invalid role: %s", args[2])
			}

			// TODO: Parse key, role and addresses from args
			msg := registry.MsgRevokeRole{
				Authority: clientCtx.GetFromAddress().String(),
				Key:       &key,
				Role:      registry.RegistryRole(role),
				Addresses: args[3:],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUnregisterNFT returns the command to unregister an NFT
func CmdUnregisterNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister-nft [asset_class_id] [nft_id]",
		Short: "Unregister an NFT",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			key := registry.RegistryKey{
				AssetClassId: args[0],
				NftId:        args[1],
			}

			// TODO: Parse key from args
			msg := registry.MsgUnregisterNFT{
				Authority: clientCtx.GetFromAddress().String(),
				Key:       &key,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
