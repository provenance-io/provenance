package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
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

// CmdQueryOrderFeeCalc TODO
func CmdQueryOrderFeeCalc() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, queryCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdQueryOrderFeeCalc
			return nil
		},
	}

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

	return cmd
}
