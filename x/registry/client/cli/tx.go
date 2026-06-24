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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/registry/types"
)

// CmdTx returns the transaction commands for the registry module
func CmdTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the registry module",
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
		CmdProposeRoleChange(),
		CmdApproveRoleChange(),
		CmdCreateRegistryClass(),
		CmdUpdateRegistryClassRoleAuthorization(),
		CmdUpdateParams(),
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
		Example: `$ provenanced tx registry bulk-update entries.json --from mykey
		where the json is formatted as follows (array of RegistryEntry type):
		[
			{
				"key": { "asset_class_id": "loan.asset", "nft_id": "loan-001" },
				"roles": [
					{ "role": 6, "addresses": ["tp1q75h9mg54a6nrr9xqldyh0r82e5lf5wqrxzhqp"] },
					{ "role": 1, "addresses": ["tp1sh49f6ze3vn7cdl2amh2gnc70z5mten3y3fv0k"] }
				]
			},
			{
				"key": { "asset_class_id": "loan.asset", "nft_id": "loan-002" },
				"roles": [
					{ "role": 6, "addresses": ["tp1q75h9mg54a6nrr9xqldyh0r82e5lf5wqrxzhqp"] }
				]
			}
		]
		
		Role values: 1=SERVICER, 2=SUBSERVICER, 3=CONTROLLER, 4=CUSTODIAN, 5=BORROWER, 6=ORIGINATOR
		`,
		Args: cobra.ExactArgs(1),
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

// CmdProposeRoleChange returns the command to propose a multi-party role change
func CmdProposeRoleChange() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-role-change <asset_class_id> <nft_id> <role=address[,address...]> [role=address[,address...]...]",
		Short: "Propose a desired-state role change that accumulates approvals until every affected role's policy is satisfied",
		Long: `Propose a batch of desired-state role updates that accumulates single-signer approvals until
every affected role's authorization policy is satisfied, then applies atomically.

Each role update is "role=address[,address...]". An empty address list clears the role.

Example:
  propose-role-change myclass nft1 controller=cosmos1new secured_party_for_enote=`,
		Args: cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			roleUpdates, err := parseRoleUpdates(args[2:])
			if err != nil {
				return err
			}

			msg := types.MsgProposeRoleChange{
				Signer: clientCtx.GetFromAddress().String(),
				Key: &types.RegistryKey{
					AssetClassId: args[0],
					NftId:        args[1],
				},
				RoleUpdates: roleUpdates,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdApproveRoleChange returns the command to approve a pending role change
func CmdApproveRoleChange() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve-role-change <change_id>",
		Short: "Approve a pending role change by its id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgApproveRoleChange{
				Signer:   clientCtx.GetFromAddress().String(),
				ChangeId: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateRegistryClass returns the command to create a registry class.
// The class definition (including role_authorizations) is read from a JSON file in proto-JSON
// format so that oneof fields in the authorization policy are decoded correctly.
func CmdCreateRegistryClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-registry-class <class_json_file>",
		Short: "Create a registry class from a JSON file",
		Long: `Create a registry class defining asset class-level authorization rules.

The file is a proto-JSON RegistryClass, e.g.:
{
  "registryClassId": "dart-loan-v1",
  "asset_class_id": "loan.asset",
  "maintainer": "pb1...",
  "role_authorizations": [ { "role": "REGISTRY_ROLE_CONTROLLER", "authorizations": [ ... ] } ]
}

If "maintainer" is omitted it defaults to the --from signer. The signer must equal the maintainer.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jsonData, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var class types.RegistryClass
			if err := clientCtx.Codec.UnmarshalJSON(jsonData, &class); err != nil {
				return fmt.Errorf("failed to unmarshal registry class JSON: %w", err)
			}

			signer := clientCtx.GetFromAddress().String()
			maintainer := class.Maintainer
			if maintainer == "" {
				maintainer = signer
			}

			msg := types.MsgCreateRegistryClass{
				Signer:             signer,
				RegistryClassId:    class.RegistryClassId,
				AssetClassId:       class.AssetClassId,
				Maintainer:         maintainer,
				RoleAuthorizations: class.RoleAuthorizations,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateRegistryClassRoleAuthorization returns the command to replace a registry class's
// authorization rules. The new rules are read from a proto-JSON RegistryClass file (only its
// role_authorizations are used).
func CmdUpdateRegistryClassRoleAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-registry-class <registry_class_id> <class_json_file>",
		Short: "Replace a registry class's authorization rules from a JSON file",
		Long: `Replace the authorization rules for an existing registry class. Only the current
maintainer may update the rules. The file is a proto-JSON RegistryClass; only its
role_authorizations field is used.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jsonData, err := os.ReadFile(args[1])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var class types.RegistryClass
			if err := clientCtx.Codec.UnmarshalJSON(jsonData, &class); err != nil {
				return fmt.Errorf("failed to unmarshal registry class JSON: %w", err)
			}

			msg := types.MsgUpdateRegistryClassRoleAuthorization{
				Signer:             clientCtx.GetFromAddress().String(),
				RegistryClassId:    args[0],
				RoleAuthorizations: class.RoleAuthorizations,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateParams returns the command to update the registry module params via a governance
// proposal. The params are read from a proto-JSON Params file. This message must be submitted
// through governance; the authority defaults to the governance module account address and may be
// overridden with --authority.
func CmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params <params_json_file>",
		Short: "Update the registry module params from a JSON file (governance)",
		Long: `Update the registry module params from a proto-JSON Params file, e.g.:
{
  "role_authorizations": [ { "role": "REGISTRY_ROLE_CONTROLLER", "authorizations": [ ... ] } ]
}

This message must be authorized by governance. The authority defaults to the governance module
account address; use --authority to override. Submit the generated message through
"tx gov submit-proposal".`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jsonData, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var params types.Params
			if err := clientCtx.Codec.UnmarshalJSON(jsonData, &params); err != nil {
				return fmt.Errorf("failed to unmarshal params JSON: %w", err)
			}

			authority, err := cmd.Flags().GetString("authority")
			if err != nil {
				return err
			}
			if authority == "" {
				authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()
			}

			msg := types.MsgUpdateParams{
				Authority: authority,
				Params:    params,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	cmd.Flags().String("authority", "", "the authority address (defaults to the gov module account)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// An empty address list (e.g. "role=") clears the role.
func parseRoleUpdates(args []string) ([]types.RoleUpdate, error) {
	updates := make([]types.RoleUpdate, 0, len(args))
	for _, arg := range args {
		name, addrCSV, ok := strings.Cut(arg, "=")
		if !ok {
			return nil, fmt.Errorf("invalid role update %q: expected format \"role=address[,address...]\"", arg)
		}

		role, err := types.ParseRegistryRole(name)
		if err != nil {
			return nil, err
		}

		var addresses []string
		if trimmed := strings.TrimSpace(addrCSV); trimmed != "" {
			for _, addr := range strings.Split(trimmed, ",") {
				addr = strings.TrimSpace(addr)
				if addr != "" {
					addresses = append(addresses, addr)
				}
			}
		}

		updates = append(updates, types.RoleUpdate{Role: role, Addresses: addresses})
	}
	return updates, nil
}
