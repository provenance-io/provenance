package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		GetCmdCreateAsset(),
		GetCmdCreateAssetClass(),
		GetCmdCreatePool(),
		GetCmdCreateTokenization(),
		GetCmdCreateSecuritization(),
	)

	return txCmd
}

// GetCmdCreateAsset returns the command for creating an asset
func GetCmdCreateAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-asset [class-id] [id] [uri] [uri-hash] [data]",
		Short: "Create a new asset",
		Long: `Create a new asset in the specified asset class.`,
		Example: `  provenanced tx asset create-asset "real-estate" "property-001" "https://example.com/metadata.json" "abc123" '{"location": "New York", "value": 500000}'`,
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

			msg := &types.MsgCreateAsset{
				Asset:       asset,
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateAssetClass returns the command for creating an asset class
func GetCmdCreateAssetClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-class [id] [name] [symbol] [description] [uri] [uri-hash] [data]",
		Short: "Create a new asset class",
		Long: `Create a new asset class with the specified properties.`,
		Example: `  provenanced tx asset create-class "real-estate" "Real Estate Assets" "REAL" "Real estate properties" "https://example.com/class-metadata.json" "def456" '{"category": "property"}'`,
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

			msg := &types.MsgCreateAssetClass{
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
		Use:   "create-pool [pool] [nfts]",
		Short: "Create a new pool marker",
		Long: `Create a new pool marker with the specified NFTs.
The nfts argument should be a semicolon-separated list of asset entries, where each entry is a comma-separated class-id and asset-id.
The entire nfts argument must be quoted to prevent shell interpretation of the semicolons.`,
		Example: `  provenanced tx asset create-pool 10pooltoken "asset_class1,asset_id1;asset_class2,asset_id2"
  provenanced tx asset create-pool 10pooltoken 'asset_class1,asset_id1;asset_class2,asset_id2'
  provenanced tx asset create-pool 1000pooltoken "real-estate,property-001;real-estate,property-002"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pool, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid pool %s", args[0])
			}

			var assets []*types.AssetKey
			assetEntries := strings.Split(args[1], ";")
			for _, entry := range assetEntries {
				parts := strings.Split(entry, ",")
				if len(parts) != 2 {
					return fmt.Errorf("invalid nft format: %s, expected class-id,asset-id", entry)
				}
				assets = append(assets, &types.AssetKey{
					ClassId: strings.TrimSpace(parts[0]),
					Id:      strings.TrimSpace(parts[1]),
				})
			}

			msg := &types.MsgCreatePool{
				Pool:        &pool,
				Assets:        assets,
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateTokenization returns the command for creating a new tokenization
func GetCmdCreateTokenization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-tokenization [amount] [nft-class-id] [nft-id]",
		Short: "Create a new tokenization marker",
		Long: `Create a new tokenization marker with the specified amount and NFT.`,
		Example: `  provenanced tx asset create-tokenization 1000pooltoken real-estate property-001`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid coin %s", args[0])
			}

			asset := &types.AssetKey{
				ClassId: args[1],
				Id:      args[2],
			}

			msg := &types.MsgCreateTokenization{
				Denom:       coin,
				Asset:         asset,
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
		Use:   "create-securitization [id] [pools] [tranches]",
		Short: "Create a new securitization marker and tranches",
		Long: `Create a new securitization marker and tranches.
The pools argument should be a comma-separated list of pool names.
The tranches argument should be a comma-separated list of coins.`,
		Example: `  provenanced tx asset create-securitization sec1 "pool1,pool2" "100tranche1,200tranche2"
  provenanced tx asset create-securitization "mortgage-sec-001" "mortgage-pool-1,mortgage-pool-2" "1000000senior-tranche,500000mezzanine-tranche"`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse the comma-separated list of pool names
			pools := strings.Split(args[1], ",")
			for i, pool := range pools {
				pools[i] = strings.TrimSpace(pool)
			}

			// Parse the comma-separated list of coins
			trancheStrings := strings.Split(args[2], ",")
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
				Pools:       pools,
				Tranches:    tranches,
				FromAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
