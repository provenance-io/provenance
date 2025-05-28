package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

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
		GetCmdCreateParticipation(),
		GetCmdCreateSecuritization(),
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
				Asset:       asset,
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
		Use:   "add-class [id] [name] [symbol] [description] [uri] [uri-hash] [data] [entry-types] [status-types]",
		Short: "Add a new asset class",
		Args:  cobra.ExactArgs(9),
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

			// Parse entry types JSON array
			var entryTypes []*ledger.LedgerClassEntryType
			if err := json.Unmarshal([]byte(args[7]), &entryTypes); err != nil {
				return fmt.Errorf("invalid entry-types JSON array: %w", err)
			}

			// Parse status types JSON array
			var statusTypes []*ledger.LedgerClassStatusType
			if err := json.Unmarshal([]byte(args[8]), &statusTypes); err != nil {
				return fmt.Errorf("invalid status-types JSON array: %w", err)
			}

			msg := &types.MsgAddAssetClass{
				AssetClass:  assetClass,
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

// GetCmdCreatePool returns the command for creating a new pool
func GetCmdCreatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [pool] [nfts]",
		Short: "Create a new pool marker",
		Long: `Create a new pool marker with the specified NFTs.
The nfts argument should be a semicolon-separated list of asset entries, where each entry is a comma-separated class-id and asset-id.
The entire nfts argument must be quoted to prevent shell interpretation of the semicolons.

Example: 
  provenanced tx asset create-pool 10pooltoken "asset_class1,asset_id1;asset_class2,asset_id2"
  provenanced tx asset create-pool 10pooltoken 'asset_class1,asset_id1;asset_class2,asset_id2'`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pool, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid pool %s", args[0])
			}

			var nfts []*types.Nft
			nftEntries := strings.Split(args[1], ";")
			for _, entry := range nftEntries {
				parts := strings.Split(entry, ",")
				if len(parts) != 2 {
					return fmt.Errorf("invalid nft format: %s, expected class-id,asset-id", entry)
				}
				nfts = append(nfts, &types.Nft{
					ClassId: strings.TrimSpace(parts[0]),
					Id:      strings.TrimSpace(parts[1]),
				})
			}

			msg := &types.MsgCreatePool{
				Pool:        &pool,
				Nfts:        nfts,
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateParticipation returns the command for creating a new participation
func GetCmdCreateParticipation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-participation [pool-id] [amount]",
		Short: "Create a new participation marker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid coin %s", args[0])
			}

			msg := &types.MsgCreateParticipation{
				Denom:       coin,
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateSecuritization returns the command for creating a new securitization
func GetCmdCreateSecuritization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-securitization [id] [tranches]",
		Short: "Create a new securitization marker and tranches",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse the comma-separated list of coins
			trancheStrings := strings.Split(args[1], ",")
			var tranches []*sdk.Coin

			for _, trancheStr := range trancheStrings {
				coin, err := sdk.ParseCoinNormalized(strings.TrimSpace(trancheStr))
				if err != nil {
					return fmt.Errorf("invalid coin %s: %w", trancheStr, err)
				}
				tranches = append(tranches, &coin)
			}

			msg := &types.MsgCreateSecuritization{
				Id:          args[0],
				Tranches:    tranches,
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
