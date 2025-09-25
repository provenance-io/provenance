package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/asset/types"
)

var cmdStart = version.AppName + " tx " + types.ModuleName

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
		GetCmdBurnAsset(),
		GetCmdCreateAsset(),
		GetCmdCreateAssetClass(),
		GetCmdCreatePool(),
		GetCmdCreateTokenization(),
		GetCmdCreateSecuritization(),
	)

	return txCmd
}

// GetCmdBurnAsset returns the command for burning an asset
func GetCmdBurnAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-asset <class-id> <id>",
		Short: "Burn an existing asset",
		Long: strings.TrimSpace(`
Burn an existing asset by removing the NFT and its registry.
Only the owner of the asset can burn it.
`),
		Example: fmt.Sprintf(strings.TrimSpace(`
$ %[1]s burn-asset "real-estate" "property-001"
`), cmdStart),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			assetKey := &types.AssetKey{
				ClassId: args[0],
				Id:      args[1],
			}

			msg := &types.MsgBurnAsset{
				Asset:  assetKey,
				Signer: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateAsset returns the command for creating an asset
func GetCmdCreateAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-asset <class-id> <id> <data> <owner>[--uri <uri>] [--uri-hash <uri-hash>]",
		Short: "Create a new asset",
		Long: strings.TrimSpace(`
Create a new asset in the specified asset class.

If no --owner <owner> is supplied, the --from address is used as the owner.
`),
		Example: fmt.Sprintf(strings.TrimSpace(`
$ %[1]s create-asset "real-estate" "property-001" \
    '{"location": "New York", "value": 500000}' \
    --uri "https://example.com/metadata.json" --uri-hash abc123 \
    --owner tp1jypkeck8vywptdltjnwspwzulkqu7jv6ey90dx
`), cmdStart),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()

			asset := &types.Asset{
				ClassId: args[0],
				Id:      args[1],
				Data:    args[2],
				Uri:     ReadFlagURI(flagSet),
				UriHash: ReadFlagURIHash(flagSet),
			}

			msg := &types.MsgCreateAsset{
				Asset:  asset,
				Owner:  args[3],
				Signer: clientCtx.GetFromAddress().String(),
			}
			if len(msg.Owner) == 0 {
				msg.Owner = msg.Signer
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddFlagsURI(cmd, "asset")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateAssetClass returns the command for creating an asset class
func GetCmdCreateAssetClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-class <id> <name> <data> [--symbol <symbol>] [--description <description>] [--uri <uri>] [--uri-hash <uri-hash>]",
		Short: "Create a new asset class",
		Long:  `Create a new asset class with the specified properties.`,
		Example: fmt.Sprintf(strings.TrimSpace(`
$ %[1]s create-class "real-estate" "Real Estate Assets" \
    '{"category": "property"}' \
    --symbol "REAL" --description "Real estate properties"
    --uri "https://example.com/class-metadata.json" --uri-hash "def456"
`), cmdStart),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()

			assetClass := &types.AssetClass{
				Id:          args[0],
				Name:        args[1],
				Symbol:      ReadFlagSymbol(flagSet),
				Description: ReadFlagDescription(flagSet),
				Uri:         ReadFlagURI(flagSet),
				UriHash:     ReadFlagURIHash(flagSet),
				Data:        args[2],
			}

			msg := &types.MsgCreateAssetClass{
				AssetClass: assetClass,
				Signer:     clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddFlagsURI(cmd, "asset class")
	AddFlagSymbol(cmd)
	AddFlagDescription(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreatePool returns the command for creating a new pool
func GetCmdCreatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool <pool> <nft1> [<nft2> ...]",
		Short: "Create a new pool marker",
		Long: strings.TrimSpace(`
Create a new pool marker with the specified NFTs.

Each <nft> argument should be a comma-separated class-id and asset-id.
`),
		Example: fmt.Sprintf(strings.TrimSpace(`
$ %[1]s create-pool 10pooltoken asset_class1,asset_id1 asset_class2,asset_id2
$ %[1]s create-pool 1000pooltoken real-estate,property-001 real-estate,property-002
`), cmdStart),
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pool, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return types.NewErrCodeInvalidField("pool", "%s", err)
			}

			var assets []*types.AssetKey
			for i, entry := range args[1:] {
				parts := strings.Split(entry, ",")
				if len(parts) != 2 {
					return types.NewErrCodeInvalidField("nft_format", "invalid nft %d format: %q, expected class-id,asset-id", i, entry)
				}
				assets = append(assets, &types.AssetKey{
					ClassId: strings.TrimSpace(parts[0]),
					Id:      strings.TrimSpace(parts[1]),
				})
			}

			msg := &types.MsgCreatePool{
				Pool:   pool,
				Assets: assets,
				Signer: clientCtx.GetFromAddress().String(),
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
		Use:   "create-tokenization <amount> <nft-class-id> <nft-id>",
		Short: "Create a new tokenization marker",
		Long:  `Create a new tokenization marker with the specified amount and NFT.`,
		Example: fmt.Sprintf(strings.TrimSpace(`
$ %[1]s create-tokenization 1000pooltoken real-estate property-001
`), cmdStart),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return types.NewErrCodeInvalidField("coin", "%s", err)
			}

			asset := &types.AssetKey{
				ClassId: args[1],
				Id:      args[2],
			}

			msg := &types.MsgCreateTokenization{
				Token:  coin,
				Asset:  asset,
				Signer: clientCtx.GetFromAddress().String(),
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
		Use:   "create-securitization <id> <pools> <tranches>",
		Short: "Create a new securitization marker and tranches",
		Long: strings.TrimSpace(`
Create a new securitization marker and tranches.

The pools argument should be a comma-separated list of pool names.
The tranches argument should be a comma-separated list of coins.
`),
		Example: fmt.Sprintf(strings.TrimSpace(`
$ %[1]s create-securitization sec1 "pool1,pool2" "100tranche1,200tranche2"
$ %[1]s create-securitization "mortgage-sec-001" "mortgage-pool-1,mortgage-pool-2" "1000000senior-tranche,500000mezzanine-tranche"
`), cmdStart),
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
					return types.NewErrCodeInvalidField("tranche_coin", "%s", err)
				}
				tranches = append(tranches, &coin)
			}

			msg := &types.MsgCreateSecuritization{
				Id:       args[0],
				Pools:    pools,
				Tranches: tranches,
				Signer:   clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
