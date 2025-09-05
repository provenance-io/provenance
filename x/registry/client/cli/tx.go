package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/provenance-io/provenance/x/registry/types"
)

// CmdTx returns the transaction commands for the registry module
func CmdTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
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
		CmdRegistryBulkUpdate(),
	)

	return cmd
}

// CmdRegisterNFT returns the command to register a new NFT
func CmdRegisterNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-nft <asset_class_id> <nft_id>",
		Short: "Register an NFT",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgRegisterNFT{
				Signer: clientCtx.GetFromAddress().String(),
				Key: &types.RegistryKey{
					AssetClassId: args[0],
					NftId:        args[1],
				},
				Roles: []types.RolesEntry{},
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
		Use:   "grant-role <asset_class_id> <nft_id> <role> <address>",
		Short: "Grant a role to an address",
		// TODO[danny]: Create Long message that utilizes ValidRolesString() to list the roles available.
		Args: cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			role, err := types.ParseRegistryRole(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgGrantRole{
				Signer: clientCtx.GetFromAddress().String(),
				Key: &types.RegistryKey{
					AssetClassId: args[0],
					NftId:        args[1],
				},
				Role:      role,
				Addresses: []string{args[3]},
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
		Use:   "revoke-role <asset_class_id> <nft_id> <role> <address>",
		Short: "Revoke a role from an address",
		// TODO[danny]: Create Long message that utilizes ValidRolesString() to list the roles available.
		Args: cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// convert arg[2] to a registry.RegistryRole enum value
			role, err := types.ParseRegistryRole(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgRevokeRole{
				Signer: clientCtx.GetFromAddress().String(),
				Key: &types.RegistryKey{
					AssetClassId: args[0],
					NftId:        args[1],
				},
				Role:      role,
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
		Use:   "unregister-nft <asset_class_id> <nft_id>",
		Short: "Unregister an NFT",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgUnregisterNFT{
				Signer: clientCtx.GetFromAddress().String(),
				Key: &types.RegistryKey{
					AssetClassId: args[0],
					NftId:        args[1],
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRegistryBulkUpdate returns the command to bulk update registry entries
func CmdRegistryBulkUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-update <entries_json_file>",
		Short: "Bulk update registry entries",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// read the json file
			jsonData, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			// unmarshal the json file to []types.RegistryEntry
			var entries []types.RegistryEntry
			if err := json.Unmarshal(jsonData, &entries); err != nil {
				return fmt.Errorf("failed to unmarshal JSON array: %w", err)
			}

			msg := types.MsgRegistryBulkUpdate{
				Signer:  clientCtx.GetFromAddress().String(),
				Entries: entries,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
