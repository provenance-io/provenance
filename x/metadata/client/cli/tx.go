package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/provenance-io/provenance/x/metadata/types"

	uuid "github.com/google/uuid"
)

// NewTxCmd is the top-level command for attribute CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"m"},
		Short:                      "Transaction commands for the metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		AddMetadataScopeCmd(),
	)

	return txCmd
}

//  AddMetadataScopeCmd creates a command for adding an account attributes.
func AddMetadataScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-scope [scope-uuid] [spec-id] [owner-addresses] [data-access] [value-owner-address] [signers]",
		Short: "Add an metadata scope to the provenance blockchain",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			scopeUUID, err := uuid.Parse(args[0])
			if err != nil {
				fmt.Printf("Invalid uuid for scope id: %s", args[0])
				return nil
			}
			specIdUUID, err := uuid.Parse(args[1])
			if err != nil {
				fmt.Printf("Invalid uuid for specification id: %s", args[0])
				return nil
			}

			specId := types.ScopeSpecMetadataAddress(specIdUUID)

			//TODO: Validate addresses
			ownerAddresses := strings.Split(args[2], ",")
			dataAccess := strings.Split(args[3], ",")
			valueOwnerAddress := args[4]
			signers := strings.Split(args[5], ",")
			scope := *types.NewScope(
				types.ScopeMetadataAddress(scopeUUID),
				specId,
				ownerAddresses,
				dataAccess,
				valueOwnerAddress)

			if err := scope.ValidateBasic(); err != nil {
				fmt.Printf("Failed to validate scope %s : %v", scope.String(), err)
				return nil
			}

			msg := types.NewMsgAddScopeRequest(
				scope,
				signers)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
