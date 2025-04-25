package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/asset/types"
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
	)

	return txCmd
}

// GetCmdAddAsset returns the command for adding an asset
func GetCmdAddAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-asset [class-id] [id] [uri] [uri-hash] [data]",
		Short: "Add a new asset",
		Args:  cobra.ExactArgs(5),
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

			msg := &types.MsgAddAsset{
				Asset: asset,
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
