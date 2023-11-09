package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/exchange"
)

var queryCmdStart = fmt.Sprintf("%s query %s", version.AppName, exchange.ModuleName)

func CmdQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        exchange.ModuleName,
		Aliases:                    []string{"ex"},
		Short:                      "Querying commands for the exchange module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryOrderFeeCalc(),
		CmdQueryGetOrder(),
		CmdQueryGetOrderByExternalID(),
		CmdQueryGetMarketOrders(),
		CmdQueryGetOwnerOrders(),
		CmdQueryGetAssetOrders(),
		CmdQueryGetAllOrders(),
		CmdQueryGetMarket(),
		CmdQueryGetAllMarkets(),
		CmdQueryParams(),
		CmdQueryValidateCreateMarket(),
		CmdQueryValidateMarket(),
		CmdQueryValidateManageFees(),
	)

	return cmd
}

// CmdQueryOrderFeeCalc creates the order-fee-calc sub-command for the exchange query command.
func CmdQueryOrderFeeCalc() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "order-fee-calc",
		Aliases: []string{"fee-calc", "order-calc"},
		Short:   "Calculate the fees for an order",
		RunE:    genericQueryRunE(MakeQueryOrderFeeCalc, exchange.QueryClient.OrderFeeCalc),
	}

	SetupCmdQueryOrderFeeCalc(cmd)
	return cmd
}

// CmdQueryGetOrder creates the order sub-command for the exchange query command.
func CmdQueryGetOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "order",
		Aliases: []string{"get-order"},
		Short:   "Get an order by id",
		RunE:    genericQueryRunE(MakeQueryGetOrder, exchange.QueryClient.GetOrder),
	}

	SetupCmdQueryGetOrder(cmd)
	return cmd
}

// CmdQueryGetOrderByExternalID creates the order-by-external-id sub-command for the exchange query command.
func CmdQueryGetOrderByExternalID() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "order-by-external-id",
		Aliases: []string{"get-order-by-external-id", "by-external-id", "external-id"},
		Short:   "Get an order by market id and external id",
		RunE:    genericQueryRunE(MakeQueryGetOrderByExternalID, exchange.QueryClient.GetOrderByExternalID),
	}

	SetupCmdQueryGetOrderByExternalID(cmd)
	return cmd
}

// CmdQueryGetMarketOrders creates the market-orders sub-command for the exchange query command.
func CmdQueryGetMarketOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-orders",
		Aliases: []string{"get-market-orders"},
		Short:   "Look up orders for a market",
		RunE:    genericQueryRunE(MakeQueryGetMarketOrders, exchange.QueryClient.GetMarketOrders),
	}

	SetupCmdQueryGetMarketOrders(cmd)
	return cmd
}

// CmdQueryGetOwnerOrders creates the owner-orders sub-command for the exchange query command.
func CmdQueryGetOwnerOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "owner-orders",
		Aliases: []string{"get-owner-orders"},
		Short:   "Look up orders with a specific owner",
		RunE:    genericQueryRunE(MakeQueryGetOwnerOrders, exchange.QueryClient.GetOwnerOrders),
	}

	SetupCmdQueryGetOwnerOrders(cmd)
	return cmd
}

// CmdQueryGetAssetOrders creates the asset-orders sub-command for the exchange query command.
func CmdQueryGetAssetOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "asset-orders",
		Aliases: []string{"get-asset-orders", "denom-orders", "get-denom-orders"},
		Short:   "Look up orders with a specific asset denom",
		RunE:    genericQueryRunE(MakeQueryGetAssetOrders, exchange.QueryClient.GetAssetOrders),
	}

	SetupCmdQueryGetAssetOrders(cmd)
	return cmd
}

// CmdQueryGetAllOrders creates the all-orders sub-command for the exchange query command.
func CmdQueryGetAllOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "all-orders",
		Aliases: []string{"get-all-orders"},
		Short:   "Get all orders",
		RunE:    genericQueryRunE(MakeQueryGetAllOrders, exchange.QueryClient.GetAllOrders),
	}

	SetupCmdQueryGetAllOrders(cmd)
	return cmd
}

// CmdQueryGetMarket creates the market sub-command for the exchange query command.
func CmdQueryGetMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market",
		Aliases: []string{"get-market"},
		Short:   "Get market setup and information",
		RunE:    genericQueryRunE(MakeQueryGetMarket, exchange.QueryClient.GetMarket),
	}

	SetupCmdQueryGetMarket(cmd)
	return cmd
}

// CmdQueryGetAllMarkets creates the all-markets sub-command for the exchange query command.
func CmdQueryGetAllMarkets() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "all-markets",
		Aliases: []string{"get-all-markets"},
		Short:   "Get all markets",
		RunE:    genericQueryRunE(MakeQueryGetAllMarkets, exchange.QueryClient.GetAllMarkets),
	}

	SetupCmdQueryGetAllMarkets(cmd)
	return cmd
}

// CmdQueryParams TODO
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryParams
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryValidateCreateMarket TODO
func CmdQueryValidateCreateMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryValidateCreateMarket
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryValidateMarket TODO
func CmdQueryValidateMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryValidateMarket
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryValidateManageFees TODO
func CmdQueryValidateManageFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryValidateManageFees
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
