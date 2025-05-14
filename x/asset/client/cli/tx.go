package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"encoding/json"
	"fmt"

	"github.com/provenance-io/provenance/x/asset/types"
	"github.com/provenance-io/provenance/x/ledger"
)

// GetTxCmd returns the transaction commands for the asset module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Asset transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdAddAsset(),
		GetCmdAddAssetClass(),
		GetCmdCreatePool(),
	)

	return txCmd
}

// GetCmdAddAsset returns the command for adding an asset
func GetCmdAddAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-asset [class-id] [id] [uri] [uri-hash] [data] [entry-types] [status-types]",
		Short: "Add a new asset",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			asset := &types.Asset{
				ClassId: args[0],
				Id:      args[1],
				Uri:     args[2],
				UriHash: args[3],
				Data:    args[4],
			}

			// Parse entry types JSON array
			var entryTypes []*ledger.LedgerClassEntryType
			if err := json.Unmarshal([]byte(args[5]), &entryTypes); err != nil {
				return fmt.Errorf("invalid entry-types JSON array: %w", err)
			}

			// Parse status types JSON array
			var statusTypes []*ledger.LedgerClassStatusType
			if err := json.Unmarshal([]byte(args[6]), &statusTypes); err != nil {
				return fmt.Errorf("invalid status-types JSON array: %w", err)
			}

			msg := &types.MsgAddAsset{
				Asset:       asset,
				EntryTypes:  entryTypes,
				StatusTypes: statusTypes,
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddAssetClass returns the command for adding an asset class
func GetCmdAddAssetClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-class [id] [name] [symbol] [description] [uri] [uri-hash] [data]",
		Short: "Add a new asset class",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			assetClass := &types.AssetClass{
				Id:          args[0],
				Name:        args[1],
				Symbol:      args[2],
				Description: args[3],
				Uri:         args[4],
				UriHash:     args[5],
				Data:        args[6],
			}

			msg := &types.MsgAddAssetClass{
				AssetClass:  assetClass,
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreatePool returns the command for creating a new pool
func GetCmdCreatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [pool-id]",
		Short: "Create a new pool marker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCreatePool{
				PoolId:      args[0],
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
