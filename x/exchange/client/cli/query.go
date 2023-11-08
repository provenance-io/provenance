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
		Args:    cobra.NoArgs,
		RunE:    genericQueryRunE(MakeQueryOrderFeeCalc, exchange.QueryClient.OrderFeeCalc),
	}

	flags.AddQueryFlagsToCmd(cmd)
	AddFlagsQueryOrderFeeCalc(cmd)
	return cmd
}

// CmdQueryGetOrder TODO
func CmdQueryGetOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryGetOrder
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryGetOrderByExternalID TODO
func CmdQueryGetOrderByExternalID() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryGetOrderByExternalID
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryGetMarketOrders TODO
func CmdQueryGetMarketOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryGetMarketOrders
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "orders")
	return cmd
}

// CmdQueryGetOwnerOrders TODO
func CmdQueryGetOwnerOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryGetOwnerOrders
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "orders")
	return cmd
}

// CmdQueryGetAssetOrders TODO
func CmdQueryGetAssetOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryGetAssetOrders
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "orders")
	return cmd
}

// CmdQueryGetAllOrders TODO
func CmdQueryGetAllOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryGetAllOrders
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "orders")
	return cmd
}

// CmdQueryGetMarket TODO
func CmdQueryGetMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryGetMarket
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryGetAllMarkets TODO
func CmdQueryGetAllMarkets() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryGetAllMarkets
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "markets")
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
