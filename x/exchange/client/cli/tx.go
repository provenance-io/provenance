package cli

import (
	"strings"

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
		RunE:    genericTxRunE(MakeMsgCreateAsk),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxCreateAsk(cmd)
	return cmd
}

// CmdTxCreateBid creates the create-bid sub-command for the exchange tx command.
func CmdTxCreateBid() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-bid",
		Aliases: []string{"bid", "create-bid-order", "bid-order"},
		Short:   "Create a bid order",
		RunE:    genericTxRunE(MakeMsgCreateBid),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxCreateBid(cmd)
	return cmd
}

// CmdTxCancelOrder creates the cancel-order sub-command for the exchange tx command.
func CmdTxCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel-order",
		Aliases: []string{"cancel"},
		Short:   "Cancel an order",
		RunE:    genericTxRunE(MakeMsgCancelOrder),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxCancelOrder(cmd)
	return cmd
}

// CmdTxFillBids creates the fill-bids sub-command for the exchange tx command.
func CmdTxFillBids() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill-bids",
		Short: "Fill one or more bid orders",
		RunE:  genericTxRunE(MakeMsgFillBids),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxFillBids(cmd)
	return cmd
}

// CmdTxFillAsks creates the fill-asks sub-command for the exchange tx command.
func CmdTxFillAsks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill-asks",
		Short: "Fill one or more ask orders",
		RunE:  genericTxRunE(MakeMsgFillAsks),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxFillAsks(cmd)
	return cmd
}

// CmdTxMarketSettle creates the market-settle sub-command for the exchange tx command.
func CmdTxMarketSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-settle",
		Aliases: []string{"settle"},
		Short:   "Settle some orders",
		RunE:    genericTxRunE(MakeMsgMarketSettle),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxMarketSettle(cmd)
	return cmd
}

// CmdTxMarketSetOrderExternalID creates the market-set-external-id sub-command for the exchange tx command.
func CmdTxMarketSetOrderExternalID() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-set-external-id",
		Aliases: []string{"market-set-order-external-id", "set-external-id", "external-id"},
		Short:   "Set an order's external id",
		RunE:    genericTxRunE(MakeMsgMarketSetOrderExternalID),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxMarketSetOrderExternalID(cmd)
	return cmd
}

// CmdTxMarketWithdraw creates the market-withdraw sub-command for the exchange tx command.
func CmdTxMarketWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-withdraw",
		Aliases: []string{"withdraw"},
		Short:   "Withdraw funds from a market account",
		RunE:    genericTxRunE(MakeMsgMarketWithdraw),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxMarketWithdraw(cmd)
	return cmd
}

// CmdTxMarketUpdateDetails creates the market-details sub-command for the exchange tx command.
func CmdTxMarketUpdateDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-details",
		Aliases: []string{"market-update-details", "update-market-details", "update-details"},
		Short:   "Update a market's details",
		RunE:    genericTxRunE(MakeMsgMarketUpdateDetails),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxMarketUpdateDetails(cmd)
	return cmd
}

// CmdTxMarketUpdateEnabled creates the market-accepting-orders sub-command for the exchange tx command.
func CmdTxMarketUpdateEnabled() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-accepting-orders",
		Aliases: []string{"market-update-accepting-orders", "update-market-accepting-orders", "update-accepting-orders"},
		Short:   "Change whether a market is accepting orders",
		RunE:    genericTxRunE(MakeMsgMarketUpdateAcceptingOrders),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxMarketUpdateAcceptingOrders(cmd)
	return cmd
}

// CmdTxMarketUpdateUserSettle creates the market-user-settle sub-command for the exchange tx command.
func CmdTxMarketUpdateUserSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-user-settle",
		Aliases: []string{"market-update-user-settle", "update-market-user-settle", "update-user-settle"},
		Short:   "Change whether a market allows settlements initiated by users",
		RunE:    genericTxRunE(MakeMsgMarketUpdateUserSettle),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxMarketUpdateUserSettle(cmd)
	return cmd
}

// CmdTxMarketManagePermissions creates the market-permissions sub-command for the exchange tx command.
func CmdTxMarketManagePermissions() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-permissions",
		Aliases: []string{"market-manage-permissions", "manage-market-permissions", "manage-permissions", "permissions"},
		Short:   "Update the account permissions for a market",
		RunE:    genericTxRunE(MakeMsgMarketManagePermissions),
	}

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxMarketManagePermissions(cmd)
	return cmd
}

// CmdTxMarketManageReqAttrs creates the market-req-attrs sub-command for the exchange tx command.
func CmdTxMarketManageReqAttrs() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-req-attrs",
		Aliases: []string{"market-manage-req-attrs", "manage-market-req-attrs", "manage-req-attrs", "req-attrs"},
		Short:   "Manage the attributes required to create orders in a market",
		RunE:    genericTxRunE(MakeMsgMarketManageReqAttrs),
	}
	newAliases := make([]string, 0, len(cmd.Aliases))
	for _, alias := range cmd.Aliases {
		if strings.Contains(alias, "req-attrs") {
			newAliases = append(newAliases, strings.Replace(alias, "req-attrs", "required-attributes", 1))
		}
	}
	cmd.Aliases = append(cmd.Aliases, newAliases...)

	flags.AddTxFlagsToCmd(cmd)
	SetupCmdTxMarketManageReqAttrs(cmd)
	return cmd
}

// CmdTxGovCreateMarket creates the gov-create-market sub-command for the exchange tx command.
func CmdTxGovCreateMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gov-create-market",
		Aliases: []string{"create-market"},
		Short:   "Submit a governance proposal to create a market",
		RunE:    govTxRunE(MakeMsgGovCreateMarket),
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	SetupCmdTxGovCreateMarket(cmd)
	return cmd
}

// CmdTxGovManageFees creates the gov-manage-fees sub-command for the exchange tx command.
func CmdTxGovManageFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gov-manage-fees",
		Aliases: []string{"manage-fees", "gov-update-fees", "update-fees"},
		Short:   "Submit a governance proposal to change a market's fees",
		RunE:    govTxRunE(MakeMsgGovManageFees),
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	SetupCmdTxGovManageFees(cmd)
	return cmd
}

// CmdTxGovUpdateParams creates the gov-update-params sub-command for the exchange tx command.
func CmdTxGovUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gov-update-params",
		Aliases: []string{"gov-params", "update-params", "params"},
		Short:   "Submit a governance proposal to update the exchange module params",
		RunE:    govTxRunE(MakeMsgGovUpdateParams),
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	SetupCmdTxGovUpdateParams(cmd)
	return cmd
}
