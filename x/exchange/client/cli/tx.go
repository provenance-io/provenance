package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/x/exchange"
)

// CmdTx creates the tx command (and sub-commands) for the exchange module.
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

// CmdTxCreateAsk creates the create-ask sub-command for the exchange tx command.
func CmdTxCreateAsk() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-ask",
		Aliases: []string{"ask", "create-ask-order", "ask-order"},
		Short:   "Create an ask order",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgCreateAsk),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgCreateAsk(cmd)
	return cmd
}

// CmdTxCreateBid creates the create-bid sub-command for the exchange tx command.
func CmdTxCreateBid() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-bid",
		Aliases: []string{"bid", "create-bid-order", "bid-order"},
		Short:   "Create a bid order",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgCreateBid),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgCreateBid(cmd)
	return cmd
}

// CmdTxCancelOrder creates the cancel-order sub-command for the exchange tx command.
func CmdTxCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel-order",
		Aliases: []string{"cancel"},
		Short:   "Cancel an order",
		Args:    cobra.MaximumNArgs(1),
		RunE:    genericTxRunE(MakeMsgCancelOrder),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgCancelOrder(cmd)
	return cmd
}

// CmdTxFillBids creates the fill-bids sub-command for the exchange tx command.
func CmdTxFillBids() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill-bids",
		Short: "Fill one or more bid orders",
		Args:  cobra.NoArgs,
		RunE:  genericTxRunE(MakeMsgFillBids),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgFillBids(cmd)
	return cmd
}

// CmdTxFillAsks creates the fill-asks sub-command for the exchange tx command.
func CmdTxFillAsks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill-asks",
		Short: "Fill one or more ask orders",
		Args:  cobra.NoArgs,
		RunE:  genericTxRunE(MakeMsgFillAsks),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgFillAsks(cmd)
	return cmd
}

// CmdTxMarketSettle creates the market-settle sub-command for the exchange tx command.
func CmdTxMarketSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-settle",
		Aliases: []string{"settle"},
		Short:   "Settle some orders",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgMarketSettle),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgMarketSettle(cmd)
	return cmd
}

// CmdTxMarketSetOrderExternalID creates the market-set-external-id sub-command for the exchange tx command.
func CmdTxMarketSetOrderExternalID() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-set-external-id",
		Aliases: []string{"market-set-order-external-id", "set-external-id", "external-id"},
		Short:   "Set an order's external id",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgMarketSetOrderExternalID),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgMarketSetOrderExternalID(cmd)
	return cmd
}

// CmdTxMarketWithdraw creates the market-withdraw sub-command for the exchange tx command.
func CmdTxMarketWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-withdraw",
		Aliases: []string{"withdraw"},
		Short:   "Withdraw funds from a market account",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgMarketWithdraw),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgMarketWithdraw(cmd)
	return cmd
}

// CmdTxMarketUpdateDetails creates the market-details sub-command for the exchange tx command.
func CmdTxMarketUpdateDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-details",
		Aliases: []string{"market-update-details", "update-details", "details"},
		Short:   "Update a market's details",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgMarketUpdateDetails),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgMarketUpdateDetails(cmd)
	return cmd
}

// CmdTxMarketUpdateEnabled creates the market-enabled sub-command for the exchange tx command.
func CmdTxMarketUpdateEnabled() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-enabled",
		Aliases: []string{"market-update-enabled", "update-enabled"},
		Short:   "Change whether a market is accepting orders",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgMarketUpdateEnabled),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgMarketUpdateEnabled(cmd)
	return cmd
}

// CmdTxMarketUpdateUserSettle creates the market-user-settle sub-command for the exchange tx command.
func CmdTxMarketUpdateUserSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-user-settle",
		Aliases: []string{"market-update-user-settle", "update-user-settle"},
		Short:   "Change whether a market allows settlements initiated by users",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgMarketUpdateUserSettle),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgMarketUpdateUserSettle(cmd)
	return cmd
}

// CmdTxMarketManagePermissions creates the market-permissions sub-command for the exchange tx command.
func CmdTxMarketManagePermissions() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-permissions",
		Aliases: []string{"market-manage-permissions", "permissions", "market-perms", "market-manage-perms", "perms"},
		Short:   "Update the account permissions for a market",
		Args:    cobra.NoArgs,
		RunE:    genericTxRunE(MakeMsgMarketManagePermissions),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgMarketManagePermissions(cmd)
	return cmd
}

// CmdTxMarketManageReqAttrs creates the market-req-attrs sub-command for the exchange tx command.
func CmdTxMarketManageReqAttrs() *cobra.Command {
	cmd := &cobra.Command{
		Use: "market-req-attrs",
		Aliases: []string{"market-manage-req-attrs", "manage-req-attrs", "req-attrs", "market-required-attributes",
			"market-manage-required-attributes", "manage-required-attributes", "required-attributes",
		},
		Short: "Manage the attributes required to create orders in a market",
		Args:  cobra.NoArgs,
		RunE:  genericTxRunE(MakeMsgMarketManageReqAttrs),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgMarketManageReqAttrs(cmd)
	return cmd
}

// CmdTxGovCreateMarket creates the gov-create-market sub-command for the exchange tx command.
func CmdTxGovCreateMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gov-create-market",
		Aliases: []string{"create-market"},
		Short:   "Submit a governance proposal to create a market",
		Args:    cobra.NoArgs,
		RunE:    govTxRunE(MakeMsgGovCreateMarket),
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	AddFlagsMsgGovCreateMarket(cmd)
	return cmd
}

// CmdTxGovManageFees creates the gov-manage-fees sub-command for the exchange tx command.
func CmdTxGovManageFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gov-manage-fees",
		Aliases: []string{"manage-fees", "gov-update-fees", "update-fees"},
		Short:   "Submit a governance proposal to change a market's fees",
		Args:    cobra.NoArgs,
		RunE:    govTxRunE(MakeMsgGovManageFees),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgGovManageFees(cmd)
	return cmd
}

// CmdTxGovUpdateParams creates the gov-update-params sub-command for the exchange tx command.
func CmdTxGovUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gov-update-params",
		Aliases: []string{"gov-params", "update-params", "params"},
		Short:   "Submit a governance proposal to update the exchange module params",
		Args:    cobra.NoArgs,
		RunE:    govTxRunE(MakeMsgGovUpdateParams),
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgGovUpdateParams(cmd)
	return cmd
}
