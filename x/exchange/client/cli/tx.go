package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/exchange"
)

var txCmdStart = fmt.Sprintf("%s tx %s", version.AppName, exchange.ModuleName)

func CmdTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        exchange.ModuleName,
		Aliases:                    []string{"ex"},
		Short:                      "Transaction commands for the exchange module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdTxCreateAsk(),
		CmdTxCreateBid(),
		CmdTxCancelOrder(),
		CmdTxFillBids(),
		CmdTxFillAsks(),
		CmdTxMarketSettle(),
		CmdTxMarketSetOrderExternalID(),
		CmdTxMarketWithdraw(),
		CmdTxMarketUpdateDetails(),
		CmdTxMarketUpdateEnabled(),
		CmdTxMarketUpdateUserSettle(),
		CmdTxMarketManagePermissions(),
		CmdTxMarketManageReqAttrs(),
		CmdTxGovCreateMarket(),
		CmdTxGovManageFees(),
		CmdTxGovUpdateParams(),
	)

	return cmd
}

// CmdTxCreateAsk TODO
func CmdTxCreateAsk() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxCreateAsk
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxCreateBid TODO
func CmdTxCreateBid() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxCreateBid
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxCancelOrder TODO
func CmdTxCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxCancelOrder
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxFillBids TODO
func CmdTxFillBids() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxFillBids
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxFillAsks TODO
func CmdTxFillAsks() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxFillAsks
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketSettle TODO
func CmdTxMarketSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketSettle
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketSetOrderExternalID TODO
func CmdTxMarketSetOrderExternalID() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketSetOrderExternalID
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketWithdraw TODO
func CmdTxMarketWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketWithdraw
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketUpdateDetails TODO
func CmdTxMarketUpdateDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketUpdateDetails
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketUpdateEnabled TODO
func CmdTxMarketUpdateEnabled() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketUpdateEnabled
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketUpdateUserSettle TODO
func CmdTxMarketUpdateUserSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketUpdateUserSettle
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketManagePermissions TODO
func CmdTxMarketManagePermissions() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketManagePermissions
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketManageReqAttrs TODO
func CmdTxMarketManageReqAttrs() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketManageReqAttrs
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxGovCreateMarket TODO
func CmdTxGovCreateMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxGovCreateMarket
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxGovManageFees TODO
func CmdTxGovManageFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxGovManageFees
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxGovUpdateParams TODO
func CmdTxGovUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxGovUpdateParams
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
